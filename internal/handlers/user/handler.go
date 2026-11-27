package user

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/internal/storage"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Handler struct {
	db     *ent.Client
	files  storage.FileStore
	logger *zap.Logger
}

func NewHandler(db *ent.Client, files storage.FileStore, logger *zap.Logger) *Handler {
	return &Handler{db: db, files: files, logger: logger}
}

func (h *Handler) Register(router *mux.Router) {
	router.HandleFunc("/user/profile", h.GetProfile).Methods(http.MethodGet)
	router.HandleFunc("/user/preferences", h.GetPreferences).Methods(http.MethodGet)
	router.HandleFunc("/user/preferences", h.UpdatePreferences).Methods(http.MethodPut)
	router.HandleFunc("/user/timezone", h.UpdateTimezone).Methods(http.MethodPut)
	router.HandleFunc("/user/password", h.ChangePassword).Methods(http.MethodPut)
	router.HandleFunc("/user", h.DeleteAccount).Methods(http.MethodDelete)
}
