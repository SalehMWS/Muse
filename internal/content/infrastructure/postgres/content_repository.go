package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/SalehMWS/Muse/internal/content/application"
	"github.com/SalehMWS/Muse/internal/content/domain"
	"github.com/SalehMWS/Muse/internal/shared/database/sqlc"
)

type ContentRepository struct {
	queries *sqlc.Queries
}

func NewContentRepository(db sqlc.DBTX) *ContentRepository {
	return &ContentRepository{queries: sqlc.New(db)}
}

func (r *ContentRepository) Create(ctx context.Context, content domain.Content) (domain.Content, error) {
	row, err := r.queries.CreateContent(ctx, sqlc.CreateContentParams{
		ID:          content.ID,
		UserID:      content.UserID,
		Title:       content.Title,
		Caption:     content.Caption,
		Status:      string(content.Status),
		Language:    content.Language,
		ContentType: string(content.ContentType),
		Visibility:  string(content.Visibility),
		Tags:        tagsOrEmpty(content.Tags),
	})
	if err != nil {
		return domain.Content{}, err
	}
	return toDomainContent(row), nil
}

func (r *ContentRepository) FindByIDForUser(ctx context.Context, id, userID uuid.UUID) (domain.Content, error) {
	row, err := r.queries.GetContentByIDForUser(ctx, sqlc.GetContentByIDForUserParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Content{}, application.ErrContentNotFound
		}
		return domain.Content{}, err
	}
	return toDomainContent(row), nil
}

func (r *ContentRepository) Update(ctx context.Context, content domain.Content) (domain.Content, error) {
	row, err := r.queries.UpdateContent(ctx, sqlc.UpdateContentParams{
		ID:          content.ID,
		UserID:      content.UserID,
		Title:       content.Title,
		Caption:     content.Caption,
		Status:      string(content.Status),
		Language:    content.Language,
		ContentType: string(content.ContentType),
		Visibility:  string(content.Visibility),
		Tags:        tagsOrEmpty(content.Tags),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Content{}, application.ErrContentNotFound
		}
		return domain.Content{}, err
	}
	return toDomainContent(row), nil
}

func tagsOrEmpty(tags []string) []string {
	if tags == nil {
		return []string{}
	}
	return tags
}

func (r *ContentRepository) List(ctx context.Context, filter application.ListFilter) ([]domain.Content, error) {
	if filter.CursorCreatedAt == nil {
		rows, err := r.queries.ListContents(ctx, sqlc.ListContentsParams{
			UserID:      filter.UserID,
			Status:      filter.Status,
			Language:    filter.Language,
			ContentType: filter.ContentType,
			Tag:         filter.Tag,
			Lim:         filter.Limit,
		})
		if err != nil {
			return nil, err
		}
		return mapContents(rows), nil
	}

	rows, err := r.queries.ListContentsAfter(ctx, sqlc.ListContentsAfterParams{
		UserID:          filter.UserID,
		Status:          filter.Status,
		Language:        filter.Language,
		ContentType:     filter.ContentType,
		Tag:             filter.Tag,
		CursorCreatedAt: pgtype.Timestamptz{Time: *filter.CursorCreatedAt, Valid: true},
		CursorID:        *filter.CursorID,
		Lim:             filter.Limit,
	})
	if err != nil {
		return nil, err
	}
	return mapContents(rows), nil
}

func mapContents(rows []sqlc.Content) []domain.Content {
	items := make([]domain.Content, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainContent(row))
	}
	return items
}

func toDomainContent(row sqlc.Content) domain.Content {
	content := domain.Content{
		ID:          row.ID,
		UserID:      row.UserID,
		Title:       row.Title,
		Caption:     row.Caption,
		Status:      domain.Status(row.Status),
		Language:    row.Language,
		ContentType: domain.ContentType(row.ContentType),
		Visibility:  domain.Visibility(row.Visibility),
		Tags:        row.Tags,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
	if row.PublishedAt.Valid {
		published := row.PublishedAt.Time
		content.PublishedAt = &published
	}
	return content
}
