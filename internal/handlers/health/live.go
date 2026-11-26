package health

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

func (h *Handler) Live(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "healthy"}); err != nil {
		h.logger.Error("failed to encode health response", zap.Error(err))
	}
}
