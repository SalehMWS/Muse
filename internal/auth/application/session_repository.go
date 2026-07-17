package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/auth/domain"
)

type SessionRepository interface {
	Create(ctx context.Context, session domain.Session) (domain.Session, error)
	FindByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (domain.Session, error)
	Rotate(ctx context.Context, sessionID uuid.UUID, newRefreshTokenHash string, newExpiresAt time.Time) (domain.Session, error)
	DeleteByRefreshTokenHash(ctx context.Context, refreshTokenHash string) error
}
