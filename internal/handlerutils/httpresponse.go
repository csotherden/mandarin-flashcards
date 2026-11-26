package handlerutils

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/csotherden/mandarin-flashcards/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

const (
	CodeNotSet           = ""
	internalErrorMessage = "Something went wrong. Please try again"
	internalErrorJSON    = `{"error":"Something went wrong. Please try again","message":"Something went wrong. Please try again","code":"internal_error"}`
)

func RespondWithJSON(w http.ResponseWriter, logger *zap.Logger, statusCode int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		logger.Error("failed to marshal json response", zap.Error(err))
		statusCode = http.StatusInternalServerError
		response = []byte(internalErrorJSON)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(statusCode)
	if _, err := w.Write(response); err != nil {
		logger.Error("failed to write json response", zap.Error(err))
	}
}

func RespondWithError(w http.ResponseWriter, logger *zap.Logger, statusCode int, code, message string, err error) {
	if strings.TrimSpace(code) == "" {
		code = models.GenericErrorCode(statusCode)
	}
	if err != nil {
		logger.Error("API error", zap.String("message", message), zap.String("code", code), zap.Int("status_code", statusCode), zap.Error(err))
	} else {
		logger.Info("API error", zap.String("message", message), zap.String("code", code), zap.Int("status_code", statusCode))
	}
	RespondWithJSON(w, logger, statusCode, models.ErrorResponse{
		Error:   message,
		Message: message,
		Code:    code,
	})
}

func RespondWithNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func PathID(w http.ResponseWriter, r *http.Request, logger *zap.Logger, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(mux.Vars(r)[name])
	if err != nil {
		RespondWithError(w, logger, http.StatusBadRequest, CodeNotSet, "Invalid id", err)
		return uuid.Nil, false
	}
	return id, true
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, logger *zap.Logger, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		RespondWithError(w, logger, http.StatusBadRequest, CodeNotSet, "Invalid request payload", err)
		return false
	}
	return true
}
