package server

import (
	"log/slog"
	"net/http"
)

func newHandler(
	logger *slog.Logger,
	jobStore *JobStore,
	runner *Runner,
	dataDir string,
) http.Handler {

	mux := http.NewServeMux()
	mux.Handle("GET /healthz", handleHealths())
	mux.Handle("POST /upload", handleUpload(logger, jobStore, runner, dataDir))
	mux.Handle("GET /jobs/{id}", handleStatus(logger, jobStore))

	return mux
}
