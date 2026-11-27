package user

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent/flashcard"
	entuser "github.com/csotherden/mandarin-flashcards/ent/user"
	appauth "github.com/csotherden/mandarin-flashcards/internal/auth"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/present"
	"go.uber.org/zap"
)

// ChangePasswordRequest changes the signed-in user's password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// DeleteAccountRequest confirms deleting the account with the password.
type DeleteAccountRequest struct {
	Password string `json:"password"`
}

// ChangePassword is PUT /api/user/password. It issues a fresh session for this device; the new
// refresh token replaces the old one, so other devices are signed out when their access expires.
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	var req ChangePasswordRequest
	if !handlerutils.DecodeJSON(w, r, h.logger, &req) {
		return
	}
	ctx := r.Context()
	row, err := h.db.User.Get(ctx, userID)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not change the password", err)
		return
	}
	if !appauth.CheckPassword(req.CurrentPassword, row.PasswordSalt, row.PasswordHash) {
		handlerutils.RespondWithError(w, h.logger, http.StatusForbidden, handlerutils.CodeNotSet, "Your current password is incorrect", nil)
		return
	}
	if len(req.NewPassword) < 8 || len(req.NewPassword) > 128 {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Use a new password of at least 8 characters", nil)
		return
	}
	salt, err := appauth.NewSalt()
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not change the password", err)
		return
	}
	hash, err := appauth.HashPassword(req.NewPassword, salt)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not change the password", err)
		return
	}
	pair, err := appauth.GenerateTokenPair(userID)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not change the password", err)
		return
	}
	updated, err := h.db.User.UpdateOneID(userID).
		SetPasswordSalt(salt).SetPasswordHash(hash).SetRefreshTokenID(pair.TokenID).
		Save(ctx)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not change the password", err)
		return
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, models.LoginResponse{
		AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken, User: present.User(updated),
	})
}

// DeleteAccount is DELETE /api/user {password}. Everything the user owns goes with the account
// (the database cascades), and their uploaded photos are removed from storage.
func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	var req DeleteAccountRequest
	if !handlerutils.DecodeJSON(w, r, h.logger, &req) {
		return
	}
	ctx := r.Context()
	row, err := h.db.User.Get(ctx, userID)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not delete the account", err)
		return
	}
	if !appauth.CheckPassword(req.Password, row.PasswordSalt, row.PasswordHash) {
		handlerutils.RespondWithError(w, h.logger, http.StatusForbidden, handlerutils.CodeNotSet, "That password is incorrect", nil)
		return
	}
	keys, err := h.db.Flashcard.Query().
		Where(flashcard.HasOwnerWith(entuser.ID(userID)), flashcard.ImageObjectKeyNotNil()).
		Select(flashcard.FieldImageObjectKey).
		Strings(ctx)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not delete the account", err)
		return
	}
	if err := h.db.User.DeleteOneID(userID).Exec(ctx); err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not delete the account", err)
		return
	}
	for _, key := range keys {
		if err := h.files.DeleteFile(ctx, key); err != nil {
			h.logger.Warn("failed to delete a deleted account's photo", zap.Error(err))
		}
	}
	h.logger.Info("account deleted", zap.String("user_id", userID.String()))
	handlerutils.RespondWithNoContent(w)
}
