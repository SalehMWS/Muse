package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/instagram/application"
	"github.com/SalehMWS/Muse/internal/instagram/domain"
)

func TestCallbackUseCase_Execute(t *testing.T) {
	t.Run("success persists encrypted token and clears it in result", func(t *testing.T) {
		userID := uuid.New()
		oauth := &fakeOAuthClient{
			exchangeToken: application.Token{AccessToken: "long-lived-token", ExpiresIn: 60 * 24 * time.Hour},
			profile:       application.Profile{UserID: "17841400000000000", Username: "brand", AccountType: "BUSINESS"},
		}
		repo := newFakeAccountRepository()
		uc := application.NewCallbackUseCase(oauth, fakeStateSigner{userID: userID}, fakeTokenCipher{}, repo)

		account, err := uc.Execute(context.Background(), application.CallbackInput{Code: "auth-code", State: "state"})
		if err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}
		if account.AccessToken != "" {
			t.Fatalf("Execute() leaked access token in result: %q", account.AccessToken)
		}
		if account.InstagramUserID != "17841400000000000" || account.Username != "brand" {
			t.Fatalf("Execute() profile not mapped: %+v", account)
		}
		if oauth.lastProfileToken != "long-lived-token" {
			t.Fatalf("FetchProfile called with %q, want long-lived token", oauth.lastProfileToken)
		}
		if len(repo.byID) != 1 {
			t.Fatalf("repo accounts = %d, want 1", len(repo.byID))
		}
		stored := repo.byID[account.ID]
		if stored.AccessToken != "enc:long-lived-token" {
			t.Fatalf("stored token = %q, want encrypted value", stored.AccessToken)
		}
		if stored.TokenExpiresAt.Before(time.Now().Add(59 * 24 * time.Hour)) {
			t.Fatalf("stored expiry too early: %v", stored.TokenExpiresAt)
		}
	})

	t.Run("invalid state", func(t *testing.T) {
		uc := application.NewCallbackUseCase(&fakeOAuthClient{}, fakeStateSigner{verifyErr: errors.New("bad")}, fakeTokenCipher{}, newFakeAccountRepository())

		if _, err := uc.Execute(context.Background(), application.CallbackInput{Code: "c", State: "s"}); !errors.Is(err, application.ErrInvalidState) {
			t.Fatalf("Execute() error = %v, want %v", err, application.ErrInvalidState)
		}
	})

	t.Run("rejects non-publishable account", func(t *testing.T) {
		repo := newFakeAccountRepository()
		oauth := &fakeOAuthClient{
			exchangeToken: application.Token{AccessToken: "t", ExpiresIn: time.Hour},
			profile:       application.Profile{UserID: "1", Username: "personal", AccountType: "PERSONAL"},
		}
		uc := application.NewCallbackUseCase(oauth, fakeStateSigner{userID: uuid.New()}, fakeTokenCipher{}, repo)

		if _, err := uc.Execute(context.Background(), application.CallbackInput{Code: "c", State: "s"}); !errors.Is(err, application.ErrAccountNotPublishable) {
			t.Fatalf("Execute() error = %v, want %v", err, application.ErrAccountNotPublishable)
		}
		if len(repo.byID) != 0 {
			t.Fatal("Execute() persisted a non-publishable account")
		}
	})

	t.Run("exchange error propagates", func(t *testing.T) {
		exErr := errors.New("exchange failed")
		oauth := &fakeOAuthClient{exchangeErr: exErr}
		uc := application.NewCallbackUseCase(oauth, fakeStateSigner{userID: uuid.New()}, fakeTokenCipher{}, newFakeAccountRepository())

		if _, err := uc.Execute(context.Background(), application.CallbackInput{Code: "c", State: "s"}); !errors.Is(err, exErr) {
			t.Fatalf("Execute() error = %v, want %v", err, exErr)
		}
	})
}

func seedAccount(repo *fakeAccountRepository, userID uuid.UUID) domain.ConnectedAccount {
	account := domain.ConnectedAccount{ //nolint:gosec
		ID:              uuid.New(),
		UserID:          userID,
		InstagramUserID: "17841400000000000",
		Username:        "brand",
		AccessToken:     "enc:old-token",
		TokenExpiresAt:  time.Now().Add(time.Hour),
		Status:          domain.AccountStatusActive,
	}
	repo.put(account)
	return account
}
