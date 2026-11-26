package autofill

import (
	"errors"
	"net/http"
	"strings"

	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/llm"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
)

func (h *Handler) Autofill(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
		return
	}
	if !h.limit.Allow(userID.String()) {
		handlerutils.RespondWithError(w, h.logger, http.StatusTooManyRequests, handlerutils.CodeNotSet, "Too many suggestions. Wait a minute and try again", nil)
		return
	}
	var req models.AutofillRequest
	if !handlerutils.DecodeJSON(w, r, h.logger, &req) {
		return
	}
	req.Hanzi = strings.TrimSpace(req.Hanzi)
	req.English = strings.TrimSpace(req.English)
	if req.Hanzi == "" && req.English == "" {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Enter hanzi or an English word", nil)
		return
	}
	filled, err := h.llm.Autofill(r.Context(), req.Hanzi, req.English)
	if err != nil {
		if errors.Is(err, llm.ErrNotConfigured) {
			handlerutils.RespondWithError(w, h.logger, http.StatusServiceUnavailable, handlerutils.CodeNotSet, "Card suggestions are not configured", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusBadGateway, handlerutils.CodeNotSet, "Could not suggest card details", err)
		return
	}
	if req.Hanzi != "" {
		filled.Hanzi = req.Hanzi
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, filled)
}
