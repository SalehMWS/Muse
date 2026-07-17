package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/auth/application"
	"github.com/SalehMWS/Muse/internal/auth/domain"
)

func TestLogoutUseCase_Execute(t *testing.T) {
	t.Run("deletes the session", func(t *testing.T) {
		sessions := newFakeSessionRepository()
		sessionID, _ := uuid.NewV7()
		rawToken := "raw-refresh-token" //nolint:gosec
		sessions.byHash[domain.HashRefreshToken(rawToken)] = domain.Session{
			ID: sessionID, RefreshTokenHash: domain.HashRefreshToken(rawToken), ExpiresAt: time.Now().Add(time.Hour),
		}
		uc := application.NewLogoutUseCase(sessions)

		if err := uc.Execute(context.Background(), application.LogoutInput{RefreshToken: rawToken}); err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}
		if len(sessions.byHash) != 0 {
			t.Fatalf("Execute() sessions remaining = %d, want 0", len(sessions.byHash))
		}
	})

	t.Run("idempotent when the session is already gone", func(t *testing.T) {
		uc := application.NewLogoutUseCase(newFakeSessionRepository())

		if err := uc.Execute(context.Background(), application.LogoutInput{RefreshToken: "unknown"}); err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}
	})
}
