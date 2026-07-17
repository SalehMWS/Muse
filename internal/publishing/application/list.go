package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/publishing/domain"
)

type ListPublicationsUseCase struct {
	repo PublicationRepository
}

func NewListPublicationsUseCase(repo PublicationRepository) *ListPublicationsUseCase {
	return &ListPublicationsUseCase{repo: repo}
}

func (uc *ListPublicationsUseCase) Execute(ctx context.Context, userID, contentID uuid.UUID) ([]domain.Publication, error) {
	return uc.repo.ListByContentForUser(ctx, userID, contentID)
}
