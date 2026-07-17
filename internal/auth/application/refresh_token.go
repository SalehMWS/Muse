package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/SalehMWS/Muse/internal/auth/domain"
)

type RefreshInput struct {
	RefreshToken string
}

type RefreshOutput struct {
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}

type RefreshUseCase struct {
	sessions        SessionRepository
	issuer          TokenIssuer
	refreshTokenTTL time.Duration
}

func NewRefreshUseCase(sessions SessionRepository, issuer TokenIssuer, refreshTokenTTL time.Duration) *RefreshUseCase {
	return &RefreshUseCase{sessions: sessions, issuer: issuer, refreshTokenTTL: refreshTokenTTL}
}

func (uc *RefreshUseCase) Execute(ctx context.Context, input RefreshInput) (RefreshOutput, error) {
	hash := domain.HashRefreshToken(input.RefreshToken)

	session, err := uc.sessions.FindByRefreshTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return RefreshOutput{}, ErrSessionNotFound
		}
		return RefreshOutput{}, fmt.Errorf("refresh: find session: %w", err)
	}

	if session.IsExpired(time.Now().UTC()) {
		return RefreshOutput{}, ErrRefreshTokenExpired
	}

	rawRefreshToken, err := domain.GenerateRefreshToken()
	if err != nil {
		return RefreshOutput{}, fmt.Errorf("refresh: generate refresh token: %w", err)
	}

	newExpiresAt := time.Now().UTC().Add(uc.refreshTokenTTL)

	rotated, err := uc.sessions.Rotate(ctx, session.ID, domain.HashRefreshToken(rawRefreshToken), newExpiresAt)
	if err != nil {
		return RefreshOutput{}, fmt.Errorf("refresh: rotate session: %w", err)
	}

	accessToken, err := uc.issuer.Issue(ctx, rotated.UserID, rotated.ID)
	if err != nil {
		return RefreshOutput{}, fmt.Errorf("refresh: issue access token: %w", err)
	}

	return RefreshOutput{
		AccessToken:           accessToken.Value,
		AccessTokenExpiresAt:  accessToken.ExpiresAt,
		RefreshToken:          rawRefreshToken,
		RefreshTokenExpiresAt: rotated.ExpiresAt,
	}, nil
}
