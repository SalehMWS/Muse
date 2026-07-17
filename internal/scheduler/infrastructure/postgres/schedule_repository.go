package postgres

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/SalehMWS/Muse/internal/scheduler/application"
	"github.com/SalehMWS/Muse/internal/scheduler/domain"
	"github.com/SalehMWS/Muse/internal/shared/database/sqlc"
)

type ScheduleRepository struct {
	queries *sqlc.Queries
}

func NewScheduleRepository(db sqlc.DBTX) *ScheduleRepository {
	return &ScheduleRepository{queries: sqlc.New(db)}
}

func (r *ScheduleRepository) Create(ctx context.Context, schedule domain.Schedule) (domain.Schedule, error) {
	row, err := r.queries.CreateSchedule(ctx, sqlc.CreateScheduleParams{
		ID:                 schedule.ID,
		UserID:             schedule.UserID,
		ContentID:          schedule.ContentID,
		InstagramAccountID: schedule.InstagramAccountID,
		ScheduledFor:       pgtype.Timestamptz{Time: schedule.ScheduledFor, Valid: true},
		Timezone:           schedule.Timezone,
		CronExpression:     schedule.CronExpression,
		MediaType:          schedule.MediaType,
		MaxRetries:         toInt32(schedule.MaxRetries),
	})
	if err != nil {
		return domain.Schedule{}, err
	}
	return toDomainSchedule(row), nil
}

func (r *ScheduleRepository) FindByIDForUser(ctx context.Context, id, userID uuid.UUID) (domain.Schedule, error) {
	row, err := r.queries.GetScheduleByIDForUser(ctx, sqlc.GetScheduleByIDForUserParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Schedule{}, application.ErrScheduleNotFound
		}
		return domain.Schedule{}, err
	}
	return toDomainSchedule(row), nil
}

func (r *ScheduleRepository) ListByContentForUser(ctx context.Context, userID, contentID uuid.UUID) ([]domain.Schedule, error) {
	rows, err := r.queries.ListSchedulesByContentForUser(ctx, sqlc.ListSchedulesByContentForUserParams{ContentID: contentID, UserID: userID})
	if err != nil {
		return nil, err
	}
	return mapSchedules(rows), nil
}

func (r *ScheduleRepository) ClaimDue(ctx context.Context, now time.Time, limit int32) ([]domain.Schedule, error) {
	rows, err := r.queries.ClaimDueSchedules(ctx, sqlc.ClaimDueSchedulesParams{
		ScheduledFor: pgtype.Timestamptz{Time: now, Valid: true},
		Limit:        limit,
	})
	if err != nil {
		return nil, err
	}
	return mapSchedules(rows), nil
}

func (r *ScheduleRepository) MarkQueued(ctx context.Context, id uuid.UUID) error {
	return r.queries.MarkScheduleQueued(ctx, id)
}

func (r *ScheduleRepository) MarkFailed(ctx context.Context, id uuid.UUID, reason string) error {
	return r.queries.MarkScheduleFailed(ctx, sqlc.MarkScheduleFailedParams{ID: id, LastError: &reason})
}

func (r *ScheduleRepository) Reschedule(ctx context.Context, id uuid.UUID, next time.Time) error {
	return r.queries.RescheduleSchedule(ctx, sqlc.RescheduleScheduleParams{
		ID:           id,
		ScheduledFor: pgtype.Timestamptz{Time: next, Valid: true},
	})
}

func (r *ScheduleRepository) Retry(ctx context.Context, id uuid.UUID, retryCount int, next time.Time, reason string) error {
	return r.queries.RetrySchedule(ctx, sqlc.RetryScheduleParams{
		ID:           id,
		RetryCount:   toInt32(retryCount),
		ScheduledFor: pgtype.Timestamptz{Time: next, Valid: true},
		LastError:    &reason,
	})
}

func toInt32(value int) int32 {
	if value < 0 {
		return 0
	}
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(value)
}

func (r *ScheduleRepository) Cancel(ctx context.Context, id, userID uuid.UUID) error {
	return r.queries.CancelSchedule(ctx, sqlc.CancelScheduleParams{ID: id, UserID: userID})
}

func mapSchedules(rows []sqlc.Schedule) []domain.Schedule {
	items := make([]domain.Schedule, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainSchedule(row))
	}
	return items
}

func toDomainSchedule(row sqlc.Schedule) domain.Schedule {
	schedule := domain.Schedule{
		ID:                 row.ID,
		UserID:             row.UserID,
		ContentID:          row.ContentID,
		InstagramAccountID: row.InstagramAccountID,
		ScheduledFor:       row.ScheduledFor.Time,
		Timezone:           row.Timezone,
		CronExpression:     row.CronExpression,
		MediaType:          row.MediaType,
		Status:             domain.Status(row.Status),
		RetryCount:         int(row.RetryCount),
		MaxRetries:         int(row.MaxRetries),
		LastError:          row.LastError,
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
	}
	if row.NextRetryAt.Valid {
		next := row.NextRetryAt.Time
		schedule.NextRetryAt = &next
	}
	return schedule
}
