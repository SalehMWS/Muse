package application

import (
	"context"
	"fmt"

	"github.com/SalehMWS/Muse/internal/auth/domain"
)

type LogoutInput struct {
	RefreshToken string
}

type LogoutUseCase struct {
	sessions SessionRepository
}

func NewLogoutUseCase(sessions SessionRepository) *LogoutUseCase {
	return &LogoutUseCase{sessions: sessions}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, input LogoutInput) error {
	hash := domain.HashRefreshToken(input.RefreshToken)

	if err := uc.sessions.DeleteByRefreshTokenHash(ctx, hash); err != nil {
		return fmt.Errorf("logout: delete session: %w", err)
	}

	return nil
}
