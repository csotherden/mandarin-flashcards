package flashcard

import (
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
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
