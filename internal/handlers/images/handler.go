package images

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/internal/ratelimit"
	"github.com/csotherden/mandarin-flashcards/internal/unsplash"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Handler struct {
	unsplash *unsplash.Client
	limit    *ratelimit.Limiter
	logger   *zap.Logger
}

func NewHandler(client *unsplash.Client, limit *ratelimit.Limiter, logger *zap.Logger) *Handler {
	return &Handler{unsplash: client, limit: limit, logger: logger}
}

func (h *Handler) Register(router *mux.Router) {
	router.HandleFunc("/images/search", h.Search).Methods(http.MethodGet)
}
