package server

import (
	"crypto/rand"
	"errors"
	"sync"
)

type JobState string

const (
	StateSubmitted   JobState = "submitted"
	StateTranscoding JobState = "transcoding"
	StateDone        JobState = "done"
)

type JobID string

type Job struct {
	ID    JobID    `json:"id"`
	State JobState `json:"state"`
}

type JobStore struct {
	mu   sync.Mutex
	jobs map[JobID]Job
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

func (store *JobStore) AddJob() Job {
	job := Job{ID: JobID(rand.Text()), State: StateSubmitted}

	store.mu.Lock()
	defer store.mu.Unlock()
	store.jobs[job.ID] = job
	return job
}
