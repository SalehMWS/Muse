package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/auth/application"
	"github.com/SalehMWS/Muse/internal/auth/domain"
)

func mustEmail(t *testing.T) domain.Email {
	t.Helper()
	email, err := domain.NewEmail("user@example.com")
	if err != nil {
		t.Fatalf("NewEmail() unexpected error: %v", err)
	}
	return email
}

func newLoginUseCase(users *fakeUserRepository, sessions *fakeSessionRepository) *application.LoginUseCase {
	return application.NewLoginUseCase(users, sessions, fakePasswordHasher{}, fakeTokenIssuer{}, 30*24*time.Hour)
}

func TestLoginUseCase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		users := newFakeUserRepository()
		userID, _ := uuid.NewV7()
		users.put(domain.User{
			ID:           userID,
			Email:        mustEmail(t),
			PasswordHash: "hashed:Str0ng!Passw0rd",
			Status:       domain.StatusActive,
		})
		sessions := newFakeSessionRepository()
		uc := newLoginUseCase(users, sessions)

		out, err := uc.Execute(context.Background(), application.LoginInput{
			Email: "user@example.com", Password: "Str0ng!Passw0rd",
		})
		if err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}
		if out.AccessToken == "" || out.RefreshToken == "" {
			t.Fatal("Execute() did not return both tokens")
		}
		if len(sessions.byHash) != 1 {
			t.Fatalf("Execute() sessions created = %d, want 1", len(sessions.byHash))
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		users := newFakeUserRepository()
		userID, _ := uuid.NewV7()
		users.put(domain.User{
			ID: userID, Email: mustEmail(t),
			PasswordHash: "hashed:Str0ng!Passw0rd", Status: domain.StatusActive,
		})
		uc := newLoginUseCase(users, newFakeSessionRepository())

		_, err := uc.Execute(context.Background(), application.LoginInput{
			Email: "user@example.com", Password: "wrong-password",
		})
		if !errors.Is(err, application.ErrInvalidCredentials) {
			t.Fatalf("Execute() error = %v, want %v", err, application.ErrInvalidCredentials)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		uc := newLoginUseCase(newFakeUserRepository(), newFakeSessionRepository())

		_, err := uc.Execute(context.Background(), application.LoginInput{
			Email: "missing@example.com", Password: "Str0ng!Passw0rd",
		})
		if !errors.Is(err, application.ErrInvalidCredentials) {
			t.Fatalf("Execute() error = %v, want %v", err, application.ErrInvalidCredentials)
		}
	})

	t.Run("suspended account", func(t *testing.T) {
		users := newFakeUserRepository()
		userID, _ := uuid.NewV7()
		users.put(domain.User{
			ID: userID, Email: mustEmail(t),
			PasswordHash: "hashed:Str0ng!Passw0rd", Status: domain.StatusSuspended,
		})
		uc := newLoginUseCase(users, newFakeSessionRepository())

		_, err := uc.Execute(context.Background(), application.LoginInput{
			Email: "user@example.com", Password: "Str0ng!Passw0rd",
		})
		if !errors.Is(err, application.ErrAccountSuspended) {
			t.Fatalf("Execute() error = %v, want %v", err, application.ErrAccountSuspended)
		}
	})

	t.Run("disabled account", func(t *testing.T) {
		users := newFakeUserRepository()
		userID, _ := uuid.NewV7()
		users.put(domain.User{
			ID: userID, Email: mustEmail(t),
			PasswordHash: "hashed:Str0ng!Passw0rd", Status: domain.StatusDisabled,
		})
		uc := newLoginUseCase(users, newFakeSessionRepository())

		_, err := uc.Execute(context.Background(), application.LoginInput{
			Email: "user@example.com", Password: "Str0ng!Passw0rd",
		})
		if !errors.Is(err, application.ErrAccountDisabled) {
			t.Fatalf("Execute() error = %v, want %v", err, application.ErrAccountDisabled)
		}
	})
}
