package category

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent/category"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/present"
)

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	rows, err := h.db.Category.Query().
		Where(category.HasOwnerWith(user.ID(userID))).
		Order(category.ByName()).
		All(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not load categories", err)
		return
	}
	out := make([]models.CategoryResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, present.Category(row))
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, out)
}
