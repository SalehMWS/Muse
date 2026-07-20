package postgres_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/audit/domain"
	"github.com/SalehMWS/Muse/internal/audit/infrastructure/postgres"
)

func newEvent(userID uuid.UUID) domain.Event {
	return domain.Event{
		ID:            uuid.New(),
		UserID:        &userID,
		Action:        domain.ActionUserLoggedIn,
		Result:        domain.ResultSuccess,
		ResourceType:  "session",
		ResourceID:    uuid.NewString(),
		IPAddress:     "203.0.113.10",
		UserAgent:     "novaflow-test",
		RequestID:     uuid.NewString(),
		CorrelationID: uuid.NewString(),
		TraceID:       uuid.NewString(),
		Metadata:      map[string]string{"source": "test"},
		CreatedAt:     time.Now().UTC(),
	}
}

func TestAppendAndListRoundTrip(t *testing.T) {
	if pool == nil {
		t.Skip("audit integration: postgres unavailable")
	}
	repo := postgres.NewEventRepository(pool)
	ctx := context.Background()
	userID := uuid.New()

	event := newEvent(userID)
	if err := repo.Append(ctx, event); err != nil {
		t.Fatalf("Append() error: %v", err)
	}

	events, err := repo.ListByUser(ctx, userID, 10)
	if err != nil {
		t.Fatalf("ListByUser() error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}

	got := events[0]
	if got.ID != event.ID {
		t.Errorf("ID = %v, want %v", got.ID, event.ID)
	}
	if got.Action != domain.ActionUserLoggedIn {
		t.Errorf("Action = %q, want %q", got.Action, domain.ActionUserLoggedIn)
	}
	if got.TraceID != event.TraceID {
		t.Errorf("TraceID = %q, want %q", got.TraceID, event.TraceID)
	}
	if got.Metadata["source"] != "test" {
		t.Errorf("Metadata[source] = %q, want %q", got.Metadata["source"], "test")
	}
}

func TestAuditLogRejectsUpdate(t *testing.T) {
	if pool == nil {
		t.Skip("audit integration: postgres unavailable")
	}
	repo := postgres.NewEventRepository(pool)
	ctx := context.Background()

	event := newEvent(uuid.New())
	if err := repo.Append(ctx, event); err != nil {
		t.Fatalf("Append() error: %v", err)
	}

	_, err := pool.Exec(ctx, "UPDATE audit_logs SET action = 'tampered' WHERE id = $1", event.ID)
	if err == nil {
		t.Fatal("UPDATE on audit_logs succeeded, want rejection")
	}
	if !strings.Contains(err.Error(), "append-only") {
		t.Errorf("error = %v, want an append-only rejection", err)
	}
}

func TestAuditLogRejectsDelete(t *testing.T) {
	if pool == nil {
		t.Skip("audit integration: postgres unavailable")
	}
	repo := postgres.NewEventRepository(pool)
	ctx := context.Background()

	event := newEvent(uuid.New())
	if err := repo.Append(ctx, event); err != nil {
		t.Fatalf("Append() error: %v", err)
	}

	_, err := pool.Exec(ctx, "DELETE FROM audit_logs WHERE id = $1", event.ID)
	if err == nil {
		t.Fatal("DELETE on audit_logs succeeded, want rejection")
	}
	if !strings.Contains(err.Error(), "append-only") {
		t.Errorf("error = %v, want an append-only rejection", err)
	}
}
