package postgres

import (
	"context"
	"errors"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/SalehMWS/Muse/internal/knowledge/application"
	"github.com/SalehMWS/Muse/internal/knowledge/domain"
	"github.com/SalehMWS/Muse/internal/shared/database/sqlc"
)

type DocumentRepository struct {
	queries *sqlc.Queries
}

func NewDocumentRepository(db sqlc.DBTX) *DocumentRepository {
	return &DocumentRepository{queries: sqlc.New(db)}
}

func (r *DocumentRepository) Create(ctx context.Context, document domain.Document) (domain.Document, error) {
	row, err := r.queries.CreateDocument(ctx, sqlc.CreateDocumentParams{
		ID:     document.ID,
		UserID: document.UserID,
		Title:  document.Title,
		Source: document.Source,
		Status: string(document.Status),
	})
	if err != nil {
		return domain.Document{}, err
	}
	return toDomainDocument(row), nil
}

func (r *DocumentRepository) FindByIDForUser(ctx context.Context, id, userID uuid.UUID) (domain.Document, error) {
	row, err := r.queries.GetDocumentByIDForUser(ctx, sqlc.GetDocumentByIDForUserParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Document{}, application.ErrDocumentNotFound
		}
		return domain.Document{}, err
	}
	return toDomainDocument(row), nil
}

func (r *DocumentRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Document, error) {
	rows, err := r.queries.ListDocumentsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]domain.Document, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainDocument(row))
	}
	return items, nil
}

func (r *DocumentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status, chunkCount int, lastError *string) (domain.Document, error) {
	row, err := r.queries.UpdateDocumentStatus(ctx, sqlc.UpdateDocumentStatusParams{
		ID:         id,
		Status:     string(status),
		ChunkCount: toInt32(chunkCount),
		LastError:  lastError,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Document{}, application.ErrDocumentNotFound
		}
		return domain.Document{}, err
	}
	return toDomainDocument(row), nil
}

func (r *DocumentRepository) DeleteForUser(ctx context.Context, id, userID uuid.UUID) error {
	return r.queries.DeleteDocumentForUser(ctx, sqlc.DeleteDocumentForUserParams{ID: id, UserID: userID})
}

func toDomainDocument(row sqlc.Document) domain.Document {
	return domain.Document{
		ID:         row.ID,
		UserID:     row.UserID,
		Title:      row.Title,
		Source:     row.Source,
		Status:     domain.Status(row.Status),
		ChunkCount: int(row.ChunkCount),
		LastError:  row.LastError,
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}
}

func toInt32(value int) int32 {
	if value < 0 {
		return 0
	}
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(value)
}
