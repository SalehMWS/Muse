package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/content/domain"
)

type GetUseCase struct {
	repo ContentRepository
}

func NewGetUseCase(repo ContentRepository) *GetUseCase {
	return &GetUseCase{repo: repo}
}

func (uc *GetUseCase) Execute(ctx context.Context, userID, id uuid.UUID) (domain.Content, error) {
	return uc.repo.FindByIDForUser(ctx, id, userID)
}
