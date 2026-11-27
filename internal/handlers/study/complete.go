package study

import (
	"github.com/csotherden/mandarin-flashcards/internal/models"
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
		overall := SessionScore(current.Results)
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

// SessionScore is the average of the cards actually graded (each card's latest take). Cards that
// were skipped, or never reached because the session ended early, don't count as zeros.
func SessionScore(results []models.CardResult) float64 {
	if len(results) == 0 {
		return 0
	}
	var sum float64
	for _, r := range results {
		sum += r.OverallScore
	}
	return sum / float64(len(results))
}
