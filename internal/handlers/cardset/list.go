package cardset

import (
	"net/http"

	"entgo.io/ent/dialect/sql"
	"github.com/csotherden/mandarin-flashcards/ent/cardset"
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
	rows, err := h.db.CardSet.Query().
		Where(cardset.HasOwnerWith(user.ID(userID))).
		WithMemberships().
		Order(cardset.ByCreatedAt(sql.OrderDesc())).
		All(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not load sets", err)
		return
	}
	out := make([]models.SetResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, present.Set(row, len(row.Edges.Memberships)))
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, out)
}
