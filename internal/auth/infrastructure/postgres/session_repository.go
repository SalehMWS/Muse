package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/SalehMWS/Muse/internal/auth/application"
	"github.com/SalehMWS/Muse/internal/auth/domain"
	"github.com/SalehMWS/Muse/internal/shared/database/sqlc"
)

type SessionRepository struct {
	queries *sqlc.Queries
}

func NewSessionRepository(db sqlc.DBTX) *SessionRepository {
	return &SessionRepository{queries: sqlc.New(db)}
}

func (r *SessionRepository) Create(ctx context.Context, session domain.Session) (domain.Session, error) {
	row, err := r.queries.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:               session.ID,
		UserID:           session.UserID,
		RefreshTokenHash: session.RefreshTokenHash,
		Device:           session.Device,
		IpAddress:        session.IPAddress,
		UserAgent:        session.UserAgent,
		ExpiresAt:        pgtype.Timestamptz{Time: session.ExpiresAt, Valid: true},
	})
	if err != nil {
		return domain.Session{}, err
	}
	return toDomainSession(row), nil
}

func (r *SessionRepository) FindByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (domain.Session, error) {
	row, err := r.queries.GetSessionByRefreshTokenHash(ctx, refreshTokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Session{}, application.ErrSessionNotFound
		}
		return domain.Session{}, err
	}
	return toDomainSession(row), nil
}

func (r *SessionRepository) Rotate(ctx context.Context, sessionID uuid.UUID, newRefreshTokenHash string, newExpiresAt time.Time) (domain.Session, error) {
	row, err := r.queries.RotateSession(ctx, sqlc.RotateSessionParams{
		ID:               sessionID,
		RefreshTokenHash: newRefreshTokenHash,
		ExpiresAt:        pgtype.Timestamptz{Time: newExpiresAt, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Session{}, application.ErrSessionNotFound
		}
		return domain.Session{}, err
	}
	return toDomainSession(row), nil
}

func (r *SessionRepository) DeleteByRefreshTokenHash(ctx context.Context, refreshTokenHash string) error {
	return r.queries.DeleteSessionByRefreshTokenHash(ctx, refreshTokenHash)
}

func toDomainSession(row sqlc.Session) domain.Session {
	return domain.Session{
		ID:               row.ID,
		UserID:           row.UserID,
		RefreshTokenHash: row.RefreshTokenHash,
		Device:           row.Device,
		IPAddress:        row.IpAddress,
		UserAgent:        row.UserAgent,
		CreatedAt:        row.CreatedAt.Time,
		ExpiresAt:        row.ExpiresAt.Time,
		LastActivityAt:   row.LastActivityAt.Time,
	}
}
