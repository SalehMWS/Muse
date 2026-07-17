package application

import (
	"context"

	"github.com/google/uuid"
)

type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	Dimension() int
}

type VectorRecord struct {
	ID         string
	UserID     uuid.UUID
	DocumentID uuid.UUID
	ChunkIndex int
	Content    string
	Embedding  []float32
}

type SearchHit struct {
	DocumentID uuid.UUID
	ChunkIndex int
	Content    string
	Score      float32
}

type VectorStore interface {
	EnsureReady(ctx context.Context, dimension int) error
	Upsert(ctx context.Context, records []VectorRecord) error
	Search(ctx context.Context, userID uuid.UUID, embedding []float32, topK int) ([]SearchHit, error)
	DeleteByDocument(ctx context.Context, userID, documentID uuid.UUID) error
}
