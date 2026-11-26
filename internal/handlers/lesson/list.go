package lesson

import (
	"net/http"

	"entgo.io/ent/dialect/sql"
	"github.com/csotherden/mandarin-flashcards/ent/lesson"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/study"
)

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	rows, err := h.db.Lesson.Query().
		Where(lesson.HasOwnerWith(user.ID(userID))).
		WithCardSets().
		Order(lesson.ByCreatedAt(sql.OrderDesc())).
		All(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not load lessons", err)
		return
	}
	out := make([]models.LessonResponse, 0, len(rows))
	for _, row := range rows {
		deck, err := study.LoadLesson(r.Context(), h.db, userID, row.ID)
		if err != nil {
			handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not load lessons", err)
			return
		}
		item := detailFrom(row, deck).LessonResponse
		out = append(out, item)
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, out)
}
