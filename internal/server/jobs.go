package server

import (
	"crypto/rand"

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
	Progress     transcode.Progress       `json:"progress"`
	OriginalName string                   `json:"file"`
	Spec         transcode.TranscoderSpec `json:"spec"`
}
