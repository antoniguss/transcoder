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
	"path/filepath"
	"time"

	"github.com/antoniguss/transcoder/main/internal/server"
)

func main() {

	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

}

func run(ctx context.Context, w io.Writer, args []string) error {
	logger := slog.New(slog.NewTextHandler(w, nil))

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	addr := fs.String("addr", ":8080", "address to listen on")
	dataDir := fs.String("data-dir", "data", "directory for uploads and outputs")
	nWorkers := fs.Int("workers", 5, "number of workers to handle transcoding jobs")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	absDir, err := filepath.Abs(*dataDir)
	if err != nil {
		return fmt.Errorf("resolving data dir: %w", err)
	}

	if err := os.MkdirAll(absDir, 0o750); err != nil {
		return fmt.Errorf("creating data dir: %w", err)
	}

	logger.Info("using data dir", "path", absDir)

	jobStore := server.NewJobStore()
	runner := server.NewRunner(ctx, jobStore, logger, *nWorkers)
	srv := &http.Server{
		Addr:              *addr,
		Handler:           server.NewHandler(logger, jobStore, runner, absDir),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", *addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("listen: %w", err)
	case <-ctx.Done():
	}

	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server shutdown", "err", err)
	}

	runner.Wait()

	return nil
}
