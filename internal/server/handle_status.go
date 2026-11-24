package server

import (
	"log/slog"
	"net/http"
)

type StatusResponse struct {
	State    string   `json:"state"`
	Progress *float64 `json:"progress"`
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

		var progress *float64

		switch job.State {
		case StateDone:
			*progress = 1.0
		case StateSubmitted, StateError:
			*progress = 0
		case StateTranscoding:
			frac, has := job.Progress.Fraction()
			if has {
				*progress = frac
			}
		}

		res := StatusResponse{State: string(job.State), Progress: progress}
		writeJSON(w, logger, http.StatusOK, res)

	})
}
