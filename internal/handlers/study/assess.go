package study

import (
	"errors"
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/ent/cardattemptlog"
	entattempt "github.com/csotherden/mandarin-flashcards/ent/lessonattempt"
	"github.com/csotherden/mandarin-flashcards/internal/audio"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/llm"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/present"
	"github.com/csotherden/mandarin-flashcards/internal/streak"
	"github.com/csotherden/mandarin-flashcards/internal/toneperfect"
	"github.com/csotherden/mandarin-flashcards/internal/xp"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *Handler) Assess(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	if !h.limit.Allow(userID.String()) {
		handlerutils.RespondWithError(w, h.logger, http.StatusTooManyRequests, handlerutils.CodeNotSet, "Too many pronunciation checks. Wait a minute and try again", nil)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20+4096)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "The recording is too large or invalid", err)
		return
	}
	attemptID, err := uuid.Parse(r.FormValue("lesson_attempt_id"))
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Invalid drill id", err)
		return
	}
	cardID, err := uuid.Parse(r.FormValue("flashcard_id"))
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Invalid flashcard id", err)
		return
	}
	attempt, err := h.ownedAttempt(r.Context(), userID, attemptID)
	if err != nil {
		if ent.IsNotFound(err) {
			handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Drill not found", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not grade pronunciation", err)
		return
	}
	if attempt.CompletedAt != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusConflict, handlerutils.CodeNotSet, "This drill is already finished", nil)
		return
	}
	if attempt.Mode != entattempt.ModeImageSpeak && attempt.Mode != entattempt.ModeHanziSpeak {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "This drill does not use speaking", nil)
		return
	}
	if !containsCard(attempt.CardIds, cardID) {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "That card is not in this drill", nil)
		return
	}
	card, err := h.db.Flashcard.Get(r.Context(), cardID)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Flashcard not found", err)
		return
	}
	wav, err := audio.FromForm(r)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, err.Error(), err)
		return
	}
	// The numbered form is what TonePerfect accepts; cards from before it fall back to the
	// display pinyin, which the client sends only if TonePerfect would accept it.
	reading := card.TonalPinyin
	if reading == "" {
		reading = card.Pinyin
	}
	grade, err := h.tones.Assess(r.Context(), wav, card.Hanzi, reading)
	if err != nil {
		if errors.Is(err, toneperfect.ErrNotConfigured) {
			handlerutils.RespondWithError(w, h.logger, http.StatusServiceUnavailable, handlerutils.CodeNotSet, "Pronunciation grading is not configured", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusBadGateway, handlerutils.CodeNotSet, "Pronunciation grading failed. Try again", err)
		return
	}

	tip := ""
	if grade.Assessable && grade.Overall < models.CoachingThreshold && h.llm != nil && h.llm.Enabled() {
		coached, coachErr := h.llm.Coach(r.Context(), card.Hanzi, grade.Overall, grade.Syllables)
		if coachErr != nil && !errors.Is(coachErr, llm.ErrNotConfigured) {
			h.logger.Warn("coaching tip failed", zap.Error(coachErr))
		} else {
			tip = coached
		}
	}
	feedback := &models.StoredFeedback{
		Assessable: grade.Assessable,
		Reason:     grade.Reason,
		Syllables:  grade.Syllables,
		Provider:   grade.Provider,
	}
	create := h.db.CardAttemptLog.Create().
		SetOwnerID(userID).
		SetUserID(userID).
		SetLessonAttemptID(attempt.ID).
		SetFlashcardID(card.ID).
		SetMode(cardattemptlog.Mode(attempt.Mode)).
		SetOverallScore(grade.Overall).
		SetRawAPIFeedback(feedback)
	if grade.ToneAccuracy != nil {
		create.SetToneAccuracyScore(*grade.ToneAccuracy)
	}
	if tip != "" {
		create.SetLlmCoachingTip(tip)
	}
	log, err := create.Save(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not save the grade", err)
		return
	}
	if err := streak.Record(r.Context(), h.db, userID, log.CreatedAt); err != nil {
		// The grade is saved; a missed streak update is fixed by the next grade today.
		h.logger.Warn("Could not update the streak", zap.Error(err))
	}
	earned, err := xp.Award(r.Context(), h.db, userID, log, card.Hanzi, xp.SpeakTier(grade.Assessable, grade.Overall, grade.Syllables), grade.Overall)
	if err != nil {
		// The grade stands; XP for it is lost rather than failing the take.
		h.logger.Warn("Could not award XP", zap.Error(err))
	} else {
		log.XpEarned, log.XpTier = earned.Earned, earned.Tier
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, models.AssessResponse{
		Result:     present.CardResult(log),
		Assessable: grade.Assessable,
		Reason:     grade.Reason,
		XPTotal:    earned.Total,
	})
}
