package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	status := http.StatusOK
	body := map[string]any{"status": "healthy", "checks": map[string]string{"database": "ok"}}
	if err := h.db.PingContext(ctx); err != nil {
		h.logger.Warn("health check: database unhealthy", zap.Error(err))
		status = http.StatusServiceUnavailable
		body = map[string]any{"status": "unhealthy", "checks": map[string]string{"database": "error"}}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		h.logger.Error("failed to encode health response", zap.Error(err))
	}
}
