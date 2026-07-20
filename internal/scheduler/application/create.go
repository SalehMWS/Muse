package application

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/scheduler/domain"
	"github.com/SalehMWS/Muse/internal/shared/metrics"
)

const defaultMaxRetries = 3

type CreateScheduleUseCase struct {
	repo     ScheduleRepository
	cron     CronParser
	content  ContentChecker
	business *metrics.Business
}

func NewCreateScheduleUseCase(repo ScheduleRepository, cron CronParser, content ContentChecker, business *metrics.Business) *CreateScheduleUseCase {
	return &CreateScheduleUseCase{repo: repo, cron: cron, content: content, business: business}
}

type CreateScheduleInput struct {
	UserID             uuid.UUID
	ContentID          uuid.UUID
	InstagramAccountID uuid.UUID
	ScheduledFor       *time.Time
	CronExpression     string
	Timezone           string
	MediaType          string
	MaxRetries         int
}

func (uc *CreateScheduleUseCase) Execute(ctx context.Context, in CreateScheduleInput) (domain.Schedule, error) {
	if err := uc.content.EnsureOwned(ctx, in.UserID, in.ContentID); err != nil {
		return domain.Schedule{}, err
	}

	timezone := strings.TrimSpace(in.Timezone)
	if timezone == "" {
		timezone = "UTC"
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return domain.Schedule{}, ErrInvalidTimezone
	}

	cronExpression := strings.TrimSpace(in.CronExpression)
	var scheduledFor time.Time
	var cronPtr *string

	if cronExpression != "" {
		if err := uc.cron.Validate(cronExpression); err != nil {
			return domain.Schedule{}, ErrInvalidCron
		}
		next, err := uc.cron.Next(cronExpression, timezone, time.Now())
		if err != nil {
			return domain.Schedule{}, ErrInvalidCron
		}
		scheduledFor = next
		cronPtr = &cronExpression
	} else {
		if in.ScheduledFor == nil {
			return domain.Schedule{}, ErrScheduleTimeRequired
		}
		if !in.ScheduledFor.After(time.Now()) {
			return domain.Schedule{}, ErrScheduleInPast
		}
		scheduledFor = in.ScheduledFor.UTC()
	}

	maxRetries := in.MaxRetries
	if maxRetries <= 0 {
		maxRetries = defaultMaxRetries
	}

	var mediaPtr *string
	if mediaType := strings.TrimSpace(in.MediaType); mediaType != "" {
		mediaPtr = &mediaType
	}

	schedule := domain.Schedule{
		ID:                 uuid.New(),
		UserID:             in.UserID,
		ContentID:          in.ContentID,
		InstagramAccountID: in.InstagramAccountID,
		ScheduledFor:       scheduledFor,
		Timezone:           timezone,
		CronExpression:     cronPtr,
		MediaType:          mediaPtr,
		Status:             domain.StatusScheduled,
		MaxRetries:         maxRetries,
	}

	created, err := uc.repo.Create(ctx, schedule)
	if err != nil {
		return domain.Schedule{}, err
	}

	uc.business.Record(metrics.EventScheduleCreated)

	return created, nil
}
