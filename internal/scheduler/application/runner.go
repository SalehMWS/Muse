package application

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/SalehMWS/Muse/internal/scheduler/domain"
)

const (
	defaultInterval  = 10 * time.Second
	defaultBatchSize = 20
)

type Runner struct {
	repo      ScheduleRepository
	publisher Publisher
	cron      CronParser
	logger    *zap.Logger
	interval  time.Duration
	batchSize int32
}

func NewRunner(repo ScheduleRepository, publisher Publisher, cron CronParser, logger *zap.Logger, interval time.Duration, batchSize int32) *Runner {
	if logger == nil {
		logger = zap.NewNop()
	}
	if interval <= 0 {
		interval = defaultInterval
	}
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}
	return &Runner{
		repo:      repo,
		publisher: publisher,
		cron:      cron,
		logger:    logger,
		interval:  interval,
		batchSize: batchSize,
	}
}

func (r *Runner) Run(ctx context.Context) {
	r.logger.Info("scheduler runner started", zap.Duration("interval", r.interval))
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("scheduler runner stopped")
			return
		case <-ticker.C:
			r.Tick(ctx)
		}
	}
}

func (r *Runner) Tick(ctx context.Context) {
	due, err := r.repo.ClaimDue(ctx, time.Now(), r.batchSize)
	if err != nil {
		r.logger.Error("scheduler: claim due schedules", zap.Error(err))
		return
	}
	for _, schedule := range due {
		r.process(ctx, schedule)
	}
}

func (r *Runner) process(ctx context.Context, schedule domain.Schedule) {
	cmd := PublishCommand{
		UserID:             schedule.UserID,
		ContentID:          schedule.ContentID,
		InstagramAccountID: schedule.InstagramAccountID,
		MediaType:          derefString(schedule.MediaType),
	}

	if err := r.publisher.Publish(ctx, cmd); err != nil {
		r.handleFailure(ctx, schedule, err)
		return
	}

	if schedule.IsRecurring() {
		next, err := r.cron.Next(*schedule.CronExpression, schedule.Timezone, time.Now())
		if err != nil {
			r.logger.Error("scheduler: compute next cron run", zap.String("schedule_id", schedule.ID.String()), zap.Error(err))
			_ = r.repo.MarkFailed(ctx, schedule.ID, "compute next cron run: "+err.Error())
			return
		}
		if err := r.repo.Reschedule(ctx, schedule.ID, next); err != nil {
			r.logger.Error("scheduler: reschedule recurring", zap.Error(err))
		}
		return
	}

	if err := r.repo.MarkPublished(ctx, schedule.ID); err != nil {
		r.logger.Error("scheduler: mark published", zap.Error(err))
	}
}

func (r *Runner) handleFailure(ctx context.Context, schedule domain.Schedule, cause error) {
	r.logger.Warn("scheduler: scheduled publish failed",
		zap.String("schedule_id", schedule.ID.String()),
		zap.Int("retry_count", schedule.RetryCount),
		zap.Error(cause),
	)

	if !schedule.CanRetry() {
		if err := r.repo.MarkFailed(ctx, schedule.ID, cause.Error()); err != nil {
			r.logger.Error("scheduler: mark failed", zap.Error(err))
		}
		return
	}

	attempt := schedule.RetryCount + 1
	next := time.Now().Add(domain.RetryBackoff(attempt))
	if err := r.repo.Retry(ctx, schedule.ID, attempt, next, cause.Error()); err != nil {
		r.logger.Error("scheduler: schedule retry", zap.Error(err))
	}
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
