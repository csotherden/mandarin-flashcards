package flashcard

import (
	"errors"
	"net/http"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/own"
	"github.com/csotherden/mandarin-flashcards/internal/present"
	"go.uber.org/zap"
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
	var req models.FlashcardInput
	if !handlerutils.DecodeJSON(w, r, h.logger, &req) {
		return
	}
	req, err := normalizeCard(req)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, err.Error(), nil)
		return
	}

	tx, err := h.db.Tx(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the flashcard", err)
		return
	}
	defer func() { _ = tx.Rollback() }()

	current, err := h.owned(r.Context(), tx.Client(), userID, id)
	if err != nil {
		if ent.IsNotFound(err) {
			handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Flashcard not found", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the flashcard", err)
		return
	}

	update := tx.Flashcard.UpdateOneID(current.ID).
		SetHanzi(req.Hanzi).
		SetPinyin(req.Pinyin).
		SetTonalPinyin(req.TonalPinyin).
		SetEnglish(req.English).
		SetExampleSentenceHanzi(req.ExampleSentenceHanzi).
		SetExampleSentencePinyin(req.ExampleSentencePinyin).
		SetImageURL(req.ImageURL)
	if req.ImageURL != "" {
		update.ClearImageObjectKey()
		if req.UnsplashAttribution != nil {
			update.SetUnsplashAttribution(req.UnsplashAttribution)
		} else {
			update.ClearUnsplashAttribution()
		}
	} else if current.ImageObjectKey == nil {
		update.ClearUnsplashAttribution()
	}
	if _, err := update.Save(r.Context()); err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the flashcard", err)
		return
	}
	linked, err := h.linkCard(r.Context(), tx.Client(), userID, current, req)
	if err != nil {
		if errors.Is(err, own.ErrNotFound) {
			handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "A category or set does not belong to this account", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the flashcard", err)
		return
	}
	if err := tx.Commit(); err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not update the flashcard", err)
		return
	}
	if req.ImageURL != "" && current.ImageObjectKey != nil {
		if err := h.files.DeleteFile(r.Context(), *current.ImageObjectKey); err != nil {
			h.logger.Warn("failed to delete replaced card image", zap.Error(err))
		}
	}
	h.trackDownload(r.Context(), req.UnsplashAttribution)
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, present.Flashcard(linked))
}
