package cardset

import (
	"net/http"

	entset "github.com/csotherden/mandarin-flashcards/ent/cardset"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/present"
)

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	id, ok := handlerutils.PathID(w, r, h.logger, "id")
	if !ok {
		return
	}
	var req models.SetInput
	if !handlerutils.DecodeJSON(w, r, h.logger, &req) {
		return
	}
	req, msg := cleanSet(req)
	if msg != "" {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, msg, nil)
		return
	}
	n, err := h.db.CardSet.Update().
		Where(entset.ID(id), entset.HasOwnerWith(user.ID(userID))).
		SetTitle(req.Title).
		SetDescription(req.Description).
		Save(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the set", err)
		return
	}
	if n == 0 {
		handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Set not found", nil)
		return
	}
	row, err := h.db.CardSet.Get(r.Context(), id)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the set", err)
		return
	}
	count, err := row.QueryMemberships().Count(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the set", err)
		return
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, present.Set(row, count))
}
