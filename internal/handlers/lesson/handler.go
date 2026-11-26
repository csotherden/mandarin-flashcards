package lesson

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/present"
	"github.com/csotherden/mandarin-flashcards/internal/study"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Handler struct {
	db     *ent.Client
	logger *zap.Logger
}

func NewHandler(db *ent.Client, logger *zap.Logger) *Handler {
	return &Handler{db: db, logger: logger}
}

func (h *Handler) Register(router *mux.Router) {
	router.HandleFunc("/lessons", h.List).Methods(http.MethodGet)
	router.HandleFunc("/lessons", h.Create).Methods(http.MethodPost)
	router.HandleFunc("/lessons/{id}", h.Get).Methods(http.MethodGet)
	router.HandleFunc("/lessons/{id}", h.Update).Methods(http.MethodPut)
	router.HandleFunc("/lessons/{id}", h.Delete).Methods(http.MethodDelete)
	router.HandleFunc("/lessons/{id}/sets", h.ReplaceSets).Methods(http.MethodPut)
	router.HandleFunc("/lessons/{id}/disabled", h.ReplaceDisabled).Methods(http.MethodPut)
}

func cleanTitle(title string) (string, string) {
	title = strings.TrimSpace(title)
	if title == "" || utf8.RuneCountInString(title) > 120 {
		return "", "Lesson title is required"
	}
	return title, ""
}

func detailFrom(row *ent.Lesson, deck *study.LessonDeck) models.LessonDetail {
	setIDs := make([]uuid.UUID, 0, len(row.Edges.CardSets))
	sets := make([]models.SetResponse, 0, len(row.Edges.CardSets))
	for _, set := range row.Edges.CardSets {
		setIDs = append(setIDs, set.ID)
		sets = append(sets, present.Set(set, len(set.Edges.Memberships)))
	}
	cards := make([]models.LessonCard, 0, len(deck.Cards))
	enabled := 0
	for _, card := range deck.Cards {
		_, off := deck.Disabled[card.ID]
		if !off {
			enabled++
		}
		cards = append(cards, models.LessonCard{Card: present.Flashcard(card), Enabled: !off})
	}
	return models.LessonDetail{
		LessonResponse: models.LessonResponse{
			ID:           row.ID,
			Title:        row.Title,
			SetIDs:       setIDs,
			CardCount:    len(deck.Cards),
			EnabledCount: enabled,
			CreatedAt:    row.CreatedAt,
		},
		Sets:  sets,
		Cards: cards,
	}
}
