package autofill

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/internal/llm"
	"github.com/csotherden/mandarin-flashcards/internal/ratelimit"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Handler struct {
	llm    *llm.Client
	limit  *ratelimit.Limiter
	logger *zap.Logger
}

func NewHandler(client *llm.Client, limit *ratelimit.Limiter, logger *zap.Logger) *Handler {
	return &Handler{llm: client, limit: limit, logger: logger}
}

func (h *Handler) Register(router *mux.Router) {
	router.HandleFunc("/flashcards/autofill", h.Autofill).Methods(http.MethodPost)
}
