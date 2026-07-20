package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/audit/domain"
)

const (
	defaultListLimit = 50
	maxListLimit     = 200
)

type ListEventsUseCase struct {
	events EventRepository
}

func NewListEventsUseCase(events EventRepository) *ListEventsUseCase {
	return &ListEventsUseCase{events: events}
}

func (uc *ListEventsUseCase) Execute(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Event, error) {
	switch {
	case limit <= 0:
		limit = defaultListLimit
	case limit > maxListLimit:
		limit = maxListLimit
	}

	events, err := uc.events.ListByUser(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("audit: list events: %w", err)
	}
	return events, nil
}
