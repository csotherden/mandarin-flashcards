package auth

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	appauth "github.com/csotherden/mandarin-flashcards/internal/auth"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/models"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if !handlerutils.DecodeJSON(w, r, h.logger, &req) {
		return
	}
	email := normalizeEmail(req.Email)
	row, err := h.db.User.Query().Where(user.EmailEQ(email)).Only(r.Context())
	if ent.IsNotFound(err) || (err == nil && !appauth.CheckPassword(req.Password, row.PasswordSalt, row.PasswordHash)) {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Invalid credentials", nil)
		return
	}
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not sign in", err)
		return
	}
	pair, err := appauth.GenerateTokenPair(row.ID)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not sign in", err)
		return
	}
	updated, err := h.db.User.UpdateOneID(row.ID).SetRefreshTokenID(pair.TokenID).Save(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not sign in", err)
		return
	}
	h.writeSession(w, http.StatusOK, updated, pair)
}
