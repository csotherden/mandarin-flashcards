package flashcard

import (
	"net/http"

	"entgo.io/ent/dialect/sql"
	entcategory "github.com/csotherden/mandarin-flashcards/ent/category"
	entflashcard "github.com/csotherden/mandarin-flashcards/ent/flashcard"
	"github.com/csotherden/mandarin-flashcards/ent/setflashcard"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/present"
	"github.com/google/uuid"
)

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	query := cardQuery(h.db).
		Where(entflashcard.HasOwnerWith(user.ID(userID))).
		Order(entflashcard.ByCreatedAt(sql.OrderDesc()))
	if term := r.URL.Query().Get("q"); term != "" {
		query = query.Where(entflashcard.Or(
			entflashcard.HanziContainsFold(term),
			entflashcard.PinyinContainsFold(term),
			entflashcard.EnglishContainsFold(term),
		))
	}
	if raw := r.URL.Query().Get("category_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Invalid category id", err)
			return
		}
		query = query.Where(entflashcard.HasCategoriesWith(entcategory.ID(id)))
	}
	if raw := r.URL.Query().Get("set_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Invalid set id", err)
			return
		}
		query = query.Where(entflashcard.HasSetMembershipsWith(setflashcard.SetID(id)))
	}
	rows, err := query.All(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not load flashcards", err)
		return
	}
	out := make([]models.FlashcardResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, present.Flashcard(row))
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, out)
}
