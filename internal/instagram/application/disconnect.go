package application

import (
	"context"

	"github.com/google/uuid"
)

type DisconnectUseCase struct {
	repo AccountRepository
}

func NewDisconnectUseCase(repo AccountRepository) *DisconnectUseCase {
	return &DisconnectUseCase{repo: repo}
}

func (uc *DisconnectUseCase) Execute(ctx context.Context, userID, accountID uuid.UUID) error {
	if _, err := uc.repo.FindByIDForUser(ctx, accountID, userID); err != nil {
		return err
	}
	return uc.repo.DeleteForUser(ctx, accountID, userID)
}
