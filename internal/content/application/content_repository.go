package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/content/domain"
)

type ListFilter struct {
	UserID          uuid.UUID
	Status          *string
	Language        *string
	ContentType     *string
	Tag             *string
	CursorCreatedAt *time.Time
	CursorID        *uuid.UUID
	Limit           int32
}

type ContentRepository interface {
	Create(ctx context.Context, content domain.Content) (domain.Content, error)
	FindByIDForUser(ctx context.Context, id, userID uuid.UUID) (domain.Content, error)
	Update(ctx context.Context, content domain.Content) (domain.Content, error)
	List(ctx context.Context, filter ListFilter) ([]domain.Content, error)
}
