package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type JobType string

const (
	TypeInstagramPublish JobType = "instagram.publish"
)

const CurrentVersion = 1

type Job struct {
	ID          string          `json:"id"`
	Type        JobType         `json:"type"`
	Version     int             `json:"version"`
	Payload     json.RawMessage `json:"payload"`
	Attempt     int             `json:"attempt"`
	MaxAttempts int             `json:"max_attempts"`
	EnqueuedAt  time.Time       `json:"enqueued_at"`
}

type PublishPayload struct {
	UserID             uuid.UUID `json:"user_id"`
	ContentID          uuid.UUID `json:"content_id"`
	InstagramAccountID uuid.UUID `json:"instagram_account_id"`
	MediaType          string    `json:"media_type,omitempty"`
}

func NewJob(jobType JobType, payload any, maxAttempts int) (Job, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return Job{}, err
	}
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	return Job{
		ID:          uuid.NewString(),
		Type:        jobType,
		Version:     CurrentVersion,
		Payload:     raw,
		Attempt:     0,
		MaxAttempts: maxAttempts,
		EnqueuedAt:  time.Now().UTC(),
	}, nil
}

func (j Job) HasAttemptsLeft() bool {
	return j.Attempt+1 < j.MaxAttempts
}

func (j Job) NextAttempt() Job {
	j.Attempt++
	return j
}
