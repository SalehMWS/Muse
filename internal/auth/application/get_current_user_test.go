package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/auth/application"
	"github.com/SalehMWS/Muse/internal/auth/domain"
)

func TestGetCurrentUserUseCase_Execute(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		users := newFakeUserRepository()
		userID, _ := uuid.NewV7()
		users.put(domain.User{ID: userID, Email: mustEmail(t), Status: domain.StatusActive})
		uc := application.NewGetCurrentUserUseCase(users)

		got, err := uc.Execute(context.Background(), userID)
		if err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}
		if got.ID != userID {
			t.Fatalf("Execute() ID = %v, want %v", got.ID, userID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		uc := application.NewGetCurrentUserUseCase(newFakeUserRepository())
		missingID, _ := uuid.NewV7()

		_, err := uc.Execute(context.Background(), missingID)
		if !errors.Is(err, application.ErrUserNotFound) {
			t.Fatalf("Execute() error = %v, want %v", err, application.ErrUserNotFound)
		}
	})
}
