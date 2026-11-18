package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

func newHandler(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /healthz", handleHealths())
	return mux
}

func handleHealths() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
}
