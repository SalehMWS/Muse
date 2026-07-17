package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/content/domain"
)

type DuplicateUseCase struct {
	repo ContentRepository
}

func NewDuplicateUseCase(repo ContentRepository) *DuplicateUseCase {
	return &DuplicateUseCase{repo: repo}
}

func (uc *DuplicateUseCase) Execute(ctx context.Context, userID, id uuid.UUID) (domain.Content, error) {
	original, err := uc.repo.FindByIDForUser(ctx, id, userID)
	if err != nil {
		return domain.Content{}, err
	}

	return uc.repo.Create(ctx, original.Duplicate())
}
