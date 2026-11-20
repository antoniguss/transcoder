package server

import (
	"fmt"
	"net/http"
)

type StatusResponse struct {
	State    string  `json:"state"`
	Progress float32 `json:"progress"`
}

func handleHealths() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
}
