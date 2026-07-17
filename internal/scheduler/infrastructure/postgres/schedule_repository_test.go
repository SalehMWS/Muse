package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authdomain "github.com/SalehMWS/Muse/internal/auth/domain"
	authpg "github.com/SalehMWS/Muse/internal/auth/infrastructure/postgres"
	contentdomain "github.com/SalehMWS/Muse/internal/content/domain"
	contentpg "github.com/SalehMWS/Muse/internal/content/infrastructure/postgres"
	igdomain "github.com/SalehMWS/Muse/internal/instagram/domain"
	igpg "github.com/SalehMWS/Muse/internal/instagram/infrastructure/postgres"
	scheddomain "github.com/SalehMWS/Muse/internal/scheduler/domain"
	"github.com/SalehMWS/Muse/internal/scheduler/infrastructure/postgres"
)

func mustV7(t *testing.T) uuid.UUID {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("NewV7() unexpected error: %v", err)
	}
	return id
}

func seed(t *testing.T, ctx context.Context, tx pgx.Tx) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()

	userID := mustV7(t)
	email, _ := authdomain.NewEmail("scheduler@example.com")
	if _, err := authpg.NewUserRepository(tx).Create(ctx, authdomain.User{
		ID: userID, Email: email, PasswordHash: "hash", DisplayName: "Sched", Status: authdomain.StatusActive,
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	content, err := contentpg.NewContentRepository(tx).Create(ctx, contentdomain.Content{
		ID: mustV7(t), UserID: userID, Title: "post", Status: contentdomain.StatusDraft,
		Language: "en", ContentType: contentdomain.TypeImage, Visibility: contentdomain.VisibilityPrivate,
	})
	if err != nil {
		t.Fatalf("seed content: %v", err)
	}
	account, err := igpg.NewAccountRepository(tx).Upsert(ctx, igdomain.ConnectedAccount{
		ID: mustV7(t), UserID: userID, InstagramUserID: "17841400000000000", Username: "brand",
		AccessToken: "encrypted", TokenExpiresAt: time.Now().Add(time.Hour).UTC(), Status: igdomain.AccountStatusActive,
	})
	if err != nil {
		t.Fatalf("seed instagram account: %v", err)
	}
	return userID, content.ID, account.ID
}

func TestScheduleRepository_ClaimAndMark(t *testing.T) {
	if pool == nil {
		t.Skip("postgres not reachable, skipping integration test")
	}

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() unexpected error: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	userID, contentID, accountID := seed(t, ctx, tx)
	repo := postgres.NewScheduleRepository(tx)

	created, err := repo.Create(ctx, scheddomain.Schedule{
		ID: mustV7(t), UserID: userID, ContentID: contentID, InstagramAccountID: accountID,
		ScheduledFor: time.Now().Add(-time.Minute).UTC(), Timezone: "UTC",
		Status: scheddomain.StatusScheduled, MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	due, err := repo.ClaimDue(ctx, time.Now(), 10)
	if err != nil {
		t.Fatalf("ClaimDue() unexpected error: %v", err)
	}
	if len(due) != 1 || due[0].ID != created.ID {
		t.Fatalf("ClaimDue() = %+v, want the created schedule", due)
	}
	if due[0].Status != scheddomain.StatusPublishing {
		t.Fatalf("ClaimDue() status = %q, want publishing (claimed)", due[0].Status)
	}

	// A second claim must not return the already-claimed row.
	again, err := repo.ClaimDue(ctx, time.Now(), 10)
	if err != nil {
		t.Fatalf("ClaimDue() second unexpected error: %v", err)
	}
	if len(again) != 0 {
		t.Fatalf("ClaimDue() second = %d rows, want 0 (already claimed)", len(again))
	}

	if err := repo.MarkQueued(ctx, created.ID); err != nil {
		t.Fatalf("MarkQueued() unexpected error: %v", err)
	}
	found, err := repo.FindByIDForUser(ctx, created.ID, userID)
	if err != nil {
		t.Fatalf("FindByIDForUser() unexpected error: %v", err)
	}
	if found.Status != scheddomain.StatusQueued {
		t.Fatalf("status = %q, want queued", found.Status)
	}
}

func TestScheduleRepository_RetryAndCancel(t *testing.T) {
	if pool == nil {
		t.Skip("postgres not reachable, skipping integration test")
	}

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() unexpected error: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	userID, contentID, accountID := seed(t, ctx, tx)
	repo := postgres.NewScheduleRepository(tx)

	created, err := repo.Create(ctx, scheddomain.Schedule{
		ID: mustV7(t), UserID: userID, ContentID: contentID, InstagramAccountID: accountID,
		ScheduledFor: time.Now().Add(time.Hour).UTC(), Timezone: "UTC",
		Status: scheddomain.StatusScheduled, MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	next := time.Now().Add(5 * time.Minute).UTC()
	if err := repo.Retry(ctx, created.ID, 1, next, "boom"); err != nil {
		t.Fatalf("Retry() unexpected error: %v", err)
	}
	found, _ := repo.FindByIDForUser(ctx, created.ID, userID)
	if found.RetryCount != 1 || found.LastError == nil || *found.LastError != "boom" {
		t.Fatalf("Retry() state = %+v, unexpected", found)
	}

	if err := repo.Cancel(ctx, created.ID, userID); err != nil {
		t.Fatalf("Cancel() unexpected error: %v", err)
	}
	found, _ = repo.FindByIDForUser(ctx, created.ID, userID)
	if found.Status != scheddomain.StatusCancelled {
		t.Fatalf("Cancel() status = %q, want cancelled", found.Status)
	}
}
