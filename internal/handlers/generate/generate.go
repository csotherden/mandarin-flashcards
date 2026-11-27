// Package generate builds a whole themed set on autopilot: the model picks words the learner does
// not have yet, each is auto-filled and given its first photo, and the cards are saved into a new
// set (or an existing one).
package generate

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/ent/cardset"
	"github.com/csotherden/mandarin-flashcards/ent/category"
	entflashcard "github.com/csotherden/mandarin-flashcards/ent/flashcard"
	"github.com/csotherden/mandarin-flashcards/ent/setflashcard"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/handlers/flashcard"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/llm"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/pinyin"
	"github.com/csotherden/mandarin-flashcards/internal/present"
	"github.com/csotherden/mandarin-flashcards/internal/ratelimit"
	"github.com/csotherden/mandarin-flashcards/internal/unsplash"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// MaxCards is the most cards one request generates.
const MaxCards = 15

// parallel is how many words are auto-filled and photographed at once.
const parallel = 5

type Handler struct {
	db     *ent.Client
	llm    *llm.Client
	photos *unsplash.Client
	cards  *flashcard.Handler
	limit  *ratelimit.Limiter
	logger *zap.Logger
}

func NewHandler(db *ent.Client, client *llm.Client, photos *unsplash.Client, cards *flashcard.Handler, limit *ratelimit.Limiter, logger *zap.Logger) *Handler {
	return &Handler{db: db, llm: client, photos: photos, cards: cards, limit: limit, logger: logger}
}

func (h *Handler) Register(router *mux.Router) {
	router.HandleFunc("/sets/generate", h.Generate).Methods(http.MethodPost)
}

// Generate is POST /api/sets/generate {theme, count, set_id?}.
func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	var req models.GenerateSetRequest
	if !handlerutils.DecodeJSON(w, r, h.logger, &req) {
		return
	}
	req.Theme = strings.TrimSpace(req.Theme)
	switch {
	case utf8.RuneCountInString(req.Theme) < 2:
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Describe the set in a few words", nil)
		return
	case utf8.RuneCountInString(req.Theme) > 80:
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Keep the theme to a few words", nil)
		return
	case req.Count < 1 || req.Count > MaxCards:
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Choose 1 to 15 cards", nil)
		return
	}
	if !h.llm.Enabled() {
		handlerutils.RespondWithError(w, h.logger, http.StatusServiceUnavailable, handlerutils.CodeNotSet, "Card suggestions are not configured", nil)
		return
	}
	if !h.limit.Allow(userID.String()) {
		handlerutils.RespondWithError(w, h.logger, http.StatusTooManyRequests, handlerutils.CodeNotSet, "Too many suggestions. Wait a minute and try again", nil)
		return
	}
	ctx := r.Context()
	fail := func(err error) {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not generate the set", err)
	}

	var set *ent.CardSet
	if req.SetID != nil {
		found, err := h.db.CardSet.Query().Where(cardset.ID(*req.SetID), cardset.HasOwnerWith(user.ID(userID))).Only(ctx)
		if ent.IsNotFound(err) {
			handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Set not found", nil)
			return
		}
		if err != nil {
			fail(err)
			return
		}
		set = found
	}

	known, err := h.db.Flashcard.Query().Where(entflashcard.HasOwnerWith(user.ID(userID))).
		Order(ent.Desc(entflashcard.FieldCreatedAt)).Select(entflashcard.FieldHanzi).Strings(ctx)
	if err != nil {
		fail(err)
		return
	}
	// Ask for a few spares: the model sometimes repeats a known word, which is dropped below.
	words, err := h.llm.SuggestWords(ctx, req.Theme, req.Count+4, firstN(known, 400))
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadGateway, handlerutils.CodeNotSet, "Could not choose words for that theme", err)
		return
	}
	chosen, skipped := pickNew(words, known, req.Count)
	if len(chosen) == 0 {
		handlerutils.RespondWithError(w, h.logger, http.StatusConflict, handlerutils.CodeNotSet, "You already have every word suggested for that theme. Try a different one", nil)
		return
	}

	categories, err := h.db.Category.Query().Where(category.HasOwnerWith(user.ID(userID))).Select(category.FieldName).Strings(ctx)
	if err != nil {
		fail(err)
		return
	}
	drafts := h.draft(ctx, chosen, categories)

	if set == nil {
		set, err = h.db.CardSet.Create().SetTitle(titleCase(req.Theme)).SetOwnerID(userID).Save(ctx)
		if err != nil {
			fail(err)
			return
		}
	}
	out := models.GenerateSetResponse{Cards: []models.FlashcardResponse{}, Skipped: skipped}
	for _, d := range drafts {
		if d.err != nil {
			h.logger.Warn("generate: could not fill a word", zap.String("word", d.word), zap.Error(d.err))
			out.Skipped = append(out.Skipped, d.word)
			continue
		}
		d.input.SetIDs = []uuid.UUID{set.ID}
		in, err := flashcard.NormalizeCard(d.input)
		if err != nil {
			h.logger.Warn("generate: dropping an invalid card", zap.String("word", d.word), zap.Error(err))
			out.Skipped = append(out.Skipped, d.word)
			continue
		}
		card, err := h.cards.CreateCard(ctx, userID, in)
		if err != nil {
			fail(err)
			return
		}
		h.cards.TrackDownload(ctx, in.UnsplashAttribution)
		out.Cards = append(out.Cards, present.Flashcard(card))
	}
	count, err := h.db.SetFlashcard.Query().Where(setflashcard.SetID(set.ID)).Count(ctx)
	if err != nil {
		fail(err)
		return
	}
	out.Set = present.Set(set, count)
	handlerutils.RespondWithJSON(w, h.logger, http.StatusCreated, out)
}

type draft struct {
	word  string
	input models.FlashcardInput
	err   error
}

// draft auto-fills each word and picks its first photo, a few words at a time, keeping order.
func (h *Handler) draft(ctx context.Context, words, categories []string) []draft {
	out := make([]draft, len(words))
	sem := make(chan struct{}, parallel)
	var wg sync.WaitGroup
	for i, word := range words {
		wg.Add(1)
		go func(i int, word string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			out[i] = h.fill(ctx, word, categories)
		}(i, word)
	}
	wg.Wait()
	return out
}

func (h *Handler) fill(ctx context.Context, word string, categories []string) draft {
	filled, err := h.llm.Autofill(ctx, word, categories)
	if err != nil {
		return draft{word: word, err: err}
	}
	in := models.FlashcardInput{
		Hanzi:                 word,
		Pinyin:                filled.Pinyin,
		English:               filled.English,
		ExampleSentenceHanzi:  filled.ExampleSentenceHanzi,
		ExampleSentencePinyin: filled.ExampleSentencePinyin,
		CategoryNames:         filled.Categories,
	}
	if len(in.CategoryNames) > 3 {
		in.CategoryNames = in.CategoryNames[:3]
	}
	if tonal := pinyin.NormalizeTonal(filled.TonalPinyin); pinyin.ValidTonal(word, tonal) {
		in.TonalPinyin = tonal
	}
	if h.photos.Enabled() && filled.ImageSearch != "" {
		photos, err := h.photos.Search(ctx, filled.ImageSearch)
		if err != nil {
			h.logger.Warn("generate: photo search failed", zap.String("word", word), zap.Error(err))
		} else if len(photos) > 0 {
			p := photos[0]
			in.ImageURL = p.FullURL
			in.UnsplashAttribution = &models.UnsplashAttribution{
				PhotographerName: p.PhotographerName,
				PhotographerURL:  p.PhotographerURL,
				PhotoURL:         p.PhotoURL,
				DownloadLocation: p.DownloadLocation,
			}
		}
	}
	return draft{word: word, input: in}
}

// pickNew keeps up to n words that are hanzi, not already known and not repeated, in order;
// the rest are returned as skipped.
func pickNew(words, known []string, n int) (chosen, skipped []string) {
	have := make(map[string]bool, len(known))
	for _, k := range known {
		have[k] = true
	}
	chosen, skipped = []string{}, []string{}
	for _, w := range words {
		w = strings.TrimSpace(w)
		switch {
		case w == "" || !allHan(w):
			continue
		case have[w]:
			skipped = append(skipped, w)
		case len(chosen) < n:
			have[w] = true
			chosen = append(chosen, w)
		}
	}
	return chosen, skipped
}

func allHan(s string) bool {
	for _, r := range s {
		if !unicode.Is(unicode.Han, r) {
			return false
		}
	}
	return s != ""
}

func firstN(s []string, n int) []string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// titleCase capitalizes each word of a theme for the set's title ("ordering coffee" → "Ordering Coffee").
func titleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		words[i] = string(r)
	}
	out := strings.Join(words, " ")
	if utf8.RuneCountInString(out) > 120 {
		out = string([]rune(out)[:120])
	}
	return out
}
