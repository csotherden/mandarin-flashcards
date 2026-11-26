package lesson

import (
	"errors"
	"net/http"

	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/own"
)

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	var req models.LessonInput
	if !handlerutils.DecodeJSON(w, r, h.logger, &req) {
		return
	}
	title, msg := cleanTitle(req.Title)
	if msg != "" {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, msg, nil)
		return
	}
	req.SetIDs = own.Unique(req.SetIDs)
	if err := own.CardSets(r.Context(), h.db, userID, req.SetIDs); err != nil {
		if errors.Is(err, own.ErrNotFound) {
			handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "A set does not belong to this account", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not create the lesson", err)
		return
	}
	row, err := h.db.Lesson.Create().
		SetTitle(title).
		SetOwnerID(userID).
		AddCardSetIDs(req.SetIDs...).
		Save(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not create the lesson", err)
		return
	}
	h.writeDetail(w, r, userID, row.ID, http.StatusCreated)
}
