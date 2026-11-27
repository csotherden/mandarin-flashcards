package user

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent"
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
	router.HandleFunc("/user/profile", h.GetProfile).Methods(http.MethodGet)
	router.HandleFunc("/user/preferences", h.GetPreferences).Methods(http.MethodGet)
	router.HandleFunc("/user/preferences", h.UpdatePreferences).Methods(http.MethodPut)
	router.HandleFunc("/user/timezone", h.UpdateTimezone).Methods(http.MethodPut)
}
