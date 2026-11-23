package server

import (
	"log/slog"
	"net/http"
)

func handleDownload(logger *slog.Logger, jobStore *JobStore, dataDir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		job, has := jobStore.GetByID(JobID(id))
		if !has {
			http.NotFound(w, r)
			return
		}

		if job.State != StateDone {
			http.Error(w, "job not finished", http.StatusConflict)
			return
		}

		w.Header().Set("Content-Desposition", `attachment; filename="`+job.OriginalName+`"`)
		http.ServeFile(w, r, job.Spec.Output)

	})
}
