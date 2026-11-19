package server

import (
	"crypto/rand"
	"errors"
	"sync"

	"github.com/antoniguss/transcoder/main/internal/transcode"
)

type JobState string

const (
	StateSubmitted   JobState = "submitted"
	StateTranscoding JobState = "transcoding"
	StateDone        JobState = "done"
	StateError       JobState = "error"
)

type JobID string

func NewJobID() JobID {
	return JobID(rand.Text())
}

type Job struct {
	ID           JobID                    `json:"id"`
	State        JobState                 `json:"state"`
	OriginalName string                   `json:"file"`
	Spec         transcode.TranscoderSpec `json:"spec"`
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

func (store *JobStore) AddJob(job Job) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	if _, has := store.jobs[job.ID]; has {
		return errors.New("duplicate job")
	}
	store.jobs[job.ID] = job
	return nil
}
