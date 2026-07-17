package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authdomain "github.com/SalehMWS/Muse/internal/auth/domain"
	authpostgres "github.com/SalehMWS/Muse/internal/auth/infrastructure/postgres"
	"github.com/SalehMWS/Muse/internal/content/application"
	"github.com/SalehMWS/Muse/internal/content/domain"
	"github.com/SalehMWS/Muse/internal/content/infrastructure/postgres"
)

func seedUser(t *testing.T, ctx context.Context, tx pgx.Tx, email string) uuid.UUID {
	t.Helper()
	userRepo := authpostgres.NewUserRepository(tx)
	userID, _ := uuid.NewV7()
	parsedEmail, err := authdomain.NewEmail(email)
	if err != nil {
		t.Fatalf("NewEmail() unexpected error: %v", err)
	}
	if _, err := userRepo.Create(ctx, authdomain.User{
		ID: userID, Email: parsedEmail, PasswordHash: "hash", DisplayName: "Owner", Status: authdomain.StatusActive,
	}); err != nil {
		t.Fatalf("Create() user unexpected error: %v", err)
	}
	return userID
}

func newContent(t *testing.T, userID uuid.UUID, title string, tags []string) domain.Content {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("NewV7() unexpected error: %v", err)
	}
	return domain.Content{
		ID:          id,
		UserID:      userID,
		Title:       title,
		Status:      domain.StatusDraft,
		Language:    "en",
		ContentType: domain.TypeImage,
		Visibility:  domain.VisibilityPrivate,
		Tags:        tags,
	}
}

func TestContentRepository_CreateGetUpdate(t *testing.T) {
	if pool == nil {
		t.Skip("postgres not reachable, skipping integration test")
	}

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() unexpected error: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	userID := seedUser(t, ctx, tx, "content-owner@example.com")
	repo := postgres.NewContentRepository(tx)

	created, err := repo.Create(ctx, newContent(t, userID, "Launch", []string{"promo", "launch"}))
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if len(created.Tags) != 2 {
		t.Fatalf("Create() tags = %v, want 2", created.Tags)
	}

	found, err := repo.FindByIDForUser(ctx, created.ID, userID)
	if err != nil {
		t.Fatalf("FindByIDForUser() unexpected error: %v", err)
	}
	if found.Title != "Launch" {
		t.Fatalf("FindByIDForUser() title = %q, want Launch", found.Title)
	}

	found.Title = "Relaunch"
	found.Archive()
	updated, err := repo.Update(ctx, found)
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if updated.Title != "Relaunch" || updated.Status != domain.StatusArchived {
		t.Fatalf("Update() = %+v, unexpected", updated)
	}

	if _, err := repo.FindByIDForUser(ctx, created.ID, uuid.New()); !errors.Is(err, application.ErrContentNotFound) {
		t.Fatalf("FindByIDForUser() cross-tenant error = %v, want %v", err, application.ErrContentNotFound)
	}
}

func TestContentRepository_ListKeysetPagination(t *testing.T) {
	if pool == nil {
		t.Skip("postgres not reachable, skipping integration test")
	}

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() unexpected error: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	userID := seedUser(t, ctx, tx, "content-paged@example.com")
	repo := postgres.NewContentRepository(tx)

	for _, title := range []string{"first", "second", "third"} {
		if _, err := repo.Create(ctx, newContent(t, userID, title, nil)); err != nil {
			t.Fatalf("Create(%q) unexpected error: %v", title, err)
		}
	}

	page1, err := repo.List(ctx, application.ListFilter{UserID: userID, Limit: 2})
	if err != nil {
		t.Fatalf("List() page1 unexpected error: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("List() page1 = %d items, want 2", len(page1))
	}

	last := page1[len(page1)-1]
	page2, err := repo.List(ctx, application.ListFilter{
		UserID:          userID,
		CursorCreatedAt: &last.CreatedAt,
		CursorID:        &last.ID,
		Limit:           2,
	})
	if err != nil {
		t.Fatalf("List() page2 unexpected error: %v", err)
	}
	if len(page2) != 1 {
		t.Fatalf("List() page2 = %d items, want 1", len(page2))
	}
	for _, item := range page2 {
		if item.ID == page1[0].ID || item.ID == page1[1].ID {
			t.Fatal("List() page2 overlaps page1")
		}
	}
}

func TestContentRepository_ListTagFilter(t *testing.T) {
	if pool == nil {
		t.Skip("postgres not reachable, skipping integration test")
	}

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() unexpected error: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	userID := seedUser(t, ctx, tx, "content-tags@example.com")
	repo := postgres.NewContentRepository(tx)

	if _, err := repo.Create(ctx, newContent(t, userID, "tagged", []string{"promo"})); err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if _, err := repo.Create(ctx, newContent(t, userID, "untagged", nil)); err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	promo := "promo"
	matched, err := repo.List(ctx, application.ListFilter{UserID: userID, Tag: &promo, Limit: 10})
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}
	if len(matched) != 1 || matched[0].Title != "tagged" {
		t.Fatalf("List() tag filter = %+v, want only tagged", matched)
	}

	missing := "nope"
	none, err := repo.List(ctx, application.ListFilter{UserID: userID, Tag: &missing, Limit: 10})
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("List() unknown tag = %d items, want 0", len(none))
	}
}
