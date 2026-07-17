package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/auth/application"
	"github.com/SalehMWS/Muse/internal/auth/domain"
	"github.com/SalehMWS/Muse/internal/auth/infrastructure/postgres"
)

func TestSessionRepository_CreateFindRotateDelete(t *testing.T) {
	if pool == nil {
		t.Skip("postgres not reachable, skipping integration test")
	}

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() unexpected error: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	userRepo := postgres.NewUserRepository(tx)
	userID, _ := uuid.NewV7()
	email, _ := domain.NewEmail("session-owner@example.com")
	if _, err := userRepo.Create(ctx, domain.User{
		ID: userID, Email: email, PasswordHash: "hash", DisplayName: "Owner", Status: domain.StatusActive,
	}); err != nil {
		t.Fatalf("Create() user unexpected error: %v", err)
	}

	sessionRepo := postgres.NewSessionRepository(tx)
	sessionID, _ := uuid.NewV7()
	rawToken := "raw-integration-refresh-token"

	created, err := sessionRepo.Create(ctx, domain.Session{
		ID: sessionID, UserID: userID, RefreshTokenHash: domain.HashRefreshToken(rawToken),
		ExpiresAt: time.Now().Add(time.Hour).UTC(),
	})
	if err != nil {
		t.Fatalf("Create() session unexpected error: %v", err)
	}
	if created.ID != sessionID {
		t.Fatalf("Create() ID = %v, want %v", created.ID, sessionID)
	}

	found, err := sessionRepo.FindByRefreshTokenHash(ctx, domain.HashRefreshToken(rawToken))
	if err != nil {
		t.Fatalf("FindByRefreshTokenHash() unexpected error: %v", err)
	}
	if found.UserID != userID {
		t.Fatalf("FindByRefreshTokenHash() UserID = %v, want %v", found.UserID, userID)
	}

	newRawToken := "rotated-integration-refresh-token"
	newHash := domain.HashRefreshToken(newRawToken)
	rotated, err := sessionRepo.Rotate(ctx, sessionID, newHash, time.Now().Add(2*time.Hour).UTC())
	if err != nil {
		t.Fatalf("Rotate() unexpected error: %v", err)
	}
	if rotated.RefreshTokenHash != newHash {
		t.Fatalf("Rotate() RefreshTokenHash = %v, want %v", rotated.RefreshTokenHash, newHash)
	}

	if _, err := sessionRepo.FindByRefreshTokenHash(ctx, domain.HashRefreshToken(rawToken)); !errors.Is(err, application.ErrSessionNotFound) {
		t.Fatalf("FindByRefreshTokenHash() with old hash error = %v, want %v", err, application.ErrSessionNotFound)
	}

	if err := sessionRepo.DeleteByRefreshTokenHash(ctx, newHash); err != nil {
		t.Fatalf("DeleteByRefreshTokenHash() unexpected error: %v", err)
	}
	if err := sessionRepo.DeleteByRefreshTokenHash(ctx, newHash); err != nil {
		t.Fatalf("DeleteByRefreshTokenHash() expected to be idempotent, got error: %v", err)
	}
}
