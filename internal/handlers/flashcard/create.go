package flashcard

import (
	"errors"
	"net/http"

	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/csotherden/mandarin-flashcards/internal/own"
	"github.com/csotherden/mandarin-flashcards/internal/present"
)

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		handlerutils.RespondWithError(w, h.logger, http.StatusUnauthorized, handlerutils.CodeNotSet, "Unauthorized", nil)
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
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not create the flashcard", err)
		return
	}
	defer func() { _ = tx.Rollback() }()

	create := tx.Flashcard.Create().
		SetHanzi(req.Hanzi).
		SetPinyin(req.Pinyin).
		SetTonalPinyin(req.TonalPinyin).
		SetEnglish(req.English).
		SetExampleSentenceHanzi(req.ExampleSentenceHanzi).
		SetExampleSentencePinyin(req.ExampleSentencePinyin).
		SetImageURL(req.ImageURL).
		SetOwnerID(userID)
	if req.UnsplashAttribution != nil && req.ImageURL != "" {
		create.SetUnsplashAttribution(req.UnsplashAttribution)
	}
	row, err := create.Save(r.Context())
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not create the flashcard", err)
		return
	}
	linked, err := h.linkCard(r.Context(), tx.Client(), userID, row, req)
	if err != nil {
		if errors.Is(err, own.ErrNotFound) {
			handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "A category or set does not belong to this account", nil)
			return
		}
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not create the flashcard", err)
		return
	}
	if err := tx.Commit(); err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not create the flashcard", err)
		return
	}
	h.trackDownload(r.Context(), req.UnsplashAttribution)
	handlerutils.RespondWithJSON(w, h.logger, http.StatusCreated, present.Flashcard(linked))
}
