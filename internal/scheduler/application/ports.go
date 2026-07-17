package application

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PublishCommand struct {
	UserID             uuid.UUID
	ContentID          uuid.UUID
	InstagramAccountID uuid.UUID
	MediaType          string
}

type Publisher interface {
	Publish(ctx context.Context, cmd PublishCommand) error
}

type CronParser interface {
	Validate(expression string) error
	Next(expression, timezone string, after time.Time) (time.Time, error)
}

type ContentChecker interface {
	EnsureOwned(ctx context.Context, userID, contentID uuid.UUID) error
}
