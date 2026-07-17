package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/scheduler/domain"
)

type ScheduleRepository interface {
	Create(ctx context.Context, schedule domain.Schedule) (domain.Schedule, error)
	FindByIDForUser(ctx context.Context, id, userID uuid.UUID) (domain.Schedule, error)
	ListByContentForUser(ctx context.Context, userID, contentID uuid.UUID) ([]domain.Schedule, error)
	ClaimDue(ctx context.Context, now time.Time, limit int32) ([]domain.Schedule, error)
	MarkQueued(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, reason string) error
	Reschedule(ctx context.Context, id uuid.UUID, next time.Time) error
	Retry(ctx context.Context, id uuid.UUID, retryCount int, next time.Time, reason string) error
	Cancel(ctx context.Context, id, userID uuid.UUID) error
}
