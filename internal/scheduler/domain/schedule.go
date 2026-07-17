package domain

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusScheduled  Status = "scheduled"
	StatusPublishing Status = "publishing"
	StatusPublished  Status = "published"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
)

type Schedule struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	ContentID          uuid.UUID
	InstagramAccountID uuid.UUID
	ScheduledFor       time.Time
	Timezone           string
	CronExpression     *string
	MediaType          *string
	Status             Status
	RetryCount         int
	MaxRetries         int
	NextRetryAt        *time.Time
	LastError          *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (s Schedule) IsRecurring() bool {
	return s.CronExpression != nil && *s.CronExpression != ""
}

func (s Schedule) CanRetry() bool {
	return s.RetryCount+1 <= s.MaxRetries
}

func RetryBackoff(attempt int) time.Duration {
	switch {
	case attempt <= 1:
		return time.Minute
	case attempt == 2:
		return 5 * time.Minute
	default:
		return 15 * time.Minute
	}
}
