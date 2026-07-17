package application_test

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/scheduler/application"
	"github.com/SalehMWS/Muse/internal/scheduler/domain"
)

type retryCall struct {
	count  int
	next   time.Time
	reason string
}

type fakeScheduleRepository struct {
	byID        map[uuid.UUID]domain.Schedule
	dueQueue    [][]domain.Schedule
	queued      []uuid.UUID
	failed      map[uuid.UUID]string
	retried     map[uuid.UUID]retryCall
	rescheduled map[uuid.UUID]time.Time
	cancelled   []uuid.UUID
}

func newFakeScheduleRepository() *fakeScheduleRepository {
	return &fakeScheduleRepository{
		byID:        map[uuid.UUID]domain.Schedule{},
		failed:      map[uuid.UUID]string{},
		retried:     map[uuid.UUID]retryCall{},
		rescheduled: map[uuid.UUID]time.Time{},
	}
}

func (f *fakeScheduleRepository) Create(_ context.Context, schedule domain.Schedule) (domain.Schedule, error) {
	f.byID[schedule.ID] = schedule
	return schedule, nil
}

func (f *fakeScheduleRepository) FindByIDForUser(_ context.Context, id, userID uuid.UUID) (domain.Schedule, error) {
	schedule, ok := f.byID[id]
	if !ok || schedule.UserID != userID {
		return domain.Schedule{}, application.ErrScheduleNotFound
	}
	return schedule, nil
}

func (f *fakeScheduleRepository) ListByContentForUser(_ context.Context, userID, contentID uuid.UUID) ([]domain.Schedule, error) {
	items := make([]domain.Schedule, 0)
	for _, schedule := range f.byID {
		if schedule.UserID == userID && schedule.ContentID == contentID {
			items = append(items, schedule)
		}
	}
	return items, nil
}

func (f *fakeScheduleRepository) ClaimDue(_ context.Context, _ time.Time, _ int32) ([]domain.Schedule, error) {
	if len(f.dueQueue) == 0 {
		return nil, nil
	}
	batch := f.dueQueue[0]
	f.dueQueue = f.dueQueue[1:]
	return batch, nil
}

func (f *fakeScheduleRepository) MarkQueued(_ context.Context, id uuid.UUID) error {
	f.queued = append(f.queued, id)
	return nil
}

func (f *fakeScheduleRepository) MarkFailed(_ context.Context, id uuid.UUID, reason string) error {
	f.failed[id] = reason
	return nil
}

func (f *fakeScheduleRepository) Reschedule(_ context.Context, id uuid.UUID, next time.Time) error {
	f.rescheduled[id] = next
	return nil
}

func (f *fakeScheduleRepository) Retry(_ context.Context, id uuid.UUID, retryCount int, next time.Time, reason string) error {
	f.retried[id] = retryCall{count: retryCount, next: next, reason: reason}
	return nil
}

func (f *fakeScheduleRepository) Cancel(_ context.Context, id, _ uuid.UUID) error {
	f.cancelled = append(f.cancelled, id)
	return nil
}

type fakeCronParser struct {
	next        time.Time
	validateErr error
	nextErr     error
}

func (f fakeCronParser) Validate(string) error { return f.validateErr }

func (f fakeCronParser) Next(string, string, time.Time) (time.Time, error) {
	return f.next, f.nextErr
}

type fakeContentChecker struct {
	err error
}

func (f fakeContentChecker) EnsureOwned(context.Context, uuid.UUID, uuid.UUID) error {
	return f.err
}

type fakePublisher struct {
	err   error
	calls int
}

func (f *fakePublisher) Publish(context.Context, application.PublishCommand) error {
	f.calls++
	return f.err
}
