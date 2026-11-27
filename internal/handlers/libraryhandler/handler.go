// Package libraryhandler serves the global card library: browse, look inside, import.
package libraryhandler

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/ent/cardset"
	"github.com/csotherden/mandarin-flashcards/ent/flashcard"
	"github.com/csotherden/mandarin-flashcards/ent/librarycard"
	"github.com/csotherden/mandarin-flashcards/ent/libraryset"
	"github.com/csotherden/mandarin-flashcards/ent/setflashcard"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/library"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/present"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

const previewWords = 4

type Handler struct {
	db     *ent.Client
	logger *zap.Logger
}

func NewHandler(db *ent.Client, logger *zap.Logger) *Handler {
	return &Handler{db: db, logger: logger}
}

func (h *Handler) Register(router *mux.Router) {
	router.HandleFunc("/library/sets", h.List).Methods(http.MethodGet)
	router.HandleFunc("/library/sets/{id}", h.Get).Methods(http.MethodGet)
	router.HandleFunc("/library/sets/{id}/import", h.Import).Methods(http.MethodPost)
}

// imported maps library set ids to the viewer's copies.
func (h *Handler) imported(r *http.Request, userID uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	sets, err := h.db.CardSet.Query().
		Where(cardset.HasOwnerWith(user.ID(userID)), cardset.LibrarySetIDNotNil()).
		Order(ent.Asc(cardset.FieldCreatedAt)).
		All(r.Context())
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]uuid.UUID{}
	for _, s := range sets {
		out[*s.LibrarySetID] = s.ID // newest copy wins
	}
	return out, nil
}

func summary(s *ent.LibrarySet, copies map[uuid.UUID]uuid.UUID) models.LibrarySetResponse {
	resp := models.LibrarySetResponse{
		ID: s.ID, Slug: s.Slug, Title: s.Title, Description: s.Description, Level: s.Level,
		CardCount: len(s.Edges.Cards), Preview: []string{},
	}
	for _, c := range s.Edges.Cards {
		if resp.CoverURL == "" && c.ImageURL != "" {
			resp.CoverURL = c.ImageURL
		}
		if len(resp.Preview) < previewWords {
			resp.Preview = append(resp.Preview, c.Hanzi)
		}
	}
	if id, ok := copies[s.ID]; ok {
		resp.ImportedSetID = &id
	}
	return resp
}

func cardsInOrder(q *ent.LibraryCardQuery) { q.Order(ent.Asc(librarycard.FieldSortOrder)) }

// List is GET /api/library/sets.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	sets, err := h.db.LibrarySet.Query().
		Where(libraryset.Published(true)).
		WithCards(cardsInOrder).
		Order(ent.Asc(libraryset.FieldSortOrder), ent.Asc(libraryset.FieldTitle)).
		All(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not load the library", err)
		return
	}
	copies, err := h.imported(r, userID)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not load the library", err)
		return
	}
	out := make([]models.LibrarySetResponse, 0, len(sets))
	for _, s := range sets {
		out = append(out, summary(s, copies))
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, out)
}

// Get is GET /api/library/sets/{id}.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	id, ok := handlerutils.PathID(w, r, h.logger, "id")
	if !ok {
		return
	}
	s, err := h.db.LibrarySet.Query().
		Where(libraryset.ID(id), libraryset.Published(true)).
		WithCards(cardsInOrder).
		Only(r.Context())
	if ent.IsNotFound(err) {
		handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Set not found", nil)
		return
	}
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not load the set", err)
		return
	}
	copies, err := h.imported(r, userID)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not load the set", err)
		return
	}
	words := make([]string, 0, len(s.Edges.Cards))
	for _, c := range s.Edges.Cards {
		words = append(words, c.Hanzi)
	}
	owned, err := h.db.Flashcard.Query().
		Where(flashcard.HasOwnerWith(user.ID(userID)), flashcard.HanziIn(words...)).
		Select(flashcard.FieldHanzi).
		Strings(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not load the set", err)
		return
	}
	have := map[string]bool{}
	for _, w := range owned {
		have[w] = true
	}
	detail := models.LibrarySetDetail{LibrarySetResponse: summary(s, copies), Cards: make([]models.LibraryCardResponse, 0, len(s.Edges.Cards))}
	for _, c := range s.Edges.Cards {
		detail.Cards = append(detail.Cards, models.LibraryCardResponse{
			ID: c.ID, Hanzi: c.Hanzi, Pinyin: c.Pinyin, TonalPinyin: c.TonalPinyin, English: c.English,
			ExampleSentenceHanzi: c.ExampleSentenceHanzi, ExampleSentencePinyin: c.ExampleSentencePinyin,
			ImageURL: c.ImageURL, UnsplashAttribution: c.UnsplashAttribution, Owned: have[c.Hanzi],
		})
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, detail)
}

// Import is POST /api/library/sets/{id}/import. Returns the viewer's copy: 201 when made now,
// 200 when they already had one.
func (h *Handler) Import(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	id, ok := handlerutils.PathID(w, r, h.logger, "id")
	if !ok {
		return
	}
	res, err := library.Import(r.Context(), h.db, userID, id)
	if ent.IsNotFound(err) {
		handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Set not found", nil)
		return
	}
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not add the set", err)
		return
	}
	count, err := h.db.SetFlashcard.Query().Where(setflashcard.SetID(res.Set.ID)).Count(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not add the set", err)
		return
	}
	status := http.StatusOK
	if res.Created {
		status = http.StatusCreated
	}
	handlerutils.RespondWithJSON(w, h.logger, status, models.LibraryImportResponse{Set: present.Set(res.Set, count), Created: res.Created, NewCards: res.NewCards})
}
