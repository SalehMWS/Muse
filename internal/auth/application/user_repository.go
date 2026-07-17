package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/auth/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	FindByEmail(ctx context.Context, email domain.Email) (domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (domain.User, error)
}
