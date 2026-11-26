package flashcard

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/internal/handlerutils"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
	"github.com/csotherden/mandarin-flashcards/internal/present"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
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
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not save the image", err)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Image upload is too large or invalid", err)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Choose an image file", err)
		return
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Could not read the image", err)
		return
	}
	ext := strings.ToLower(path.Ext(header.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
	default:
		handlerutils.RespondWithError(w, h.logger, http.StatusBadRequest, handlerutils.CodeNotSet, "Use a jpg, png, webp, or gif image", nil)
		return
	}
	key := fmt.Sprintf("users/%s/cards/%s/%s%s", userID, row.ID, uuid.NewString(), ext)
	if err := h.files.UploadFile(r.Context(), key, content); err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not save the image", err)
		return
	}
	updated, err := h.db.Flashcard.UpdateOneID(row.ID).
		SetImageObjectKey(key).
		SetImageURL("").
		ClearUnsplashAttribution().
		Save(r.Context())
	if err != nil {
		_ = h.files.DeleteFile(r.Context(), key)
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not save the image", err)
		return
	}
	if row.ImageObjectKey != nil && *row.ImageObjectKey != key {
		if err := h.files.DeleteFile(r.Context(), *row.ImageObjectKey); err != nil {
			h.logger.Warn("failed to delete previous card image", zap.Error(err))
		}
	}
	fresh, err := h.owned(r.Context(), h.db, userID, updated.ID)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusInternalServerError, handlerutils.CodeNotSet, "Could not save the image", err)
		return
	}
	handlerutils.RespondWithJSON(w, h.logger, http.StatusOK, present.Flashcard(fresh))
}

func (h *Handler) DownloadImage(w http.ResponseWriter, r *http.Request) {
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
	if err != nil || row.ImageObjectKey == nil || *row.ImageObjectKey == "" {
		handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Image not found", err)
		return
	}
	content, err := h.files.DownloadFile(r.Context(), *row.ImageObjectKey)
	if err != nil {
		handlerutils.RespondWithError(w, h.logger, http.StatusNotFound, handlerutils.CodeNotSet, "Image not found", err)
		return
	}
	w.Header().Set("Content-Type", contentType(*row.ImageObjectKey))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func contentType(key string) string {
	switch strings.ToLower(path.Ext(key)) {
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return "image/jpeg"
	}
}
