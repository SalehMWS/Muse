package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/content/domain"
)

type ArchiveUseCase struct {
	repo ContentRepository
}

func NewArchiveUseCase(repo ContentRepository) *ArchiveUseCase {
	return &ArchiveUseCase{repo: repo}
}

func (uc *ArchiveUseCase) Execute(ctx context.Context, userID, id uuid.UUID) (domain.Content, error) {
	content, err := uc.repo.FindByIDForUser(ctx, id, userID)
	if err != nil {
		return domain.Content{}, err
	}

	content.Archive()
	return uc.repo.Update(ctx, content)
}
