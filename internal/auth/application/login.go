package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/auth/domain"
)

type LoginInput struct {
	Email     string
	Password  string
	Device    string
	IPAddress string
	UserAgent string
}

type LoginOutput struct {
	User                  domain.User
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}

type LoginUseCase struct {
	users           UserRepository
	sessions        SessionRepository
	hasher          PasswordHasher
	issuer          TokenIssuer
	refreshTokenTTL time.Duration
}

func NewLoginUseCase(users UserRepository, sessions SessionRepository, hasher PasswordHasher, issuer TokenIssuer, refreshTokenTTL time.Duration) *LoginUseCase {
	return &LoginUseCase{
		users:           users,
		sessions:        sessions,
		hasher:          hasher,
		issuer:          issuer,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (LoginOutput, error) {
	email, err := domain.NewEmail(input.Email)
	if err != nil {
		return LoginOutput{}, ErrInvalidCredentials
	}

	user, err := uc.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return LoginOutput{}, ErrInvalidCredentials
		}
		return LoginOutput{}, fmt.Errorf("login: find user: %w", err)
	}

	match, err := uc.hasher.Verify(user.PasswordHash, input.Password)
	if err != nil {
		return LoginOutput{}, fmt.Errorf("login: verify password: %w", err)
	}
	if !match {
		return LoginOutput{}, ErrInvalidCredentials
	}

	if !user.CanAuthenticate() {
		switch user.Status {
		case domain.StatusSuspended:
			return LoginOutput{}, ErrAccountSuspended
		default:
			return LoginOutput{}, ErrAccountDisabled
		}
	}

	sessionID, err := uuid.NewV7()
	if err != nil {
		return LoginOutput{}, fmt.Errorf("login: generate session id: %w", err)
	}

	rawRefreshToken, err := domain.GenerateRefreshToken()
	if err != nil {
		return LoginOutput{}, fmt.Errorf("login: generate refresh token: %w", err)
	}

	now := time.Now().UTC()
	refreshTokenExpiresAt := now.Add(uc.refreshTokenTTL)

	var device, ipAddress, userAgent *string
	if input.Device != "" {
		device = &input.Device
	}
	if input.IPAddress != "" {
		ipAddress = &input.IPAddress
	}
	if input.UserAgent != "" {
		userAgent = &input.UserAgent
	}

	session := domain.Session{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: domain.HashRefreshToken(rawRefreshToken),
		Device:           device,
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		ExpiresAt:        refreshTokenExpiresAt,
	}

	if _, err := uc.sessions.Create(ctx, session); err != nil {
		return LoginOutput{}, fmt.Errorf("login: create session: %w", err)
	}

	accessToken, err := uc.issuer.Issue(ctx, user.ID, sessionID)
	if err != nil {
		return LoginOutput{}, fmt.Errorf("login: issue access token: %w", err)
	}

	return LoginOutput{
		User:                  user,
		AccessToken:           accessToken.Value,
		AccessTokenExpiresAt:  accessToken.ExpiresAt,
		RefreshToken:          rawRefreshToken,
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
	}, nil
}
