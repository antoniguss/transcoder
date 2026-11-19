package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path"

	"github.com/antoniguss/transcoder/main/internal/server"
)

func newHandler(logger *slog.Logger, jobStore *server.JobStore) http.Handler {

	mux := http.NewServeMux()
	mux.Handle("GET /healthz", handleHealths())
	mux.Handle("POST /upload", handleUpload(logger, jobStore))
	return mux
}

type UploadResponse struct {
	Id uint64 `json:"id"`
}

func handleUpload(logger *slog.Logger, jobStore *server.JobStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(32 << 2) // limit max input len

		file, header, err := r.FormFile("file")
		if err != nil {
			// TODO: what to do here?
			panic(err)
		}

		logger.Info(fmt.Sprintf("header file: %v\n", header.Filename))
		logger.Info(fmt.Sprintf("header file extension: %v\n", path.Ext(header.Filename)))
		// TODO: More robust way to verify received file is a video
		if path.Ext(header.Filename) != ".mp4" && path.Ext(header.Filename) != ".mkv" {
			panic(fmt.Errorf("invalid file extension, got file %s", header.Filename))
		}

		id := jobStore.AddJob()

		// Store file
		_ = file

		w.Header().Set("Content-Type", "application/json")
		res := UploadResponse{Id: uint64(id)}

		json.NewEncoder(w).Encode(res)

	})
}

func handleHealths() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
}
