package instrumented

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/knowledge/application"
	"github.com/SalehMWS/Muse/internal/shared/metrics"
)

type Embedder struct {
	inner    application.Embedder
	name     string
	recorder *metrics.Knowledge
}

var _ application.Embedder = (*Embedder)(nil)

func NewEmbedder(inner application.Embedder, name string, recorder *metrics.Knowledge) *Embedder {
	return &Embedder{inner: inner, name: name, recorder: recorder}
}

func (e *Embedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	start := time.Now()
	vectors, err := e.inner.Embed(ctx, texts)
	e.recorder.Embedded(e.name, time.Since(start))
	return vectors, err
}

func (e *Embedder) Dimension() int {
	return e.inner.Dimension()
}

type VectorStore struct {
	inner    application.VectorStore
	name     string
	recorder *metrics.Knowledge
}

var _ application.VectorStore = (*VectorStore)(nil)

func NewVectorStore(inner application.VectorStore, name string, recorder *metrics.Knowledge) *VectorStore {
	return &VectorStore{inner: inner, name: name, recorder: recorder}
}

func (s *VectorStore) EnsureReady(ctx context.Context, dimension int) error {
	return s.inner.EnsureReady(ctx, dimension)
}

func (s *VectorStore) Ping(ctx context.Context) error {
	return s.inner.Ping(ctx)
}

func (s *VectorStore) Upsert(ctx context.Context, records []application.VectorRecord) error {
	return s.inner.Upsert(ctx, records)
}

func (s *VectorStore) Search(ctx context.Context, userID uuid.UUID, embedding []float32, topK int) ([]application.SearchHit, error) {
	start := time.Now()
	hits, err := s.inner.Search(ctx, userID, embedding, topK)
	s.recorder.Searched(s.name, time.Since(start))
	return hits, err
}

func (s *VectorStore) DeleteByDocument(ctx context.Context, userID, documentID uuid.UUID) error {
	return s.inner.DeleteByDocument(ctx, userID, documentID)
}
