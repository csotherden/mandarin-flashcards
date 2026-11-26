package cardset

import (
	"errors"
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent"
	entset "github.com/csotherden/mandarin-flashcards/ent/cardset"
	"github.com/csotherden/mandarin-flashcards/ent/setflashcard"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/own"
	"github.com/csotherden/mandarin-flashcards/internal/present"
	"github.com/csotherden/mandarin-flashcards/internal/study"
)

func (h *Handler) ReplaceCards(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	id, ok := handlerutils.PathID(w, r, h.logger, "id")
	if !ok {
		return
	}
	var req models.SetCardsInput
	if !handlerutils.DecodeJSON(w, r, h.logger, &req) {
		return
	}
	req.FlashcardIDs = own.Unique(req.FlashcardIDs)

	tx, err := h.db.Tx(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the set", err)
		return
	}
	defer func() { _ = tx.Rollback() }()

	n, err := tx.CardSet.Query().Where(entset.ID(id), entset.HasOwnerWith(user.ID(userID))).Count(r.Context())
	if err != nil || n == 0 {
		if err != nil && !ent.IsNotFound(err) {
			handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the set", err)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Set not found", nil)
		return
	}
	if err := own.Flashcards(r.Context(), tx.Client(), userID, req.FlashcardIDs); err != nil {
		if errors.Is(err, own.ErrNotFound) {
			handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "A card does not belong to this account", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the set", err)
		return
	}
	if _, err := tx.SetFlashcard.Delete().Where(setflashcard.SetID(id)).Exec(r.Context()); err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the set", err)
		return
	}
	for i, cardID := range req.FlashcardIDs {
		if _, err := tx.SetFlashcard.Create().SetSetID(id).SetFlashcardID(cardID).SetSortOrder(i).Save(r.Context()); err != nil {
			handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the set", err)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the set", err)
		return
	}
	row, err := h.db.CardSet.Get(r.Context(), id)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the set", err)
		return
	}
	cards, err := study.LoadSet(r.Context(), h.db, userID, id)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the set", err)
		return
	}
	detail := models.SetDetail{SetResponse: present.Set(row, len(cards)), Cards: make([]models.FlashcardResponse, 0, len(cards))}
	for _, card := range cards {
		detail.Cards = append(detail.Cards, present.Flashcard(card))
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, detail)
}
