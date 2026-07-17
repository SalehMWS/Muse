package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/instagram/application"
)

func TestRefreshUseCase_Execute(t *testing.T) {
	t.Run("success re-encrypts new token", func(t *testing.T) {
		userID := uuid.New()
		repo := newFakeAccountRepository()
		account := seedAccount(repo, userID)
		oauth := &fakeOAuthClient{refreshedToken: application.Token{AccessToken: "new-token", ExpiresIn: 60 * 24 * time.Hour}}
		uc := application.NewRefreshUseCase(oauth, fakeTokenCipher{}, repo)

		out, err := uc.Execute(context.Background(), userID, account.ID)
		if err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}
		if out.AccessToken != "" {
			t.Fatalf("Execute() leaked token: %q", out.AccessToken)
		}
		if oauth.lastRefreshToken != "old-token" {
			t.Fatalf("RefreshToken called with %q, want decrypted old token", oauth.lastRefreshToken)
		}
		if repo.byID[account.ID].AccessToken != "enc:new-token" {
			t.Fatalf("stored token = %q, want enc:new-token", repo.byID[account.ID].AccessToken)
		}
	})

	t.Run("account not found", func(t *testing.T) {
		uc := application.NewRefreshUseCase(&fakeOAuthClient{}, fakeTokenCipher{}, newFakeAccountRepository())
		if _, err := uc.Execute(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, application.ErrAccountNotFound) {
			t.Fatalf("Execute() error = %v, want %v", err, application.ErrAccountNotFound)
		}
	})

	t.Run("other users account is not found", func(t *testing.T) {
		repo := newFakeAccountRepository()
		account := seedAccount(repo, uuid.New())
		uc := application.NewRefreshUseCase(&fakeOAuthClient{}, fakeTokenCipher{}, repo)

		if _, err := uc.Execute(context.Background(), uuid.New(), account.ID); !errors.Is(err, application.ErrAccountNotFound) {
			t.Fatalf("Execute() error = %v, want %v", err, application.ErrAccountNotFound)
		}
	})
}
