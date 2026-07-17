package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/SalehMWS/Muse/internal/publishing/domain"
	"github.com/SalehMWS/Muse/internal/shared/database/sqlc"
)

type PublicationRepository struct {
	queries *sqlc.Queries
}

func NewPublicationRepository(db sqlc.DBTX) *PublicationRepository {
	return &PublicationRepository{queries: sqlc.New(db)}
}

func (r *PublicationRepository) Create(ctx context.Context, publication domain.Publication) (domain.Publication, error) {
	params := sqlc.CreatePublicationParams{
		ID:                 publication.ID,
		UserID:             publication.UserID,
		ContentID:          publication.ContentID,
		InstagramAccountID: publication.InstagramAccountID,
		Platform:           publication.Platform,
		PlatformPostID:     publication.PlatformPostID,
		Status:             string(publication.Status),
		Permalink:          publication.Permalink,
		ResponseJson:       publication.ResponseJSON,
	}
	if publication.PublishedAt != nil {
		params.PublishedAt = pgtype.Timestamptz{Time: *publication.PublishedAt, Valid: true}
	}

	row, err := r.queries.CreatePublication(ctx, params)
	if err != nil {
		return domain.Publication{}, err
	}
	return toDomainPublication(row), nil
}

func (r *PublicationRepository) ListByContentForUser(ctx context.Context, userID, contentID uuid.UUID) ([]domain.Publication, error) {
	rows, err := r.queries.ListPublicationsByContentForUser(ctx, sqlc.ListPublicationsByContentForUserParams{
		ContentID: contentID,
		UserID:    userID,
	})
	if err != nil {
		return nil, err
	}
	items := make([]domain.Publication, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainPublication(row))
	}
	return items, nil
}

func toDomainPublication(row sqlc.Publication) domain.Publication {
	publication := domain.Publication{
		ID:                 row.ID,
		UserID:             row.UserID,
		ContentID:          row.ContentID,
		InstagramAccountID: row.InstagramAccountID,
		Platform:           row.Platform,
		PlatformPostID:     row.PlatformPostID,
		Status:             domain.Status(row.Status),
		Permalink:          row.Permalink,
		ResponseJSON:       row.ResponseJson,
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
	}
	if row.PublishedAt.Valid {
		published := row.PublishedAt.Time
		publication.PublishedAt = &published
	}
	return publication
}
