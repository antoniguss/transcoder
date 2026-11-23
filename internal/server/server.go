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
	mux.Handle("POST /upload", handleUpload(logger, jobStore, runner, dataDir))
	mux.Handle("GET /jobs/{id}", handleStatus(logger, jobStore))
	mux.Handle("GET /jobs/{id}/out", handleDownload(logger, jobStore, dataDir))

	return mux
}
