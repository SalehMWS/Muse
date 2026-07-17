package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/publishing/domain"
)

type PublicationRepository interface {
	Create(ctx context.Context, publication domain.Publication) (domain.Publication, error)
	ListByContentForUser(ctx context.Context, userID, contentID uuid.UUID) ([]domain.Publication, error)
}
