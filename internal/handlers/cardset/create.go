package cardset

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/present"
)

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
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
	row, err := h.db.CardSet.Create().
		SetTitle(req.Title).
		SetDescription(req.Description).
		SetOwnerID(userID).
		Save(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not create the set", err)
		return
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusCreated, present.Set(row, 0))
}
