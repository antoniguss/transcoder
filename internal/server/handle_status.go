package server

import (
	"log/slog"
	"net/http"
)

type StatusResponse struct {
	State    string  `json:"state"`
	Progress float32 `json:"progress"`
}

func handleStatus(logger *slog.Logger, jobStore *JobStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		id := r.PathValue("id")
		if id == "" {
			http.NotFound(w, r)
			return
		}

		job, has := jobStore.GetByID(JobID(id))
		if !has {
			http.NotFound(w, r)
			return
		}

		res := StatusResponse{State: string(job.State), Progress: 0.0}
		writeJSON(w, logger, http.StatusOK, res)

	})
}
