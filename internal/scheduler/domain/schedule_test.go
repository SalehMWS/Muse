package domain_test

import (
	"testing"
	"time"

	"github.com/SalehMWS/Muse/internal/scheduler/domain"
)

func TestSchedule_IsRecurring(t *testing.T) {
	cron := "0 12 * * *"
	if !(domain.Schedule{CronExpression: &cron}).IsRecurring() {
		t.Fatal("IsRecurring() = false, want true for a cron schedule")
	}
	if (domain.Schedule{}).IsRecurring() {
		t.Fatal("IsRecurring() = true, want false for a one-time schedule")
	}
	empty := ""
	if (domain.Schedule{CronExpression: &empty}).IsRecurring() {
		t.Fatal("IsRecurring() = true, want false for empty cron")
	}
}

func TestSchedule_CanRetry(t *testing.T) {
	if !(domain.Schedule{RetryCount: 0, MaxRetries: 3}).CanRetry() {
		t.Fatal("CanRetry() = false, want true")
	}
	if (domain.Schedule{RetryCount: 3, MaxRetries: 3}).CanRetry() {
		t.Fatal("CanRetry() = true, want false when exhausted")
	}
}

func TestRetryBackoff(t *testing.T) {
	if domain.RetryBackoff(1) != time.Minute {
		t.Fatalf("RetryBackoff(1) = %v, want 1m", domain.RetryBackoff(1))
	}
	if domain.RetryBackoff(2) != 5*time.Minute {
		t.Fatalf("RetryBackoff(2) = %v, want 5m", domain.RetryBackoff(2))
	}
	if domain.RetryBackoff(9) != 15*time.Minute {
		t.Fatalf("RetryBackoff(9) = %v, want 15m", domain.RetryBackoff(9))
	}
}
