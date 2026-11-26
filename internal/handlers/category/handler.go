package category

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var colorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type Handler struct {
	db     *ent.Client
	logger *zap.Logger
}

func NewHandler(db *ent.Client, logger *zap.Logger) *Handler {
	return &Handler{db: db, logger: logger}
}

func (h *Handler) Register(router *mux.Router) {
	router.HandleFunc("/categories", h.List).Methods(http.MethodGet)
	router.HandleFunc("/categories", h.Create).Methods(http.MethodPost)
	router.HandleFunc("/categories/{id}", h.Update).Methods(http.MethodPut)
	router.HandleFunc("/categories/{id}", h.Delete).Methods(http.MethodDelete)
}

func cleanCategory(in models.CategoryInput) (models.CategoryInput, string) {
	in.Name = strings.TrimSpace(in.Name)
	in.Color = strings.TrimSpace(in.Color)
	if in.Color == "" {
		in.Color = "#9f2d20"
	}
	if in.Name == "" || len(in.Name) > 80 {
		return in, "Category name is required"
	}
	if !colorPattern.MatchString(in.Color) {
		return in, "Color must be a hex value like #9f2d20"
	}
	return in, ""
}
