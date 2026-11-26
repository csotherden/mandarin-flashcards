package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/auth"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// AuthMiddleware validates a local bearer JWT and confirms the user still exists.
func AuthMiddleware(client *ent.Client, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				handlerutils.RespondWithError(w, logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Authorization header missing", nil)
				return
			}
			splitToken := strings.Split(authHeader, "Bearer ")
			if len(splitToken) != 2 || strings.TrimSpace(splitToken[1]) == "" {
				handlerutils.RespondWithError(w, logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Invalid authorization format. Expected 'Bearer {token}'", nil)
				return
			}

			claims, err := auth.ValidateAccessToken(strings.TrimSpace(splitToken[1]))
			if err != nil {
				logger.Info("Invalid token", zap.Error(err))
				handlerutils.RespondWithError(w, logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Your session has expired. Please sign in again", nil)
				return
			}

			exists, err := client.User.Query().Where(user.ID(claims.UserID)).Exist(r.Context())
			if err != nil || !exists {
				logger.Info("Failed to resolve user from token", zap.Error(err))
				handlerutils.RespondWithError(w, logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Your session is no longer valid. Please sign in again", nil)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext returns the authenticated user id.
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}
