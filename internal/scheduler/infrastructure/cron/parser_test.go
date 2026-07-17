package cron_test

import (
	"testing"
	"time"

	"github.com/SalehMWS/Muse/internal/scheduler/infrastructure/cron"
)

func TestParser_Validate(t *testing.T) {
	p := cron.New()
	if err := p.Validate("0 12 * * *"); err != nil {
		t.Fatalf("Validate() valid expr error: %v", err)
	}
	if err := p.Validate("not a cron"); err == nil {
		t.Fatal("Validate() invalid expr = nil error, want error")
	}
}

func TestParser_NextUTC(t *testing.T) {
	p := cron.New()
	after := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

	next, err := p.Next("0 12 * * *", "UTC", after)
	if err != nil {
		t.Fatalf("Next() unexpected error: %v", err)
	}
	want := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("Next() = %v, want %v", next, want)
	}
}

func TestParser_NextTimezone(t *testing.T) {
	p := cron.New()
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skipf("tz database unavailable: %v", err)
	}

	after := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	next, err := p.Next("0 9 * * *", "America/New_York", after)
	if err != nil {
		t.Fatalf("Next() unexpected error: %v", err)
	}
	if next.Location() != time.UTC {
		t.Fatalf("Next() location = %v, want UTC", next.Location())
	}
	if h := next.In(loc).Hour(); h != 9 {
		t.Fatalf("Next() local hour = %d, want 9 in New York", h)
	}
}

func TestParser_InvalidTimezone(t *testing.T) {
	p := cron.New()
	if _, err := p.Next("0 12 * * *", "Nowhere/Nowhere", time.Now()); err == nil {
		t.Fatal("Next() invalid tz = nil error, want error")
	}
}
