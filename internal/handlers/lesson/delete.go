package lesson

import (
	"net/http"

	entlesson "github.com/csotherden/mandarin-flashcards/ent/lesson"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
)

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	id, ok := handlerutils.PathID(w, r, h.logger, "id")
	if !ok {
		return
	}
	n, err := h.db.Lesson.Delete().
		Where(entlesson.ID(id), entlesson.HasOwnerWith(user.ID(userID))).
		Exec(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not delete the lesson", err)
		return
	}
	if n == 0 {
		handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Lesson not found", nil)
		return
	}
	handlerutils.RespondWithNoContent(w)
}
