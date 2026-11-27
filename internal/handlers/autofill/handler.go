package autofill

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/internal/llm"
	"github.com/csotherden/mandarin-flashcards/internal/ratelimit"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Handler struct {
	db     *ent.Client
	llm    *llm.Client
	limit  *ratelimit.Limiter
	logger *zap.Logger
}

func NewHandler(db *ent.Client, client *llm.Client, limit *ratelimit.Limiter, logger *zap.Logger) *Handler {
	return &Handler{db: db, llm: client, limit: limit, logger: logger}
}

func (h *Handler) Register(router *mux.Router) {
	router.HandleFunc("/flashcards/autofill", h.Autofill).Methods(http.MethodPost)
}
