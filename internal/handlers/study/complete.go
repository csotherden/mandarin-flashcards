package study

import (
	"net/http"
	"time"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
)

func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	id, ok := handlerutils.PathID(w, r, h.logger, "id")
	if !ok {
		return
	}
	attempt, err := h.ownedAttempt(r.Context(), userID, id)
	if err != nil {
		if ent.IsNotFound(err) {
			handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Drill not found", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not finish the drill", err)
		return
	}
	if attempt.CompletedAt == nil {
		current, err := h.session(r.Context(), userID, attempt)
		if err != nil {
			handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not finish the drill", err)
			return
		}
		scores := map[string]float64{}
		for _, result := range current.Results {
			scores[result.FlashcardID.String()] = result.OverallScore
		}
		var sum float64
		for _, cardID := range attempt.CardIds {
			sum += scores[cardID.String()]
		}
		overall := 0.0
		if len(attempt.CardIds) > 0 {
			overall = sum / float64(len(attempt.CardIds))
		}
		now := time.Now()
		attempt, err = h.db.LessonAttempt.UpdateOneID(attempt.ID).
			SetOverallScore(overall).
			SetCompletedAt(now).
			Save(r.Context())
		if err != nil {
			handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not finish the drill", err)
			return
		}
	}
	body, err := h.session(r.Context(), userID, attempt)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not finish the drill", err)
		return
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, body)
}
