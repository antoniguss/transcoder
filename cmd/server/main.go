package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/antoniguss/transcoder/main/internal/server"
)

func main() {

	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

}

func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	addr := fs.String("addr", ":8080", "address to listen on")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	logger := slog.New(slog.NewTextHandler(w, nil))
	jobStore := server.NewJobStore()
	server := &http.Server{
		Addr:    *addr,
		Handler: newHandler(logger, jobStore),
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", *addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("listen: %w", err)
	case <-ctx.Done():
	}

	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}
