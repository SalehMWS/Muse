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

func TestRefreshUseCase_Execute(t *testing.T) {
	t.Run("success rotates the session", func(t *testing.T) {
		sessions := newFakeSessionRepository()
		sessionID, _ := uuid.NewV7()
		userID, _ := uuid.NewV7()
		rawToken := "raw-refresh-token" //nolint:gosec
		sessions.byHash[domain.HashRefreshToken(rawToken)] = domain.Session{
			ID: sessionID, UserID: userID, RefreshTokenHash: domain.HashRefreshToken(rawToken),
			ExpiresAt: time.Now().Add(time.Hour),
		}
		uc := application.NewRefreshUseCase(sessions, fakeTokenIssuer{}, 30*24*time.Hour)

		out, err := uc.Execute(context.Background(), application.RefreshInput{RefreshToken: rawToken})
		if err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}
		if out.RefreshToken == rawToken {
			t.Fatal("Execute() did not rotate the refresh token")
		}
		if _, err := sessions.FindByRefreshTokenHash(context.Background(), domain.HashRefreshToken(rawToken)); err == nil {
			t.Fatal("Execute() old refresh token hash is still valid")
		}
		if _, err := sessions.FindByRefreshTokenHash(context.Background(), domain.HashRefreshToken(out.RefreshToken)); err != nil {
			t.Fatalf("Execute() new refresh token not persisted: %v", err)
		}
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		uc := application.NewRefreshUseCase(newFakeSessionRepository(), fakeTokenIssuer{}, 30*24*time.Hour)

		_, err := uc.Execute(context.Background(), application.RefreshInput{RefreshToken: "unknown"})
		if !errors.Is(err, application.ErrSessionNotFound) {
			t.Fatalf("Execute() error = %v, want %v", err, application.ErrSessionNotFound)
		}
	})

	t.Run("expired refresh token", func(t *testing.T) {
		sessions := newFakeSessionRepository()
		sessionID, _ := uuid.NewV7()
		rawToken := "expired-token"
		sessions.byHash[domain.HashRefreshToken(rawToken)] = domain.Session{
			ID: sessionID, RefreshTokenHash: domain.HashRefreshToken(rawToken),
			ExpiresAt: time.Now().Add(-time.Hour),
		}
		uc := application.NewRefreshUseCase(sessions, fakeTokenIssuer{}, 30*24*time.Hour)

		_, err := uc.Execute(context.Background(), application.RefreshInput{RefreshToken: rawToken})
		if !errors.Is(err, application.ErrRefreshTokenExpired) {
			t.Fatalf("Execute() error = %v, want %v", err, application.ErrRefreshTokenExpired)
		}
	})
}
