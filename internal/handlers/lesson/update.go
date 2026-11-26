package lesson

import (
	"net/http"

	entlesson "github.com/csotherden/mandarin-flashcards/ent/lesson"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
)

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	id, ok := handlerutils.PathID(w, r, h.logger, "id")
	if !ok {
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
	n, err := h.db.Lesson.Update().
		Where(entlesson.ID(id), entlesson.HasOwnerWith(user.ID(userID))).
		SetTitle(title).
		Save(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the lesson", err)
		return
	}
	if n == 0 {
		handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Lesson not found", nil)
		return
	}
	h.writeDetail(w, r, userID, id, http.StatusOK)
}
