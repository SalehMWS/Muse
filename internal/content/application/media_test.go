package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/content/application"
	"github.com/SalehMWS/Muse/internal/content/domain"
)

func TestAttachMediaUseCase_Execute(t *testing.T) {
	repo := newFakeContentRepository()
	mediaRepo := newFakeMediaRepository()
	userID := uuid.New()
	created, _ := application.NewCreateUseCase(repo).Execute(context.Background(), userID, domain.NewContentInput{Title: "Post"})

	uc := application.NewAttachMediaUseCase(repo, mediaRepo)

	media, err := uc.Execute(context.Background(), userID, created.ID, "https://cdn.example.com/a.jpg", "image", 0)
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if media.URL != "https://cdn.example.com/a.jpg" || media.MediaType != domain.MediaImage {
		t.Fatalf("Execute() = %+v, unexpected", media)
	}

	t.Run("cross-tenant content is not found", func(t *testing.T) {
		if _, err := uc.Execute(context.Background(), uuid.New(), created.ID, "https://x/y.jpg", "image", 0); !errors.Is(err, application.ErrContentNotFound) {
			t.Fatalf("Execute() error = %v, want %v", err, application.ErrContentNotFound)
		}
	})

	t.Run("empty url rejected", func(t *testing.T) {
		if _, err := uc.Execute(context.Background(), userID, created.ID, "  ", "image", 0); !errors.Is(err, domain.ErrMediaURLRequired) {
			t.Fatalf("Execute() error = %v, want %v", err, domain.ErrMediaURLRequired)
		}
	})
}

func TestDeleteMediaUseCase_Execute(t *testing.T) {
	repo := newFakeContentRepository()
	mediaRepo := newFakeMediaRepository()
	userID := uuid.New()
	created, _ := application.NewCreateUseCase(repo).Execute(context.Background(), userID, domain.NewContentInput{Title: "Post"})
	media, _ := application.NewAttachMediaUseCase(repo, mediaRepo).Execute(context.Background(), userID, created.ID, "https://x/y.jpg", "image", 0)

	uc := application.NewDeleteMediaUseCase(repo, mediaRepo)
	if err := uc.Execute(context.Background(), userID, created.ID, media.ID); err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if len(mediaRepo.byID) != 0 {
		t.Fatal("Execute() did not delete media")
	}

	if err := uc.Execute(context.Background(), userID, created.ID, uuid.New()); !errors.Is(err, application.ErrMediaNotFound) {
		t.Fatalf("Execute() error = %v, want %v", err, application.ErrMediaNotFound)
	}
}
