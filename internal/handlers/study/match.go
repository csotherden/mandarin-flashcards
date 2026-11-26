package study

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/ent/cardattemptlog"
	entattempt "github.com/csotherden/mandarin-flashcards/ent/lessonattempt"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/present"
)

func (h *Handler) Match(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	id, ok := handlerutils.PathID(w, r, h.logger, "id")
	if !ok {
		return
	}
	var req models.MatchInput
	if !handlerutils.DecodeJSON(w, r, h.logger, &req) {
		return
	}
	attempt, err := h.ownedAttempt(r.Context(), userID, id)
	if err != nil {
		if ent.IsNotFound(err) {
			handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Drill not found", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not record the match", err)
		return
	}
	if attempt.CompletedAt != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusConflict, handlerutils.CodeNotSet, "This drill is already finished", nil)
		return
	}
	if attempt.Mode != entattempt.ModePictureHanziMatch {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "This drill is not a matching round", nil)
		return
	}
	if !containsCard(attempt.CardIds, req.FlashcardID) {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "That card is not in this drill", nil)
		return
	}
	if req.MismatchCount < 0 {
		req.MismatchCount = 0
	}
	score := 0.0
	if req.Correct {
		score = 100 - float64(req.MismatchCount)*25
		if score < 0 {
			score = 0
		}
	}
	log, err := h.db.CardAttemptLog.Create().
		SetOwnerID(userID).
		SetUserID(userID).
		SetLessonAttemptID(attempt.ID).
		SetFlashcardID(req.FlashcardID).
		SetMode(cardattemptlog.ModePictureHanziMatch).
		SetOverallScore(score).
		SetRawAPIFeedback(&models.StoredFeedback{Match: &models.MatchFeedback{Correct: req.Correct, MismatchCount: req.MismatchCount}}).
		Save(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not record the match", err)
		return
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusCreated, present.CardResult(log))
}
