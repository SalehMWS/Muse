package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/instagram/domain"
)

type ListUseCase struct {
	repo AccountRepository
}

func NewListUseCase(repo AccountRepository) *ListUseCase {
	return &ListUseCase{repo: repo}
}

func (uc *ListUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]domain.ConnectedAccount, error) {
	accounts, err := uc.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	for i := range accounts {
		accounts[i].AccessToken = ""
	}
	return accounts, nil
}
