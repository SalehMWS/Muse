package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authdomain "github.com/SalehMWS/Muse/internal/auth/domain"
	authpostgres "github.com/SalehMWS/Muse/internal/auth/infrastructure/postgres"
	"github.com/SalehMWS/Muse/internal/instagram/application"
	"github.com/SalehMWS/Muse/internal/instagram/domain"
	"github.com/SalehMWS/Muse/internal/instagram/infrastructure/postgres"
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

func TestAccountRepository_UpsertFindUpdateDelete(t *testing.T) {
	if pool == nil {
		t.Skip("postgres not reachable, skipping integration test")
	}

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() unexpected error: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	userID := seedUser(t, ctx, tx, "instagram-owner@example.com")
	repo := postgres.NewAccountRepository(tx)

	accountType := "BUSINESS"
	scopes := "instagram_business_basic,instagram_business_content_publish"
	accountID, _ := uuid.NewV7()
	created, err := repo.Upsert(ctx, domain.ConnectedAccount{
		ID:              accountID,
		UserID:          userID,
		InstagramUserID: "17841400000000000",
		Username:        "brand",
		AccountType:     &accountType,
		AccessToken:     "encrypted-token",
		TokenExpiresAt:  time.Now().Add(60 * 24 * time.Hour).UTC(),
		Scopes:          &scopes,
		Status:          domain.AccountStatusActive,
	})
	if err != nil {
		t.Fatalf("Upsert() unexpected error: %v", err)
	}
	if created.ID != accountID {
		t.Fatalf("Upsert() ID = %v, want %v", created.ID, accountID)
	}

	found, err := repo.FindByIDForUser(ctx, accountID, userID)
	if err != nil {
		t.Fatalf("FindByIDForUser() unexpected error: %v", err)
	}
	if found.Username != "brand" || found.AccessToken != "encrypted-token" {
		t.Fatalf("FindByIDForUser() = %+v, unexpected", found)
	}

	updated, err := repo.UpdateToken(ctx, accountID, "new-encrypted-token", time.Now().Add(90*24*time.Hour).UTC(), domain.AccountStatusActive)
	if err != nil {
		t.Fatalf("UpdateToken() unexpected error: %v", err)
	}
	if updated.AccessToken != "new-encrypted-token" {
		t.Fatalf("UpdateToken() token = %q, want new-encrypted-token", updated.AccessToken)
	}
	if updated.LastRefreshedAt == nil {
		t.Fatal("UpdateToken() LastRefreshedAt = nil, want set")
	}

	accounts, err := repo.ListByUser(ctx, userID)
	if err != nil {
		t.Fatalf("ListByUser() unexpected error: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("ListByUser() = %d accounts, want 1", len(accounts))
	}

	if err := repo.DeleteForUser(ctx, accountID, userID); err != nil {
		t.Fatalf("DeleteForUser() unexpected error: %v", err)
	}
	if _, err := repo.FindByIDForUser(ctx, accountID, userID); !errors.Is(err, application.ErrAccountNotFound) {
		t.Fatalf("FindByIDForUser() after delete error = %v, want %v", err, application.ErrAccountNotFound)
	}
}

func TestAccountRepository_UpsertReconnectUpdates(t *testing.T) {
	if pool == nil {
		t.Skip("postgres not reachable, skipping integration test")
	}

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() unexpected error: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	userID := seedUser(t, ctx, tx, "instagram-reconnect@example.com")
	repo := postgres.NewAccountRepository(tx)

	firstType := "BUSINESS"
	first, err := repo.Upsert(ctx, domain.ConnectedAccount{
		ID:              mustV7(t),
		UserID:          userID,
		InstagramUserID: "17841400000000000",
		Username:        "old-handle",
		AccountType:     &firstType,
		AccessToken:     "token-1",
		TokenExpiresAt:  time.Now().Add(time.Hour).UTC(),
		Status:          domain.AccountStatusActive,
	})
	if err != nil {
		t.Fatalf("Upsert() first unexpected error: %v", err)
	}

	second, err := repo.Upsert(ctx, domain.ConnectedAccount{
		ID:              mustV7(t),
		UserID:          userID,
		InstagramUserID: "17841400000000000",
		Username:        "new-handle",
		AccountType:     &firstType,
		AccessToken:     "token-2",
		TokenExpiresAt:  time.Now().Add(48 * time.Hour).UTC(),
		Status:          domain.AccountStatusActive,
	})
	if err != nil {
		t.Fatalf("Upsert() second unexpected error: %v", err)
	}

	if second.ID != first.ID {
		t.Fatalf("Upsert() reconnect created a new row: %v != %v", second.ID, first.ID)
	}
	if second.Username != "new-handle" || second.AccessToken != "token-2" {
		t.Fatalf("Upsert() reconnect did not update: %+v", second)
	}

	accounts, err := repo.ListByUser(ctx, userID)
	if err != nil {
		t.Fatalf("ListByUser() unexpected error: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("ListByUser() = %d accounts, want 1 after reconnect", len(accounts))
	}
}

func mustV7(t *testing.T) uuid.UUID {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("NewV7() unexpected error: %v", err)
	}
	return id
}
