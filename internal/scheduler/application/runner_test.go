package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/scheduler/application"
	"github.com/SalehMWS/Muse/internal/scheduler/domain"
)

func oneTimeSchedule() domain.Schedule {
	return domain.Schedule{
		ID: uuid.New(), UserID: uuid.New(), ContentID: uuid.New(), InstagramAccountID: uuid.New(),
		Status: domain.StatusPublishing, MaxRetries: 3,
	}
}

func TestRunner_OneTimeSuccess(t *testing.T) {
	repo := newFakeScheduleRepository()
	schedule := oneTimeSchedule()
	repo.dueQueue = [][]domain.Schedule{{schedule}}
	publisher := &fakePublisher{}

	runner := application.NewRunner(repo, publisher, fakeCronParser{}, nil, time.Second, 10)
	runner.Tick(context.Background())

	if publisher.calls != 1 {
		t.Fatalf("publisher calls = %d, want 1", publisher.calls)
	}
	if len(repo.queued) != 1 || repo.queued[0] != schedule.ID {
		t.Fatalf("published = %v, want [%v]", repo.queued, schedule.ID)
	}
}

func TestRunner_RecurringReschedules(t *testing.T) {
	repo := newFakeScheduleRepository()
	cron := "0 12 * * *"
	schedule := oneTimeSchedule()
	schedule.CronExpression = &cron
	repo.dueQueue = [][]domain.Schedule{{schedule}}

	next := time.Now().Add(24 * time.Hour).UTC()
	runner := application.NewRunner(repo, &fakePublisher{}, fakeCronParser{next: next}, nil, time.Second, 10)
	runner.Tick(context.Background())

	got, ok := repo.rescheduled[schedule.ID]
	if !ok || !got.Equal(next) {
		t.Fatalf("rescheduled[%v] = %v, want %v", schedule.ID, got, next)
	}
	if len(repo.queued) != 0 {
		t.Fatal("recurring schedule should not be marked published")
	}
}

func TestRunner_FailureRetries(t *testing.T) {
	repo := newFakeScheduleRepository()
	schedule := oneTimeSchedule()
	schedule.RetryCount = 0
	repo.dueQueue = [][]domain.Schedule{{schedule}}

	runner := application.NewRunner(repo, &fakePublisher{err: context.DeadlineExceeded}, fakeCronParser{}, nil, time.Second, 10)
	runner.Tick(context.Background())

	call, ok := repo.retried[schedule.ID]
	if !ok || call.count != 1 {
		t.Fatalf("retried[%v] = %+v, want attempt 1", schedule.ID, call)
	}
	if _, failed := repo.failed[schedule.ID]; failed {
		t.Fatal("schedule should be retried, not failed, while retries remain")
	}
}

func TestRunner_FailureExhaustsToFailed(t *testing.T) {
	repo := newFakeScheduleRepository()
	schedule := oneTimeSchedule()
	schedule.RetryCount = 3
	schedule.MaxRetries = 3
	repo.dueQueue = [][]domain.Schedule{{schedule}}

	runner := application.NewRunner(repo, &fakePublisher{err: context.DeadlineExceeded}, fakeCronParser{}, nil, time.Second, 10)
	runner.Tick(context.Background())

	if _, ok := repo.failed[schedule.ID]; !ok {
		t.Fatal("schedule should be marked failed when retries are exhausted")
	}
	if _, retried := repo.retried[schedule.ID]; retried {
		t.Fatal("schedule should not be retried when exhausted")
	}
}
