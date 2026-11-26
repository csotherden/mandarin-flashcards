package auth

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent/user"
	appauth "github.com/csotherden/mandarin-flashcards/internal/auth"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/models"
)

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if !handlerutils.DecodeJSON(w, r, h.logger, &req) {
		return
	}
	claims, err := appauth.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Your session has expired. Please sign in again", nil)
		return
	}
	row, err := h.db.User.Query().Where(user.ID(claims.UserID)).Only(r.Context())
	if err != nil || row.RefreshTokenID == nil || *row.RefreshTokenID != claims.TokenID {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Your session has expired. Please sign in again", nil)
		return
	}
	pair, err := appauth.GenerateTokenPair(row.ID)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not refresh the session", err)
		return
	}
	if _, err := h.db.User.UpdateOneID(row.ID).SetRefreshTokenID(pair.TokenID).Save(r.Context()); err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not refresh the session", err)
		return
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, models.TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}
