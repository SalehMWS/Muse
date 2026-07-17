package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/content/domain"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

type ListUseCase struct {
	repo ContentRepository
}

func NewListUseCase(repo ContentRepository) *ListUseCase {
	return &ListUseCase{repo: repo}
}

type ListInput struct {
	UserID      uuid.UUID
	Status      *string
	Language    *string
	ContentType *string
	Tag         *string
	Limit       int
	Cursor      string
}

type ListOutput struct {
	Items      []domain.Content
	NextCursor string
}

func (uc *ListUseCase) Execute(ctx context.Context, in ListInput) (ListOutput, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}

	filter := ListFilter{
		UserID:      in.UserID,
		Status:      in.Status,
		Language:    in.Language,
		ContentType: in.ContentType,
		Tag:         in.Tag,
		Limit:       int32(limit),
	}

	if in.Cursor != "" {
		createdAt, id, err := decodeCursor(in.Cursor)
		if err != nil {
			return ListOutput{}, err
		}
		filter.CursorCreatedAt = &createdAt
		filter.CursorID = &id
	}

	items, err := uc.repo.List(ctx, filter)
	if err != nil {
		return ListOutput{}, err
	}

	out := ListOutput{Items: items}
	if len(items) == limit {
		last := items[len(items)-1]
		out.NextCursor = encodeCursor(last.CreatedAt, last.ID)
	}
	return out, nil
}
