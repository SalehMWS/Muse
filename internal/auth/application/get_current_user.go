package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/auth/domain"
)

type GetCurrentUserUseCase struct {
	users UserRepository
}

func NewGetCurrentUserUseCase(users UserRepository) *GetCurrentUserUseCase {
	return &GetCurrentUserUseCase{users: users}
}

func (uc *GetCurrentUserUseCase) Execute(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	user, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("get current user: %w", err)
	}

	return user, nil
}
