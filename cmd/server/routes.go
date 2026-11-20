package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/antoniguss/transcoder/main/internal/server"
	"github.com/antoniguss/transcoder/main/internal/transcode"
)

func newHandler(
	logger *slog.Logger,
	jobStore *server.JobStore,
	runner *Runner,
	dataDir string,
) http.Handler {

	mux := http.NewServeMux()
	mux.Handle("GET /healthz", handleHealths())
	mux.Handle("POST /upload", handleUpload(logger, jobStore, runner, dataDir))
	mux.Handle("GET /jobs/{id}", handleStatus(logger, jobStore))

	return mux
}

func handleStatus(logger *slog.Logger, jobStore *server.JobStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		id := r.PathValue("id")
		if id == "" {
			http.NotFound(w, r)
			return
		}

		job, has := jobStore.GetByID(server.JobID(id))
		if !has {
			http.NotFound(w, r)
			return
		}

		// TODO: return progress
		res := StatusResponse{State: string(job.State), Progress: 0.0}
		writeJSON(w, logger, http.StatusOK, res)

	})
}

func writeJSON(w http.ResponseWriter, logger *slog.Logger, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Warn("writing response", "err", err)
	}
}

type StatusResponse struct {
	State    string  `json:"state"`
	Progress float32 `json:"progress"`
}

type UploadResponse struct {
	ID string `json:"id"`
}

type Runner struct {
	store  *server.JobStore
	jobs   chan server.Job
	logger *slog.Logger
	wg     sync.WaitGroup
}

func NewRunner(ctx context.Context, store *server.JobStore, logger *slog.Logger, nWorkers int) *Runner {
	runner := &Runner{
		store:  store,
		jobs:   make(chan server.Job, 100),
		logger: logger,
	}
	runner.Run(ctx, nWorkers)
	return runner
}

func (runner *Runner) Submit(job server.Job) error {
	select {
	case runner.jobs <- job:
		return nil
	default:
		return errors.New("runner queue full")
	}
}

func (runner *Runner) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-runner.jobs:
			if !ok {
				return // channel close, exit
			}

			if err := runner.store.SetState(job.ID, server.StateTranscoding); err != nil {
				runner.logger.Error("changing state", "err", err, "job_id", job.ID)
			}

			if err := processTranscode(ctx, job); err != nil {
				runner.logger.Error("worker transcoding", "err", err)

				if err := runner.store.SetState(job.ID, server.StateError); err != nil {
					runner.logger.Error("changing state", "err", err, "job_id", job.ID)
				}

			}

			if err := runner.store.SetState(job.ID, server.StateDone); err != nil {
				runner.logger.Error("changing state", "err", err, "job_id", job.ID)
			}

		}
	}

}

func (runner *Runner) Run(ctx context.Context, nWorkers int) {

	for i := 0; i < nWorkers; i++ {
		runner.wg.Add(1)
		go func() {
			defer runner.wg.Done()
			runner.worker(ctx)
		}()

	}
}

func (runner *Runner) Wait() {
	runner.wg.Wait()
}

func processTranscode(ctx context.Context, job server.Job) error {
	// TODO: Launch transcoder. Pass

	if err := transcode.Run(ctx, job.Spec, func(p transcode.Progress) {}); err != nil {
		return fmt.Errorf("transcoding job: %w", err)
	}

	return nil
}

const (
	maxUploadSize = 2 << 30  // 2 GiB
	maxMemory     = 32 << 20 // 32 mIb
)

func handleUpload(logger *slog.Logger, jobStore *server.JobStore, runner *Runner, dataDir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

		if err := r.ParseMultipartForm(maxMemory); err != nil {
			var tooBig *http.MaxBytesError
			if errors.As(err, &tooBig) {
				http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
				return
			}

			http.Error(w, "invalid multipart form", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "error forming file", http.StatusInternalServerError)
		}

		// TODO: More robust way to verify received file is a video
		if path.Ext(header.Filename) != ".mp4" && path.Ext(header.Filename) != ".mkv" {
			http.Error(w, "invalid file extension, got file %s", http.StatusBadRequest)
		}

		id := server.NewJobID()

		jobDir, err := saveUpload(dataDir, id, file)
		if err != nil {
			logger.Error("saving upload", "job_id", id, "err", err)
			http.Error(w, "could not save upload", http.StatusInternalServerError)
			return
		}

		job := server.Job{
			ID:           id,
			State:        server.StateSubmitted,
			OriginalName: header.Filename,
			Spec: transcode.TranscoderSpec{
				Input:  filepath.Join(jobDir, "input"),
				Output: filepath.Join(jobDir, "output.mp4"),
			},
		}

		if err := jobStore.AddJob(job); err != nil {
			logger.Error("adding job", "job_id", id, "err", err)
			http.Error(w, "could not add job", http.StatusInternalServerError)
			return
		}

		if err := runner.Submit(job); err != nil {
			logger.Error("submitting job", "job_id", id, "err", err)
			http.Error(w, "error while submiting job", http.StatusInternalServerError)
		}

		w.Header().Set("Location", "/jobs/"+string(job.ID))
		res := UploadResponse{ID: string(job.ID)}
		writeJSON(w, logger, http.StatusAccepted, res)

	})
}

func saveUpload(dataDir string, jobID server.JobID, src io.Reader) (string, error) {
	jobDir := filepath.Join(dataDir, string(jobID))
	if err := os.MkdirAll(jobDir, 0o750); err != nil {
		return "", fmt.Errorf("creating job dir: %w", err)
	}

	inputPath := filepath.Join(jobDir, "input")
	dst, err := os.Create(inputPath)
	if err != nil {
		return "", fmt.Errorf("creating input file: %w", err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return "", fmt.Errorf("writing to input file: %w", err)
	}
	if err := dst.Close(); err != nil {
		return "", fmt.Errorf("closing input file: %w", err)
	}

	return jobDir, nil

}

func handleHealths() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
}
