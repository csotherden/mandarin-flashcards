package user

import (
	"net/http"
	"strings"

	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/present"
)

func (h *Handler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	row, err := h.db.User.Query().Where(user.ID(userID)).Only(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not load preferences", err)
		return
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, row.Preferences)
}

func (h *Handler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	var prefs models.Preferences
	if !handlerutils.DecodeJSON(w, r, h.logger, &prefs) {
		return
	}
	if prefs.DefaultTTSSpeed < 0.5 || prefs.DefaultTTSSpeed > 1.5 {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Speech speed must be between 0.5 and 1.5", nil)
		return
	}
	prefs.PreferredVoice = strings.TrimSpace(prefs.PreferredVoice)
	if prefs.PreferredVoice == "" {
		prefs.PreferredVoice = "zh-CN"
	}
	if len(prefs.PreferredVoice) > 64 {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Voice name is too long", nil)
		return
	}
	row, err := h.db.User.UpdateOneID(userID).SetPreferences(prefs).Save(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not save preferences", err)
		return
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, present.User(row).Preferences)
}
