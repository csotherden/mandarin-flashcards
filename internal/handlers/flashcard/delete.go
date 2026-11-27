package flashcard

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent"
	entflashcard "github.com/csotherden/mandarin-flashcards/ent/flashcard"
	"github.com/csotherden/mandarin-flashcards/ent/user"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	id, ok := handlerutils.PathID(w, r, h.logger, "id")
	if !ok {
		return
	}
	row, err := h.owned(r.Context(), h.db, userID, id)
	if err != nil {
		if ent.IsNotFound(err) {
			handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Flashcard not found", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not delete the flashcard", err)
		return
	}
	if err := h.db.Flashcard.DeleteOneID(row.ID).Exec(r.Context()); err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not delete the flashcard", err)
		return
	}
	if row.ImageObjectKey != nil {
		if err := h.files.DeleteFile(r.Context(), *row.ImageObjectKey); err != nil {
			h.logger.Warn("failed to delete card image", zap.Error(err))
		}
	}
	handlerutils.RespondWithNoContent(w)
}

// BulkDeleteRequest lists cards to delete (at most 500 at a time).
type BulkDeleteRequest struct {
	IDs []uuid.UUID `json:"ids"`
}

// BulkDelete is POST /api/flashcards/bulk-delete {ids}. Only the caller's own cards are
// deleted; ids that aren't theirs (or are already gone) are skipped. Returns how many went.
func (h *Handler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	var req BulkDeleteRequest
	if !handlerutils.DecodeJSON(w, r, h.logger, &req) {
		return
	}
	if len(req.IDs) == 0 || len(req.IDs) > 500 {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Choose 1 to 500 cards", nil)
		return
	}
	ctx := r.Context()
	rows, err := h.db.Flashcard.Query().
		Where(entflashcard.IDIn(req.IDs...), entflashcard.HasOwnerWith(user.ID(userID))).
		All(ctx)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not delete the cards", err)
		return
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	n, err := h.db.Flashcard.Delete().Where(entflashcard.IDIn(ids...), entflashcard.HasOwnerWith(user.ID(userID))).Exec(ctx)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not delete the cards", err)
		return
	}
	for _, row := range rows {
		if row.ImageObjectKey != nil {
			if err := h.files.DeleteFile(ctx, *row.ImageObjectKey); err != nil {
				h.logger.Warn("failed to delete card image", zap.Error(err))
			}
		}
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, map[string]int{"deleted": n})
}
