package health

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Handler struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewHandler(db *sql.DB, logger *zap.Logger) *Handler {
	return &Handler{db: db, logger: logger}
}

func (h *Handler) Register(router *mux.Router) {
	router.HandleFunc("/health/live", h.Live).Methods(http.MethodGet)
	router.HandleFunc("/health/ready", h.Ready).Methods(http.MethodGet)
	router.HandleFunc("/health", h.Ready).Methods(http.MethodGet)
}
