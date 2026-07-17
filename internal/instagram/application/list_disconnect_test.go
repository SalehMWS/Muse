package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/instagram/application"
)

func TestListUseCase_Execute(t *testing.T) {
	userID := uuid.New()
	repo := newFakeAccountRepository()
	seedAccount(repo, userID)
	seedAccount(repo, uuid.New())
	uc := application.NewListUseCase(repo)

	accounts, err := uc.Execute(context.Background(), userID)
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("Execute() returned %d accounts, want 1", len(accounts))
	}
	if accounts[0].AccessToken != "" {
		t.Fatalf("Execute() leaked token: %q", accounts[0].AccessToken)
	}
}

func TestDisconnectUseCase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userID := uuid.New()
		repo := newFakeAccountRepository()
		account := seedAccount(repo, userID)
		uc := application.NewDisconnectUseCase(repo)

		if err := uc.Execute(context.Background(), userID, account.ID); err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}
		if len(repo.byID) != 0 {
			t.Fatal("Execute() did not delete the account")
		}
	})

	t.Run("missing account", func(t *testing.T) {
		uc := application.NewDisconnectUseCase(newFakeAccountRepository())
		if err := uc.Execute(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, application.ErrAccountNotFound) {
			t.Fatalf("Execute() error = %v, want %v", err, application.ErrAccountNotFound)
		}
	})
}
