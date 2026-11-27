package cardset

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/internal/models"
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
	router.HandleFunc("/sets", h.List).Methods(http.MethodGet)
	router.HandleFunc("/sets", h.Create).Methods(http.MethodPost)
	router.HandleFunc("/sets/{id}", h.Get).Methods(http.MethodGet)
	router.HandleFunc("/sets/{id}", h.Update).Methods(http.MethodPut)
	router.HandleFunc("/sets/{id}", h.Delete).Methods(http.MethodDelete)
	router.HandleFunc("/sets/{id}/cards", h.ReplaceCards).Methods(http.MethodPut)
	router.HandleFunc("/sets/{id}/cards/add", h.AddCards).Methods(http.MethodPost)
}

func cleanSet(in models.SetInput) (models.SetInput, string) {
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	if in.Title == "" || utf8.RuneCountInString(in.Title) > 120 {
		return in, "Set title is required"
	}
	if utf8.RuneCountInString(in.Description) > 500 {
		return in, "Description is too long"
	}
	return in, ""
}
