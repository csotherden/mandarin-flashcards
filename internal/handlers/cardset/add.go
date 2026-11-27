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
)

// AddCardsResponse says how many of the requested cards were new to the set.
type AddCardsResponse struct {
	Added     int `json:"added"`
	CardCount int `json:"card_count"`
}

// AddCards is POST /api/sets/{id}/cards/add {flashcard_ids}: appends cards to the end of a set,
// skipping any it already holds. Unlike PUT .../cards it never drops or reorders cards, so it's
// safe for bulk "add to set" from the library.
func (h *Handler) AddCards(w http.ResponseWriter, r *http.Request) {
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
	ctx := r.Context()
	fail := func(err error) {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not add the cards", err)
	}
	tx, err := h.db.Tx(ctx)
	if err != nil {
		fail(err)
		return
	}
	defer func() { _ = tx.Rollback() }()
	// Lock the set row so concurrent appends number their cards one after another.
	if _, err := tx.CardSet.Query().Where(entset.ID(id), entset.HasOwnerWith(user.ID(userID))).ForUpdate().Only(ctx); err != nil {
		if ent.IsNotFound(err) {
			handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Set not found", nil)
			return
		}
		fail(err)
		return
	}
	if err := own.Flashcards(ctx, tx.Client(), userID, req.FlashcardIDs); err != nil {
		if errors.Is(err, own.ErrNotFound) {
			handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "A card does not belong to this account", nil)
			return
		}
		fail(err)
		return
	}
	members, err := tx.SetFlashcard.Query().Where(setflashcard.SetID(id)).All(ctx)
	if err != nil {
		fail(err)
		return
	}
	have := map[string]bool{}
	next := 0
	for _, m := range members {
		have[m.FlashcardID.String()] = true
		next = max(next, m.SortOrder+1)
	}
	added := 0
	for _, cardID := range req.FlashcardIDs {
		if have[cardID.String()] {
			continue
		}
		if err := tx.SetFlashcard.Create().SetSetID(id).SetFlashcardID(cardID).SetSortOrder(next).Exec(ctx); err != nil {
			fail(err)
			return
		}
		next++
		added++
	}
	if err := tx.Commit(); err != nil {
		fail(err)
		return
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, AddCardsResponse{Added: added, CardCount: len(members) + added})
}
