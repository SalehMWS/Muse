package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/content/domain"
)

type MediaRepository interface {
	Create(ctx context.Context, media domain.Media) (domain.Media, error)
	FindByIDForContent(ctx context.Context, id, contentID uuid.UUID) (domain.Media, error)
	ListByContent(ctx context.Context, contentID uuid.UUID) ([]domain.Media, error)
	DeleteForContent(ctx context.Context, id, contentID uuid.UUID) error
}
