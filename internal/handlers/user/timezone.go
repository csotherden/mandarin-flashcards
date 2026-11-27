package user

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/present"
	"github.com/csotherden/mandarin-flashcards/internal/streak"
)

// UpdateTimezone is PUT /api/user/timezone {timezone}. The app sends the device's zone when it
// differs from the saved one, so the user's own days (not a viewer's) decide their streak.
func (h *Handler) UpdateTimezone(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	var req models.TimezoneRequest
	if !handlerutils.DecodeJSON(w, r, h.logger, &req) {
		return
	}
	if !streak.ValidTimezone(req.Timezone) {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Unknown timezone", nil)
		return
	}
	row, err := h.db.User.UpdateOneID(userID).SetTimezone(req.Timezone).Save(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not save the timezone", err)
		return
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, present.User(row))
}
