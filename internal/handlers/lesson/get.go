package lesson

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent"
	entlesson "github.com/csotherden/mandarin-flashcards/ent/lesson"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/study"
	"github.com/google/uuid"
)

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	id, ok := handlerutils.PathID(w, r, h.logger, "id")
	if !ok {
		return
	}
	h.writeDetail(w, r, userID, id, http.StatusOK)
}

func (h *Handler) writeDetail(w http.ResponseWriter, r *http.Request, userID, id uuid.UUID, status int) {
	row, err := h.db.Lesson.Query().
		Where(entlesson.ID(id), entlesson.HasOwnerWith(user.ID(userID))).
		WithCardSets(func(q *ent.CardSetQuery) {
			q.WithMemberships()
		}).
		Only(r.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Lesson not found", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not load the lesson", err)
		return
	}
	deck, err := study.LoadLesson(r.Context(), h.db, userID, id)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not load the lesson", err)
		return
	}
	handlerutils.RespondWithJSON(w, h.logger, status, detailFrom(row, deck))
}
