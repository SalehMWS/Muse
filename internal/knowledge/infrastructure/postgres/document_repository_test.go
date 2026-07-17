package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authdomain "github.com/SalehMWS/Muse/internal/auth/domain"
	authpg "github.com/SalehMWS/Muse/internal/auth/infrastructure/postgres"
	"github.com/SalehMWS/Muse/internal/knowledge/application"
	"github.com/SalehMWS/Muse/internal/knowledge/domain"
	"github.com/SalehMWS/Muse/internal/knowledge/infrastructure/postgres"
)

func seedUser(t *testing.T, ctx context.Context, tx pgx.Tx, email string) uuid.UUID {
	t.Helper()
	userID, _ := uuid.NewV7()
	parsed, _ := authdomain.NewEmail(email)
	if _, err := authpg.NewUserRepository(tx).Create(ctx, authdomain.User{
		ID: userID, Email: parsed, PasswordHash: "hash", DisplayName: "KB", Status: authdomain.StatusActive,
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return userID
}

func TestDocumentRepository_Lifecycle(t *testing.T) {
	if pool == nil {
		t.Skip("postgres not reachable, skipping integration test")
	}

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() unexpected error: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	userID := seedUser(t, ctx, tx, "kb-owner@example.com")
	repo := postgres.NewDocumentRepository(tx)

	id, _ := uuid.NewV7()
	created, err := repo.Create(ctx, domain.Document{
		ID: id, UserID: userID, Title: "Brand", Source: "manual", Status: domain.StatusPending,
	})
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if created.Status != domain.StatusPending {
		t.Fatalf("Create() status = %q, want pending", created.Status)
	}

	updated, err := repo.UpdateStatus(ctx, id, domain.StatusIndexed, 7, nil)
	if err != nil {
		t.Fatalf("UpdateStatus() unexpected error: %v", err)
	}
	if updated.Status != domain.StatusIndexed || updated.ChunkCount != 7 {
		t.Fatalf("UpdateStatus() = %+v, unexpected", updated)
	}

	list, err := repo.ListByUser(ctx, userID)
	if err != nil {
		t.Fatalf("ListByUser() unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("ListByUser() = %d, want 1", len(list))
	}

	if _, err := repo.FindByIDForUser(ctx, id, uuid.New()); !errors.Is(err, application.ErrDocumentNotFound) {
		t.Fatalf("FindByIDForUser() cross-tenant error = %v, want %v", err, application.ErrDocumentNotFound)
	}

	if err := repo.DeleteForUser(ctx, id, userID); err != nil {
		t.Fatalf("DeleteForUser() unexpected error: %v", err)
	}
	if _, err := repo.FindByIDForUser(ctx, id, userID); !errors.Is(err, application.ErrDocumentNotFound) {
		t.Fatalf("FindByIDForUser() after delete error = %v, want %v", err, application.ErrDocumentNotFound)
	}
}
