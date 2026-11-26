package lesson

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent"
	entlesson "github.com/csotherden/mandarin-flashcards/ent/lesson"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/own"
	"github.com/csotherden/mandarin-flashcards/internal/study"
)

func (h *Handler) ReplaceDisabled(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	id, ok := handlerutils.PathID(w, r, h.logger, "id")
	if !ok {
		return
	}
	var req models.IDsInput
	if !handlerutils.DecodeJSON(w, r, h.logger, &req) {
		return
	}
	exists, err := h.db.Lesson.Query().Where(entlesson.ID(id), entlesson.HasOwnerWith(user.ID(userID))).Exist(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the lesson", err)
		return
	}
	if !exists {
		handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Lesson not found", nil)
		return
	}
	deck, err := study.LoadLesson(r.Context(), h.db, userID, id)
	if err != nil {
		if ent.IsNotFound(err) {
			handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Lesson not found", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the lesson", err)
		return
	}
	allowed := map[string]struct{}{}
	for _, card := range deck.Cards {
		allowed[card.ID.String()] = struct{}{}
	}
	req.FlashcardIDs = own.Unique(req.FlashcardIDs)
	for _, cardID := range req.FlashcardIDs {
		if _, ok := allowed[cardID.String()]; !ok {
			handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Disabled cards have to belong to this lesson", nil)
			return
		}
	}
	if _, err := h.db.Lesson.UpdateOneID(id).ClearDisabledCards().AddDisabledCardIDs(req.FlashcardIDs...).Save(r.Context()); err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the lesson", err)
		return
	}
	h.writeDetail(w, r, userID, id, http.StatusOK)
}
