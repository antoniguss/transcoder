package server

import "crypto/rand"


type JobState string const (
	StateSubmitted   JobState = "submitted"
	StateTranscoding JobState = "transcoding"
	StateDone        JobState = "done"
)

type JobID uint64

type Job struct {
	Id    JobID
	State JobState
}

type JobStore map[JobID]Job

func (store JobStore) GetByID(id JobID) (val Job, has bool) {
	val, has = store[id]
	return
}

func (store JobStore) SetState(id JobID, state JobState) (set bool) {
	job, has := store[id]
	if !has {
		return false
	}

	job.State = state
	return true
}

func (store JobStore) getNewID() JobID {

	id := JobID(rand.Uint64())
	if _, has := store.GetByID(id); has {
		return store.getNewID()
	}

	return id

}

func (store JobStore) AddJob() JobID {
	id := store.getNewID()
	job := Job{Id: id, State: StateSubmitted}

	store[id] = job
	return id
}

func NewJobStore() *JobStore {
	jobStore := make(JobStore)
	return &jobStore
}
