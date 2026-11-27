package auth

import (
	"github.com/csotherden/mandarin-flashcards/internal/streak"
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent"
	appauth "github.com/csotherden/mandarin-flashcards/internal/auth"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/models"
)

func (h *Handler) RegisterAccount(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if !handlerutils.DecodeJSON(w, r, h.logger, &req) {
		return
	}
	email := normalizeEmail(req.Email)
	if !validEmail(email) || !validPassword(req.Password) {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Enter a valid email and a password of at least 8 characters", nil)
		return
	}

	salt, err := appauth.NewSalt()
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not create the account", err)
		return
	}
	hash, err := appauth.HashPassword(req.Password, salt)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not create the account", err)
		return
	}

	create := h.db.User.Create().
		SetEmail(email).
		SetPasswordHash(hash).
		SetPasswordSalt(salt)
	if streak.ValidTimezone(req.Timezone) {
		create.SetTimezone(req.Timezone)
	}
	userRow, err := create.Save(r.Context())
	if err != nil {
		if ent.IsConstraintError(err) {
			handlerutils.RespondWithError(w, h.logger, http.StatusConflict, handlerutils.CodeNotSet, "Email already registered", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not create the account", err)
		return
	}

	pair, err := appauth.GenerateTokenPair(userRow.ID)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not create the account", err)
		return
	}
	updated, err := h.db.User.UpdateOneID(userRow.ID).SetRefreshTokenID(pair.TokenID).Save(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not create the account", err)
		return
	}
	h.writeSession(w, http.StatusCreated, updated, pair)
}
