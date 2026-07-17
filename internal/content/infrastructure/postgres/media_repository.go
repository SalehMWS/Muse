package postgres

import (
	"context"
	"errors"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/SalehMWS/Muse/internal/content/application"
	"github.com/SalehMWS/Muse/internal/content/domain"
	"github.com/SalehMWS/Muse/internal/shared/database/sqlc"
)

type MediaRepository struct {
	queries *sqlc.Queries
}

func NewMediaRepository(db sqlc.DBTX) *MediaRepository {
	return &MediaRepository{queries: sqlc.New(db)}
}

func (r *MediaRepository) Create(ctx context.Context, media domain.Media) (domain.Media, error) {
	position := media.Position
	if position < 0 {
		position = 0
	}
	if position > math.MaxInt32 {
		position = math.MaxInt32
	}

	row, err := r.queries.CreateMedia(ctx, sqlc.CreateMediaParams{
		ID:        media.ID,
		ContentID: media.ContentID,
		Url:       media.URL,
		MediaType: string(media.MediaType),
		Position:  int32(position),
	})
	if err != nil {
		return domain.Media{}, err
	}
	return toDomainMedia(row), nil
}

func (r *MediaRepository) FindByIDForContent(ctx context.Context, id, contentID uuid.UUID) (domain.Media, error) {
	row, err := r.queries.GetMediaByIDForContent(ctx, sqlc.GetMediaByIDForContentParams{ID: id, ContentID: contentID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Media{}, application.ErrMediaNotFound
		}
		return domain.Media{}, err
	}
	return toDomainMedia(row), nil
}

func (r *MediaRepository) ListByContent(ctx context.Context, contentID uuid.UUID) ([]domain.Media, error) {
	rows, err := r.queries.ListMediaByContent(ctx, contentID)
	if err != nil {
		return nil, err
	}
	items := make([]domain.Media, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainMedia(row))
	}
	return items, nil
}

func (r *MediaRepository) DeleteForContent(ctx context.Context, id, contentID uuid.UUID) error {
	return r.queries.DeleteMediaForContent(ctx, sqlc.DeleteMediaForContentParams{ID: id, ContentID: contentID})
}

func toDomainMedia(row sqlc.Medium) domain.Media {
	return domain.Media{
		ID:        row.ID,
		ContentID: row.ContentID,
		URL:       row.Url,
		MediaType: domain.MediaType(row.MediaType),
		Position:  int(row.Position),
		CreatedAt: row.CreatedAt.Time,
	}
}
