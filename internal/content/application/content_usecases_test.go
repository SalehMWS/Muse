package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/content/application"
	"github.com/SalehMWS/Muse/internal/content/domain"
)

func TestCreateUseCase_Execute(t *testing.T) {
	repo := newFakeContentRepository()
	uc := application.NewCreateUseCase(repo)
	userID := uuid.New()

	content, err := uc.Execute(context.Background(), userID, domain.NewContentInput{Title: "Post", Tags: []string{"a"}})
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if content.Status != domain.StatusDraft || content.UserID != userID {
		t.Fatalf("Execute() = %+v, unexpected", content)
	}
	if len(repo.byID) != 1 {
		t.Fatalf("repo contents = %d, want 1", len(repo.byID))
	}

	if _, err := uc.Execute(context.Background(), userID, domain.NewContentInput{Title: strings.Repeat("x", domain.MaxTitleLength+1)}); !errors.Is(err, domain.ErrTitleTooLong) {
		t.Fatalf("Execute() error = %v, want %v", err, domain.ErrTitleTooLong)
	}
	if len(repo.byID) != 1 {
		t.Fatal("invalid content was persisted")
	}
}

func TestUpdateUseCase_Execute(t *testing.T) {
	repo := newFakeContentRepository()
	userID := uuid.New()
	created, _ := application.NewCreateUseCase(repo).Execute(context.Background(), userID, domain.NewContentInput{Title: "Old"})
	uc := application.NewUpdateUseCase(repo)

	newTitle := "New"
	updated, err := uc.Execute(context.Background(), userID, created.ID, domain.UpdateContentInput{Title: &newTitle})
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if updated.Title != "New" {
		t.Fatalf("Execute() title = %q, want New", updated.Title)
	}

	if _, err := uc.Execute(context.Background(), uuid.New(), created.ID, domain.UpdateContentInput{Title: &newTitle}); !errors.Is(err, application.ErrContentNotFound) {
		t.Fatalf("Execute() cross-tenant error = %v, want %v", err, application.ErrContentNotFound)
	}
}

func TestArchiveUseCase_Execute(t *testing.T) {
	repo := newFakeContentRepository()
	userID := uuid.New()
	created, _ := application.NewCreateUseCase(repo).Execute(context.Background(), userID, domain.NewContentInput{Title: "Post"})

	archived, err := application.NewArchiveUseCase(repo).Execute(context.Background(), userID, created.ID)
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if archived.Status != domain.StatusArchived {
		t.Fatalf("Execute() status = %q, want archived", archived.Status)
	}
}

func TestDuplicateUseCase_Execute(t *testing.T) {
	repo := newFakeContentRepository()
	userID := uuid.New()
	created, _ := application.NewCreateUseCase(repo).Execute(context.Background(), userID, domain.NewContentInput{Title: "Post"})

	dup, err := application.NewDuplicateUseCase(repo).Execute(context.Background(), userID, created.ID)
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if dup.ID == created.ID || dup.Title != "Copy of Post" {
		t.Fatalf("Execute() = %+v, unexpected", dup)
	}
	if len(repo.byID) != 2 {
		t.Fatalf("repo contents = %d, want 2", len(repo.byID))
	}
}

func TestGetUseCase_Execute(t *testing.T) {
	repo := newFakeContentRepository()
	uc := application.NewGetUseCase(repo)
	if _, err := uc.Execute(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, application.ErrContentNotFound) {
		t.Fatalf("Execute() error = %v, want %v", err, application.ErrContentNotFound)
	}
}

func TestListUseCase_Execute(t *testing.T) {
	t.Run("no next cursor when under limit", func(t *testing.T) {
		repo := newFakeContentRepository()
		userID := uuid.New()
		repo.listReturn = []domain.Content{{ID: uuid.New(), UserID: userID}}
		out, err := application.NewListUseCase(repo).Execute(context.Background(), application.ListInput{UserID: userID, Limit: 20})
		if err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}
		if out.NextCursor != "" {
			t.Fatalf("NextCursor = %q, want empty", out.NextCursor)
		}
	})

	t.Run("next cursor when page full", func(t *testing.T) {
		repo := newFakeContentRepository()
		userID := uuid.New()
		page := make([]domain.Content, 2)
		for i := range page {
			page[i] = domain.Content{ID: uuid.New(), UserID: userID}
		}
		repo.listReturn = page
		out, err := application.NewListUseCase(repo).Execute(context.Background(), application.ListInput{UserID: userID, Limit: 2})
		if err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}
		if out.NextCursor == "" {
			t.Fatal("NextCursor empty, want a cursor for a full page")
		}
	})

	t.Run("filters forwarded", func(t *testing.T) {
		repo := newFakeContentRepository()
		status := "draft"
		out := application.ListInput{UserID: uuid.New(), Status: &status, Limit: 5}
		if _, err := application.NewListUseCase(repo).Execute(context.Background(), out); err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}
		if repo.lastFilter.Status == nil || *repo.lastFilter.Status != "draft" {
			t.Fatalf("filter status = %v, want draft", repo.lastFilter.Status)
		}
		if repo.lastFilter.Limit != 5 {
			t.Fatalf("filter limit = %d, want 5", repo.lastFilter.Limit)
		}
	})

	t.Run("invalid cursor", func(t *testing.T) {
		repo := newFakeContentRepository()
		_, err := application.NewListUseCase(repo).Execute(context.Background(), application.ListInput{UserID: uuid.New(), Cursor: "!!!not-valid!!!"})
		if !errors.Is(err, application.ErrInvalidCursor) {
			t.Fatalf("Execute() error = %v, want %v", err, application.ErrInvalidCursor)
		}
	})
}
