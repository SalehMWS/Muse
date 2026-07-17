package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/worker/domain"
)

func TestNewJob(t *testing.T) {
	payload := domain.PublishPayload{UserID: uuid.New(), ContentID: uuid.New(), InstagramAccountID: uuid.New(), MediaType: "image"}
	job, err := domain.NewJob(domain.TypeInstagramPublish, payload, 3)
	if err != nil {
		t.Fatalf("NewJob() unexpected error: %v", err)
	}
	if job.ID == "" || job.Type != domain.TypeInstagramPublish || job.Version != domain.CurrentVersion {
		t.Fatalf("NewJob() = %+v, unexpected", job)
	}
	if job.Attempt != 0 || job.MaxAttempts != 3 {
		t.Fatalf("NewJob() attempts = %d/%d, want 0/3", job.Attempt, job.MaxAttempts)
	}

	var decoded domain.PublishPayload
	if err := json.Unmarshal(job.Payload, &decoded); err != nil {
		t.Fatalf("payload not decodable: %v", err)
	}
	if decoded.UserID != payload.UserID {
		t.Fatalf("payload round-trip mismatch")
	}
}

func TestJob_AttemptsAndNext(t *testing.T) {
	job := domain.Job{Attempt: 0, MaxAttempts: 3}
	if !job.HasAttemptsLeft() {
		t.Fatal("HasAttemptsLeft() = false at attempt 0/3, want true")
	}
	job = job.NextAttempt()
	if job.Attempt != 1 {
		t.Fatalf("NextAttempt() attempt = %d, want 1", job.Attempt)
	}

	exhausted := domain.Job{Attempt: 2, MaxAttempts: 3}
	if exhausted.HasAttemptsLeft() {
		t.Fatal("HasAttemptsLeft() = true at attempt 2/3, want false (final attempt)")
	}
}
