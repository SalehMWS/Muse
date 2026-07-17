package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/content/domain"
)

type UpdateUseCase struct {
	repo ContentRepository
}

func NewUpdateUseCase(repo ContentRepository) *UpdateUseCase {
	return &UpdateUseCase{repo: repo}
}

func (uc *UpdateUseCase) Execute(ctx context.Context, userID, id uuid.UUID, in domain.UpdateContentInput) (domain.Content, error) {
	content, err := uc.repo.FindByIDForUser(ctx, id, userID)
	if err != nil {
		return domain.Content{}, err
	}

	if err := content.Apply(in); err != nil {
		return domain.Content{}, err
	}

	return uc.repo.Update(ctx, content)
}
