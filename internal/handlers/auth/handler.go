package auth

import (
	"net/http"
	"strings"

	"github.com/csotherden/mandarin-flashcards/ent"
	appauth "github.com/csotherden/mandarin-flashcards/internal/auth"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/present"
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
	router.HandleFunc("/auth/register", h.RegisterAccount).Methods(http.MethodPost)
	router.HandleFunc("/auth/login", h.Login).Methods(http.MethodPost)
	router.HandleFunc("/auth/refresh", h.Refresh).Methods(http.MethodPost)
}

func (h *Handler) writeSession(w http.ResponseWriter, status int, user *ent.User, pair *appauth.TokenPair) {
	handlerutils.RespondWithJSON(w, h.logger, status, models.LoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		User:         present.User(user),
	})
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validEmail(email string) bool {
	if len(email) < 3 || len(email) > 255 {
		return false
	}
	at := strings.Index(email, "@")
	return at > 0 && at < len(email)-1 && strings.Contains(email[at:], ".")
}

func validPassword(password string) bool {
	return len(password) >= 8 && len(password) <= 128
}
