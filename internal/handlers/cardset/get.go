package cardset

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent"
	entset "github.com/csotherden/mandarin-flashcards/ent/cardset"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/present"
	"github.com/csotherden/mandarin-flashcards/internal/study"
)

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	id, ok := handlerutils.PathID(w, r, h.logger, "id")
	if !ok {
		return
	}
	row, err := h.db.CardSet.Query().
		Where(entset.ID(id), entset.HasOwnerWith(user.ID(userID))).
		Only(r.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Set not found", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not load the set", err)
		return
	}
	cards, err := study.LoadSet(r.Context(), h.db, userID, id)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not load the set", err)
		return
	}
	detail := models.SetDetail{SetResponse: present.Set(row, len(cards)), Cards: make([]models.FlashcardResponse, 0, len(cards))}
	for _, card := range cards {
		detail.Cards = append(detail.Cards, present.Flashcard(card))
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, detail)
}
