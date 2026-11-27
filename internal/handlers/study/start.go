package study

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent"
	entset "github.com/csotherden/mandarin-flashcards/ent/cardset"
	entflashcard "github.com/csotherden/mandarin-flashcards/ent/flashcard"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/own"
	deck "github.com/csotherden/mandarin-flashcards/internal/study"
	"github.com/google/uuid"
)

func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	var req models.StartSessionRequest
	if !handlerutils.DecodeJSON(w, r, h.logger, &req) {
		return
	}
	mode, ok := parseMode(req.Mode)
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Choose a lesson mode", nil)
		return
	}
	if req.CardCount < 0 {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Card count cannot be negative", nil)
		return
	}
	sources := 0
	if req.LessonID != nil {
		sources++
	}
	if req.SetID != nil {
		sources++
	}
	if req.SourceAttemptID != nil {
		sources++
	}
	req.SetIDs = own.Unique(req.SetIDs)
	if len(req.SetIDs) > 0 {
		sources++
	}
	if req.All {
		sources++
	}
	if sources != 1 {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Choose sets, a saved session, all cards, or missed cards from a previous drill", nil)
		return
	}

	var cards []*ent.Flashcard
	var lessonID *uuid.UUID
	switch {
	case req.LessonID != nil:
		loaded, err := deck.LoadLesson(r.Context(), h.db, userID, *req.LessonID)
		if err != nil {
			if ent.IsNotFound(err) {
				handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Lesson not found", nil)
				return
			}
			handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not start the drill", err)
			return
		}
		for _, card := range loaded.Cards {
			if _, off := loaded.Disabled[card.ID]; !off {
				cards = append(cards, card)
			}
		}
		lessonID = req.LessonID
	case req.SetID != nil:
		exists, err := h.db.CardSet.Query().Where(entset.ID(*req.SetID), entset.HasOwnerWith(user.ID(userID))).Exist(r.Context())
		if err != nil {
			handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not start the drill", err)
			return
		}
		if !exists {
			handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Set not found", nil)
			return
		}
		cards, err = deck.LoadSet(r.Context(), h.db, userID, *req.SetID)
		if err != nil {
			handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not start the drill", err)
			return
		}
	case len(req.SetIDs) > 0:
		if err := own.CardSets(r.Context(), h.db, userID, req.SetIDs); err != nil {
			handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Set not found", nil)
			return
		}
		seen := map[uuid.UUID]bool{}
		for _, setID := range req.SetIDs {
			loaded, err := deck.LoadSet(r.Context(), h.db, userID, setID)
			if err != nil {
				handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not start the drill", err)
				return
			}
			for _, card := range loaded {
				if !seen[card.ID] {
					seen[card.ID] = true
					cards = append(cards, card)
				}
			}
		}
	case req.All:
		ids, err := h.db.Flashcard.Query().Where(entflashcard.HasOwnerWith(user.ID(userID))).IDs(r.Context())
		if err == nil {
			cards, err = deck.LoadOwned(r.Context(), h.db, userID, ids)
		}
		if err != nil {
			handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not start the drill", err)
			return
		}
	default:
		source, err := h.ownedAttempt(r.Context(), userID, *req.SourceAttemptID)
		if err != nil {
			if ent.IsNotFound(err) {
				handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Drill not found", nil)
				return
			}
			handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not start the drill", err)
			return
		}
		previous, err := h.session(r.Context(), userID, source)
		if err != nil {
			handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not start the drill", err)
			return
		}
		scores := map[uuid.UUID]float64{}
		for _, result := range previous.Results {
			scores[result.FlashcardID] = result.OverallScore
		}
		missed := make([]uuid.UUID, 0)
		for _, id := range source.CardIds {
			score, seen := scores[id]
			if !seen || score < models.CoachingThreshold {
				missed = append(missed, id)
			}
		}
		if len(missed) == 0 {
			handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Nothing in that drill scored under 85", nil)
			return
		}
		cards, err = deck.LoadOwned(r.Context(), h.db, userID, missed)
		if err != nil {
			handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not start the drill", err)
			return
		}
	}
	if len(cards) == 0 {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "This drill has no cards", nil)
		return
	}

	ids := make([]uuid.UUID, len(cards))
	for i, card := range cards {
		ids[i] = card.ID
	}
	stats, err := deck.History(r.Context(), h.db, userID, ids)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not start the drill", err)
		return
	}
	chosen := deck.Prioritize(stats, req.CardCount)
	requested := req.CardCount
	if requested == 0 {
		requested = len(chosen)
	}
	create := h.db.LessonAttempt.Create().
		SetOwnerID(userID).
		SetMode(mode).
		SetCardCountRequested(requested).
		SetCardIds(chosen).
		SetNillableLessonID(lessonID)
	attempt, err := create.Save(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not start the drill", err)
		return
	}
	body, err := h.session(r.Context(), userID, attempt)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not start the drill", err)
		return
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusCreated, body)
}
