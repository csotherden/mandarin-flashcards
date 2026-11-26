package images

import (
	"errors"
	"net/http"
	"strings"

	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/unsplash"
)

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	if !h.limit.Allow(userID.String()) {
		handlerutils.RespondWithError(w, h.logger, http.StatusTooManyRequests, handlerutils.CodeNotSet, "Too many image searches. Wait a minute and try again", nil)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Enter a search term", nil)
		return
	}
	photos, err := h.unsplash.Search(r.Context(), query)
	if err != nil {
		if errors.Is(err, unsplash.ErrNotConfigured) {
			handlerutils.RespondWithError(w, h.logger, http.StatusServiceUnavailable, handlerutils.CodeNotSet, "Image search is not configured", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusBadGateway, handlerutils.CodeNotSet, "Could not search images", err)
		return
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, photos)
}
