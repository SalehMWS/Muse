package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/scheduler/application"
	"github.com/SalehMWS/Muse/internal/scheduler/domain"
)

func baseInput() application.CreateScheduleInput {
	return application.CreateScheduleInput{
		UserID:             uuid.New(),
		ContentID:          uuid.New(),
		InstagramAccountID: uuid.New(),
		Timezone:           "UTC",
	}
}

func TestCreateScheduleUseCase_OneTime(t *testing.T) {
	repo := newFakeScheduleRepository()
	uc := application.NewCreateScheduleUseCase(repo, fakeCronParser{}, fakeContentChecker{})

	future := time.Now().Add(time.Hour)
	in := baseInput()
	in.ScheduledFor = &future

	schedule, err := uc.Execute(context.Background(), in)
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if schedule.Status != domain.StatusScheduled || schedule.IsRecurring() {
		t.Fatalf("Execute() = %+v, unexpected", schedule)
	}
	if schedule.MaxRetries != 3 {
		t.Fatalf("MaxRetries = %d, want default 3", schedule.MaxRetries)
	}
}

func TestCreateScheduleUseCase_Cron(t *testing.T) {
	repo := newFakeScheduleRepository()
	next := time.Now().Add(30 * time.Minute).UTC()
	uc := application.NewCreateScheduleUseCase(repo, fakeCronParser{next: next}, fakeContentChecker{})

	in := baseInput()
	in.CronExpression = "0 12 * * *"

	schedule, err := uc.Execute(context.Background(), in)
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if !schedule.IsRecurring() {
		t.Fatal("Execute() schedule is not recurring")
	}
	if !schedule.ScheduledFor.Equal(next) {
		t.Fatalf("ScheduledFor = %v, want cron next %v", schedule.ScheduledFor, next)
	}
}

func TestCreateScheduleUseCase_Validation(t *testing.T) {
	repo := newFakeScheduleRepository()

	t.Run("past time", func(t *testing.T) {
		uc := application.NewCreateScheduleUseCase(repo, fakeCronParser{}, fakeContentChecker{})
		past := time.Now().Add(-time.Hour)
		in := baseInput()
		in.ScheduledFor = &past
		if _, err := uc.Execute(context.Background(), in); !errors.Is(err, application.ErrScheduleInPast) {
			t.Fatalf("error = %v, want %v", err, application.ErrScheduleInPast)
		}
	})

	t.Run("no time and no cron", func(t *testing.T) {
		uc := application.NewCreateScheduleUseCase(repo, fakeCronParser{}, fakeContentChecker{})
		if _, err := uc.Execute(context.Background(), baseInput()); !errors.Is(err, application.ErrScheduleTimeRequired) {
			t.Fatalf("error = %v, want %v", err, application.ErrScheduleTimeRequired)
		}
	})

	t.Run("invalid timezone", func(t *testing.T) {
		uc := application.NewCreateScheduleUseCase(repo, fakeCronParser{}, fakeContentChecker{})
		in := baseInput()
		in.Timezone = "Mars/Phobos"
		future := time.Now().Add(time.Hour)
		in.ScheduledFor = &future
		if _, err := uc.Execute(context.Background(), in); !errors.Is(err, application.ErrInvalidTimezone) {
			t.Fatalf("error = %v, want %v", err, application.ErrInvalidTimezone)
		}
	})

	t.Run("invalid cron", func(t *testing.T) {
		uc := application.NewCreateScheduleUseCase(repo, fakeCronParser{validateErr: errors.New("bad")}, fakeContentChecker{})
		in := baseInput()
		in.CronExpression = "not cron"
		if _, err := uc.Execute(context.Background(), in); !errors.Is(err, application.ErrInvalidCron) {
			t.Fatalf("error = %v, want %v", err, application.ErrInvalidCron)
		}
	})

	t.Run("content not owned", func(t *testing.T) {
		uc := application.NewCreateScheduleUseCase(repo, fakeCronParser{}, fakeContentChecker{err: application.ErrContentNotFound})
		future := time.Now().Add(time.Hour)
		in := baseInput()
		in.ScheduledFor = &future
		if _, err := uc.Execute(context.Background(), in); !errors.Is(err, application.ErrContentNotFound) {
			t.Fatalf("error = %v, want %v", err, application.ErrContentNotFound)
		}
	})
}
