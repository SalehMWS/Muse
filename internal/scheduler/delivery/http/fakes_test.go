package http_test

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/scheduler/application"
	"github.com/SalehMWS/Muse/internal/scheduler/domain"
)

type fakeScheduleRepository struct {
	byID map[uuid.UUID]domain.Schedule
}

func newFakeScheduleRepository() *fakeScheduleRepository {
	return &fakeScheduleRepository{byID: map[uuid.UUID]domain.Schedule{}}
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

func (f *fakeScheduleRepository) ClaimDue(context.Context, time.Time, int32) ([]domain.Schedule, error) {
	return nil, nil
}

func (f *fakeScheduleRepository) MarkQueued(context.Context, uuid.UUID) error { return nil }
func (f *fakeScheduleRepository) MarkFailed(context.Context, uuid.UUID, string) error {
	return nil
}
func (f *fakeScheduleRepository) Reschedule(context.Context, uuid.UUID, time.Time) error { return nil }
func (f *fakeScheduleRepository) Retry(context.Context, uuid.UUID, int, time.Time, string) error {
	return nil
}

func (f *fakeScheduleRepository) Cancel(_ context.Context, id, _ uuid.UUID) error {
	if schedule, ok := f.byID[id]; ok {
		schedule.Status = domain.StatusCancelled
		f.byID[id] = schedule
	}
	return nil
}

type fakeCronParser struct{}

func (fakeCronParser) Validate(string) error { return nil }
func (fakeCronParser) Next(string, string, time.Time) (time.Time, error) {
	return time.Now().Add(time.Hour), nil
}

type fakeContentChecker struct {
	err error
}

func (f fakeContentChecker) EnsureOwned(context.Context, uuid.UUID, uuid.UUID) error {
	return f.err
}
