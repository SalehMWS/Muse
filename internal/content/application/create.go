package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/content/domain"
)

type CreateUseCase struct {
	repo ContentRepository
}

func NewCreateUseCase(repo ContentRepository) *CreateUseCase {
	return &CreateUseCase{repo: repo}
}

func (uc *CreateUseCase) Execute(ctx context.Context, userID uuid.UUID, in domain.NewContentInput) (domain.Content, error) {
	content, err := domain.NewContent(userID, in)
	if err != nil {
		return domain.Content{}, err
	}
	return uc.repo.Create(ctx, content)
}
