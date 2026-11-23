package server

import (
	"errors"
	"sync"

	"github.com/antoniguss/transcoder/main/internal/transcode"
)

type JobStore struct {
	mu   sync.Mutex
	jobs map[JobID]Job
}

func (store *JobStore) SetProgress(id JobID, progress transcode.Progress) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	job, has := store.jobs[id]

	if !has {
		return errors.New("job not found")
	}

	job.Progress = progress
	store.jobs[id] = job

	return nil
}

func NewJobStore() *JobStore {
	return &JobStore{jobs: make(map[JobID]Job)}
}

func (store *JobStore) GetByID(id JobID) (Job, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()
	val, has := store.jobs[id]
	return val, has
}

func (store *JobStore) SetState(id JobID, state JobState) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	job, has := store.jobs[id]
	if !has {
		return errors.New("job not found")
	}

	job.State = state
	store.jobs[id] = job
	return nil
}

func (store *JobStore) AddJob(job Job) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	if _, has := store.jobs[job.ID]; has {
		return errors.New("duplicate job")
	}
	store.jobs[job.ID] = job
	return nil
}
