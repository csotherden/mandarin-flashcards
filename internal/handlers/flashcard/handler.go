package flashcard

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/csotherden/mandarin-flashcards/ent"
	entflashcard "github.com/csotherden/mandarin-flashcards/ent/flashcard"
	"github.com/csotherden/mandarin-flashcards/ent/setflashcard"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/own"
	"github.com/csotherden/mandarin-flashcards/internal/pinyin"
	"github.com/csotherden/mandarin-flashcards/internal/storage"
	"github.com/csotherden/mandarin-flashcards/internal/unsplash"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

const uuidPattern = `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`

type Handler struct {
	db       *ent.Client
	files    storage.FileStore
	unsplash *unsplash.Client
	logger   *zap.Logger
}

func NewHandler(db *ent.Client, files storage.FileStore, photos *unsplash.Client, logger *zap.Logger) *Handler {
	return &Handler{db: db, files: files, unsplash: photos, logger: logger}
}

func (h *Handler) Register(router *mux.Router) {
	router.HandleFunc("/flashcards", h.List).Methods(http.MethodGet)
	router.HandleFunc("/flashcards", h.Create).Methods(http.MethodPost)
	router.HandleFunc("/flashcards/{id:"+uuidPattern+"}", h.Get).Methods(http.MethodGet)
	router.HandleFunc("/flashcards/{id:"+uuidPattern+"}", h.Update).Methods(http.MethodPut)
	router.HandleFunc("/flashcards/{id:"+uuidPattern+"}", h.Delete).Methods(http.MethodDelete)
	router.HandleFunc("/flashcards/{id:"+uuidPattern+"}/image", h.UploadImage).Methods(http.MethodPost)
	router.HandleFunc("/flashcards/{id:"+uuidPattern+"}/image", h.DownloadImage).Methods(http.MethodGet)
}

func cardQuery(client *ent.Client) *ent.FlashcardQuery {
	return client.Flashcard.Query().WithCategories().WithSetMemberships()
}

func (h *Handler) owned(ctx context.Context, client *ent.Client, userID, id uuid.UUID) (*ent.Flashcard, error) {
	row, err := cardQuery(client).
		Where(entflashcard.ID(id), entflashcard.HasOwnerWith(user.ID(userID))).
		Only(ctx)
	return row, err
}

// NormalizeCard trims and validates a card input.
func NormalizeCard(in models.FlashcardInput) (models.FlashcardInput, error) {
	in.Hanzi = strings.TrimSpace(in.Hanzi)
	in.Pinyin = strings.TrimSpace(in.Pinyin)
	in.TonalPinyin = pinyin.NormalizeTonal(in.TonalPinyin)
	in.English = strings.TrimSpace(in.English)
	in.ExampleSentenceHanzi = strings.TrimSpace(in.ExampleSentenceHanzi)
	in.ExampleSentencePinyin = strings.TrimSpace(in.ExampleSentencePinyin)
	in.ImageURL = strings.TrimSpace(in.ImageURL)
	in.CategoryIDs = own.Unique(in.CategoryIDs)
	in.SetIDs = own.Unique(in.SetIDs)
	switch {
	case in.Hanzi == "":
		return in, errors.New("Hanzi is required")
	case utf8.RuneCountInString(in.Hanzi) > 64:
		return in, errors.New("Hanzi is too long")
	case utf8.RuneCountInString(in.Pinyin) > 160:
		return in, errors.New("Pinyin is too long")
	case len(in.CategoryNames) > 10:
		return in, errors.New("Choose at most 10 categories")
	case in.TonalPinyin != "" && !pinyin.ValidTonal(in.Hanzi, in.TonalPinyin):
		return in, errors.New("Tonal pinyin needs one numbered syllable per character, like ma1 ma5 (5 for the neutral tone)")
	case utf8.RuneCountInString(in.English) > 300:
		return in, errors.New("English is too long")
	case utf8.RuneCountInString(in.ExampleSentenceHanzi) > 500 || utf8.RuneCountInString(in.ExampleSentencePinyin) > 500:
		return in, errors.New("Example sentence is too long")
	}
	if in.ImageURL != "" {
		parsed, err := url.Parse(in.ImageURL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return in, errors.New("Image URL must start with https")
		}
		in.ImageURL = parsed.String()
	}
	return in, nil
}

func (h *Handler) linkCard(ctx context.Context, client *ent.Client, userID uuid.UUID, card *ent.Flashcard, in models.FlashcardInput) (*ent.Flashcard, error) {
	if err := own.Categories(ctx, client, userID, in.CategoryIDs); err != nil {
		return nil, err
	}
	named, err := own.CategoryNames(ctx, client, userID, in.CategoryNames)
	if err != nil {
		return nil, err
	}
	in.CategoryIDs = own.Unique(append(in.CategoryIDs, named...))
	if err := own.CardSets(ctx, client, userID, in.SetIDs); err != nil {
		return nil, err
	}
	if _, err := client.SetFlashcard.Delete().Where(setflashcard.FlashcardID(card.ID)).Exec(ctx); err != nil {
		return nil, err
	}
	for _, setID := range in.SetIDs {
		count, err := client.SetFlashcard.Query().Where(setflashcard.SetID(setID)).Count(ctx)
		if err != nil {
			return nil, err
		}
		if _, err := client.SetFlashcard.Create().
			SetSetID(setID).
			SetFlashcardID(card.ID).
			SetSortOrder(count).
			Save(ctx); err != nil {
			return nil, err
		}
	}
	updated, err := client.Flashcard.UpdateOneID(card.ID).
		ClearCategories().
		AddCategoryIDs(in.CategoryIDs...).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return h.owned(ctx, client, userID, updated.ID)
}

func (h *Handler) trackDownload(ctx context.Context, attribution *models.UnsplashAttribution) {
	if attribution == nil || h.unsplash == nil || attribution.DownloadLocation == "" {
		return
	}
	if err := h.unsplash.TrackDownload(ctx, attribution.DownloadLocation); err != nil {
		h.logger.Warn("unsplash download tracking failed", zap.Error(err))
	}
}
