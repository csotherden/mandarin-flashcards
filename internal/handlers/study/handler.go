package study

import (
	"context"
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/ent/cardattemptlog"
	entattempt "github.com/csotherden/mandarin-flashcards/ent/lessonattempt"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/llm"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/present"
	"github.com/csotherden/mandarin-flashcards/internal/ratelimit"
	deck "github.com/csotherden/mandarin-flashcards/internal/study"
	"github.com/csotherden/mandarin-flashcards/internal/toneperfect"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Handler struct {
	db     *ent.Client
	tones  *toneperfect.Client
	llm    *llm.Client
	limit  *ratelimit.Limiter
	logger *zap.Logger
}

func NewHandler(db *ent.Client, tones *toneperfect.Client, coach *llm.Client, limit *ratelimit.Limiter, logger *zap.Logger) *Handler {
	return &Handler{db: db, tones: tones, llm: coach, limit: limit, logger: logger}
}

func (h *Handler) Register(router *mux.Router) {
	router.HandleFunc("/study/sessions", h.Start).Methods(http.MethodPost)
	router.HandleFunc("/study/sessions/{id}", h.Get).Methods(http.MethodGet)
	router.HandleFunc("/study/sessions/{id}/complete", h.Complete).Methods(http.MethodPost)
	router.HandleFunc("/study/sessions/{id}/matches", h.Match).Methods(http.MethodPost)
	router.HandleFunc("/assess", h.Assess).Methods(http.MethodPost)
}

func parseMode(raw string) (entattempt.Mode, bool) {
	switch raw {
	case models.ModeImageSpeak:
		return entattempt.ModeImageSpeak, true
	case models.ModeHanziSpeak:
		return entattempt.ModeHanziSpeak, true
	case models.ModePictureHanziMatch:
		return entattempt.ModePictureHanziMatch, true
	default:
		return "", false
	}
}

func (h *Handler) ownedAttempt(ctx context.Context, userID, id uuid.UUID) (*ent.LessonAttempt, error) {
	return h.db.LessonAttempt.Query().
		Where(entattempt.ID(id), entattempt.HasOwnerWith(user.ID(userID))).
		Only(ctx)
}

func (h *Handler) session(ctx context.Context, userID uuid.UUID, attempt *ent.LessonAttempt) (*models.SessionResponse, error) {
	cards, err := deck.LoadOwned(ctx, h.db, userID, attempt.CardIds)
	if err != nil {
		return nil, err
	}
	logs, err := h.db.CardAttemptLog.Query().
		Where(cardattemptlog.LessonAttemptID(attempt.ID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	latest := map[uuid.UUID]*ent.CardAttemptLog{}
	earned := 0
	for _, log := range logs {
		earned += log.XpEarned
		prev := latest[log.FlashcardID]
		if prev == nil || log.CreatedAt.After(prev.CreatedAt) {
			latest[log.FlashcardID] = log
		}
	}
	results := make([]models.CardResult, 0, len(latest))
	for _, id := range attempt.CardIds {
		if log, ok := latest[id]; ok {
			results = append(results, present.CardResult(log))
		}
	}
	cardViews := make([]models.FlashcardResponse, 0, len(cards))
	for _, card := range cards {
		cardViews = append(cardViews, present.Flashcard(card))
	}
	return &models.SessionResponse{
		Attempt:      present.Attempt(attempt),
		Cards:        cardViews,
		Results:      results,
		OverallScore: attempt.OverallScore,
		XPEarned:     earned,
	}, nil
}

func containsCard(ids []uuid.UUID, id uuid.UUID) bool {
	for _, candidate := range ids {
		if candidate == id {
			return true
		}
	}
	return false
}
