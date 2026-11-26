package lesson

import (
	"errors"
	"net/http"

	entlesson "github.com/csotherden/mandarin-flashcards/ent/lesson"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/own"
	"github.com/csotherden/mandarin-flashcards/internal/study"
	"github.com/google/uuid"
)

func (h *Handler) ReplaceSets(w http.ResponseWriter, r *http.Request) {
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
	req.SetIDs = own.Unique(req.SetIDs)
	if err := own.CardSets(r.Context(), h.db, userID, req.SetIDs); err != nil {
		if errors.Is(err, own.ErrNotFound) {
			handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "A set does not belong to this account", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the lesson", err)
		return
	}
	n, err := h.db.Lesson.Update().
		Where(entlesson.ID(id), entlesson.HasOwnerWith(user.ID(userID))).
		ClearCardSets().
		AddCardSetIDs(req.SetIDs...).
		Save(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the lesson", err)
		return
	}
	if n == 0 {
		handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Lesson not found", nil)
		return
	}
	deck, err := study.LoadLesson(r.Context(), h.db, userID, id)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the lesson", err)
		return
	}
	keep := make([]uuid.UUID, 0)
	for _, card := range deck.Cards {
		if _, off := deck.Disabled[card.ID]; off {
			keep = append(keep, card.ID)
		}
	}
	if _, err := h.db.Lesson.UpdateOneID(id).ClearDisabledCards().AddDisabledCardIDs(keep...).Save(r.Context()); err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the lesson", err)
		return
	}
	h.writeDetail(w, r, userID, id, http.StatusOK)
}
