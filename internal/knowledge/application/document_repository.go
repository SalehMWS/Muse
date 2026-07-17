package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/knowledge/domain"
)

type DocumentRepository interface {
	Create(ctx context.Context, document domain.Document) (domain.Document, error)
	FindByIDForUser(ctx context.Context, id, userID uuid.UUID) (domain.Document, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Document, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status, chunkCount int, lastError *string) (domain.Document, error)
	DeleteForUser(ctx context.Context, id, userID uuid.UUID) error
}
