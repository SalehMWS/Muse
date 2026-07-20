package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/audit/domain"
)

type EventRepository interface {
	Append(ctx context.Context, event domain.Event) error
	ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Event, error)
}
