package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	http.ListenAndServe(":8080", mux)

	// ctx := context.Background()
	// if err := run(ctx, os.Stdout, os.Args); err != nil {
	// 	fmt.Fprintf(os.Stderr, "%s\n", err)
	// 	os.Exit(1)
	// }

}

func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	return nil
}
