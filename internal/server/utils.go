package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func writeJSON(w http.ResponseWriter, logger *slog.Logger, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Warn("writing response", "err", err)
	}
}
