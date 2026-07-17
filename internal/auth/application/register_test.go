package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/SalehMWS/Muse/internal/auth/application"
)

func TestRegisterUseCase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		uc := application.NewRegisterUseCase(newFakeUserRepository(), fakePasswordHasher{})

		out, err := uc.Execute(context.Background(), application.RegisterInput{
			Email:       "new@example.com",
			Password:    "Str0ng!Passw0rd",
			DisplayName: "New User",
		})
		if err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}
		if out.User.Email.String() != "new@example.com" {
			t.Fatalf("Execute() email = %q, want %q", out.User.Email.String(), "new@example.com")
		}
		if out.User.PasswordHash == "" || out.User.PasswordHash == "Str0ng!Passw0rd" {
			t.Fatalf("Execute() did not hash the password: %q", out.User.PasswordHash)
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		users := newFakeUserRepository()
		uc := application.NewRegisterUseCase(users, fakePasswordHasher{})

		_, err := uc.Execute(context.Background(), application.RegisterInput{
			Email: "dup@example.com", Password: "Str0ng!Passw0rd", DisplayName: "First",
		})
		if err != nil {
			t.Fatalf("Execute() first registration unexpected error: %v", err)
		}

		_, err = uc.Execute(context.Background(), application.RegisterInput{
			Email: "dup@example.com", Password: "An0ther!Passw0rd", DisplayName: "Second",
		})
		if !errors.Is(err, application.ErrEmailAlreadyExists) {
			t.Fatalf("Execute() error = %v, want %v", err, application.ErrEmailAlreadyExists)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		uc := application.NewRegisterUseCase(newFakeUserRepository(), fakePasswordHasher{})

		_, err := uc.Execute(context.Background(), application.RegisterInput{
			Email: "not-an-email", Password: "Str0ng!Passw0rd", DisplayName: "New User",
		})
		if err == nil {
			t.Fatal("Execute() expected an error for an invalid email")
		}
	})

	t.Run("weak password", func(t *testing.T) {
		uc := application.NewRegisterUseCase(newFakeUserRepository(), fakePasswordHasher{})

		_, err := uc.Execute(context.Background(), application.RegisterInput{
			Email: "weak@example.com", Password: "weak", DisplayName: "New User",
		})
		if err == nil {
			t.Fatal("Execute() expected an error for a weak password")
		}
	})
}
