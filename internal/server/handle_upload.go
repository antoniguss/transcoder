package server

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/antoniguss/transcoder/main/internal/transcode"
)

type UploadResponse struct {
	ID string `json:"id"`
}

const (
	maxUploadSize = 2 << 30  // 2 GiB
	maxMemory     = 32 << 20 // 32 mIb
)

func handleUpload(logger *slog.Logger, jobStore *JobStore, runner *Runner, dataDir string) http.Handler {
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

		id := NewJobID()

		jobDir, err := saveUpload(dataDir, id, file)
		if err != nil {
			logger.Error("saving upload", "job_id", id, "err", err)
			http.Error(w, "could not save upload", http.StatusInternalServerError)
			return
		}

		job := Job{
			ID:           id,
			State:        StateSubmitted,
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

func saveUpload(dataDir string, jobID JobID, src io.Reader) (string, error) {
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
