package server

import (
	"log/slog"
	"net/http"
)

func NewHandler(
	logger *slog.Logger,
	jobStore *JobStore,
	runner *Runner,
	dataDir string,
) http.Handler {

	mux := http.NewServeMux()
	mux.Handle("GET /healthz", handleHealthz())
	mux.Handle("POST /upload", handleUpload(logger, jobStore, runner, dataDir, nWorkers))
	mux.Handle("GET /jobs/{id}", handleStatus(logger, jobStore))

	return mux
}
