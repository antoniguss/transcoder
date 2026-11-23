package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/antoniguss/transcoder/main/internal/transcode"
)

type Runner struct {
	store  *JobStore
	jobs   chan Job
	logger *slog.Logger
	wg     sync.WaitGroup
}

func NewRunner(ctx context.Context, store *JobStore, logger *slog.Logger, nWorkers int) *Runner {
	runner := &Runner{
		store:  store,
		jobs:   make(chan Job, 100),
		logger: logger,
	}
	runner.Run(ctx, nWorkers)
	return runner
}

func (runner *Runner) Submit(job Job) error {
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

			if err := runner.store.SetState(job.ID, StateTranscoding); err != nil {
				runner.logger.Error("changing state", "err", err, "job_id", job.ID)
			}

			onProgress := func(progress transcode.Progress) {
				if err := runner.store.SetProgress(job.ID, progress); err != nil {
					runner.logger.Error("updating progress", "err", err)
				}
			}

			if err := processTranscode(ctx, &job, onProgress); err != nil {
				runner.logger.Error("worker transcoding", "err", err)

				if err := runner.store.SetState(job.ID, StateError); err != nil {
					runner.logger.Error("changing state", "err", err, "job_id", job.ID)
				}

			}

			if err := runner.store.SetState(job.ID, StateDone); err != nil {
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

func processTranscode(ctx context.Context, job *Job, onProgress func(transcode.Progress)) error {

	if err := transcode.Run(ctx, job.Spec, onProgress); err != nil {
		return fmt.Errorf("transcoding job: %w", err)
	}

	return nil
}
