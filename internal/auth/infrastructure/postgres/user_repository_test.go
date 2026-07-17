package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/auth/application"
	"github.com/SalehMWS/Muse/internal/auth/domain"
	"github.com/SalehMWS/Muse/internal/auth/infrastructure/postgres"
)

func TestUserRepository_CreateAndFind(t *testing.T) {
	if pool == nil {
		t.Skip("postgres not reachable, skipping integration test")
	}

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() unexpected error: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	repo := postgres.NewUserRepository(tx)

	id, _ := uuid.NewV7()
	email, _ := domain.NewEmail("integration@example.com")
	user := domain.User{
		ID: id, Email: email, PasswordHash: "hashed-value", DisplayName: "Integration Test",
		Status: domain.StatusActive, EmailVerified: false,
	}

	created, err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if created.ID != id {
		t.Fatalf("Create() ID = %v, want %v", created.ID, id)
	}

	byEmail, err := repo.FindByEmail(ctx, email)
	if err != nil {
		t.Fatalf("FindByEmail() unexpected error: %v", err)
	}
	if byEmail.ID != id {
		t.Fatalf("FindByEmail() ID = %v, want %v", byEmail.ID, id)
	}

	byID, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatalf("FindByID() unexpected error: %v", err)
	}
	if byID.Email.String() != email.String() {
		t.Fatalf("FindByID() email = %v, want %v", byID.Email.String(), email.String())
	}
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	if pool == nil {
		t.Skip("postgres not reachable, skipping integration test")
	}

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() unexpected error: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	repo := postgres.NewUserRepository(tx)
	email, _ := domain.NewEmail("duplicate@example.com")

	firstID, _ := uuid.NewV7()
	if _, err := repo.Create(ctx, domain.User{
		ID: firstID, Email: email, PasswordHash: "hash", DisplayName: "First", Status: domain.StatusActive,
	}); err != nil {
		t.Fatalf("Create() first user unexpected error: %v", err)
	}

	secondID, _ := uuid.NewV7()
	_, err = repo.Create(ctx, domain.User{
		ID: secondID, Email: email, PasswordHash: "hash", DisplayName: "Second", Status: domain.StatusActive,
	})
	if !errors.Is(err, application.ErrEmailAlreadyExists) {
		t.Fatalf("Create() error = %v, want %v", err, application.ErrEmailAlreadyExists)
	}
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	if pool == nil {
		t.Skip("postgres not reachable, skipping integration test")
	}

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() unexpected error: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	repo := postgres.NewUserRepository(tx)
	email, _ := domain.NewEmail("missing@example.com")

	_, err = repo.FindByEmail(ctx, email)
	if !errors.Is(err, application.ErrUserNotFound) {
		t.Fatalf("FindByEmail() error = %v, want %v", err, application.ErrUserNotFound)
	}
}
