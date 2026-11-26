package category

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent"
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
	var req models.CategoryInput
	if !handlerutils.DecodeJSON(w, r, h.logger, &req) {
		return
	}
	req, msg := cleanCategory(req)
	if msg != "" {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, msg, nil)
		return
	}
	row, err := h.db.Category.Create().
		SetName(req.Name).
		SetColor(req.Color).
		SetOwnerID(userID).
		Save(r.Context())
	if err != nil {
		if ent.IsConstraintError(err) {
			handlerutils.RespondWithError(w, h.logger, http.StatusConflict, handlerutils.CodeNotSet, "You already have a category with that name", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not create the category", err)
		return
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusCreated, present.Category(row))
}
