package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/instagram/domain"
)

type AccountRepository interface {
	Upsert(ctx context.Context, account domain.ConnectedAccount) (domain.ConnectedAccount, error)
	FindByIDForUser(ctx context.Context, id, userID uuid.UUID) (domain.ConnectedAccount, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.ConnectedAccount, error)
	UpdateToken(ctx context.Context, id uuid.UUID, accessToken string, expiresAt time.Time, status domain.AccountStatus) (domain.ConnectedAccount, error)
	DeleteForUser(ctx context.Context, id, userID uuid.UUID) error
}
