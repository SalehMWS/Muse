package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/content/application"
	"github.com/SalehMWS/Muse/internal/content/domain"
	"github.com/SalehMWS/Muse/internal/content/infrastructure/postgres"
)

func TestMediaRepository_CRUD(t *testing.T) {
	if pool == nil {
		t.Skip("postgres not reachable, skipping integration test")
	}

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() unexpected error: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	userID := seedUser(t, ctx, tx, "media-owner@example.com")
	contentRepo := postgres.NewContentRepository(tx)
	created, err := contentRepo.Create(ctx, newContent(t, userID, "post", nil))
	if err != nil {
		t.Fatalf("Create() content unexpected error: %v", err)
	}

	mediaRepo := postgres.NewMediaRepository(tx)

	first, err := domain.NewMedia(created.ID, "https://cdn.example.com/a.jpg", "image", 0)
	if err != nil {
		t.Fatalf("NewMedia() unexpected error: %v", err)
	}
	if _, err := mediaRepo.Create(ctx, first); err != nil {
		t.Fatalf("Create() media unexpected error: %v", err)
	}

	second, _ := domain.NewMedia(created.ID, "https://cdn.example.com/b.mp4", "video", 1)
	if _, err := mediaRepo.Create(ctx, second); err != nil {
		t.Fatalf("Create() media 2 unexpected error: %v", err)
	}

	list, err := mediaRepo.ListByContent(ctx, created.ID)
	if err != nil {
		t.Fatalf("ListByContent() unexpected error: %v", err)
	}
	if len(list) != 2 || list[0].Position != 0 || list[1].Position != 1 {
		t.Fatalf("ListByContent() = %+v, want ordered by position", list)
	}

	found, err := mediaRepo.FindByIDForContent(ctx, first.ID, created.ID)
	if err != nil {
		t.Fatalf("FindByIDForContent() unexpected error: %v", err)
	}
	if found.URL != "https://cdn.example.com/a.jpg" {
		t.Fatalf("FindByIDForContent() url = %q, unexpected", found.URL)
	}

	if err := mediaRepo.DeleteForContent(ctx, first.ID, created.ID); err != nil {
		t.Fatalf("DeleteForContent() unexpected error: %v", err)
	}
	if _, err := mediaRepo.FindByIDForContent(ctx, first.ID, created.ID); !errors.Is(err, application.ErrMediaNotFound) {
		t.Fatalf("FindByIDForContent() after delete error = %v, want %v", err, application.ErrMediaNotFound)
	}

	if _, err := mediaRepo.FindByIDForContent(ctx, uuid.New(), created.ID); !errors.Is(err, application.ErrMediaNotFound) {
		t.Fatalf("FindByIDForContent() unknown error = %v, want %v", err, application.ErrMediaNotFound)
	}
}
