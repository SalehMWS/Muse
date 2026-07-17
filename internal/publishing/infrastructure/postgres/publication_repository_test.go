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
	pubdomain "github.com/SalehMWS/Muse/internal/publishing/domain"
	"github.com/SalehMWS/Muse/internal/publishing/infrastructure/postgres"
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
	email, _ := authdomain.NewEmail("publisher@example.com")
	if _, err := authpg.NewUserRepository(tx).Create(ctx, authdomain.User{
		ID: userID, Email: email, PasswordHash: "hash", DisplayName: "Pub", Status: authdomain.StatusActive,
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

func TestPublicationRepository_CreateAndList(t *testing.T) {
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
	repo := postgres.NewPublicationRepository(tx)

	postID := "17895695668004550"
	now := time.Now().UTC()
	created, err := repo.Create(ctx, pubdomain.Publication{
		ID: mustV7(t), UserID: userID, ContentID: contentID, InstagramAccountID: accountID,
		Platform: "instagram", PlatformPostID: &postID, Status: pubdomain.StatusPublished,
		ResponseJSON: []byte(`{"media_id":"17895695668004550"}`), PublishedAt: &now,
	})
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if created.Status != pubdomain.StatusPublished || created.PlatformPostID == nil {
		t.Fatalf("Create() = %+v, unexpected", created)
	}
	if created.PublishedAt == nil {
		t.Fatal("Create() PublishedAt = nil, want set")
	}

	list, err := repo.ListByContentForUser(ctx, userID, contentID)
	if err != nil {
		t.Fatalf("ListByContentForUser() unexpected error: %v", err)
	}
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("ListByContentForUser() = %+v, want the created publication", list)
	}

	other, err := repo.ListByContentForUser(ctx, uuid.New(), contentID)
	if err != nil {
		t.Fatalf("ListByContentForUser() unexpected error: %v", err)
	}
	if len(other) != 0 {
		t.Fatalf("ListByContentForUser() cross-tenant = %d, want 0", len(other))
	}
}
