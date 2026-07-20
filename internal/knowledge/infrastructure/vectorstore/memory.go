package vectorstore

import (
	"context"
	"math"
	"sort"
	"sync"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/knowledge/application"
)

type MemoryStore struct {
	mu      sync.RWMutex
	records map[string]application.VectorRecord
}

var _ application.VectorStore = (*MemoryStore)(nil)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{records: map[string]application.VectorRecord{}}
}

func (s *MemoryStore) Ping(context.Context) error {
	return nil
}

func (s *MemoryStore) EnsureReady(context.Context, int) error {
	return nil
}

func (s *MemoryStore) Upsert(_ context.Context, records []application.VectorRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, record := range records {
		s.records[record.ID] = record
	}
	return nil
}

func (s *MemoryStore) Search(_ context.Context, userID uuid.UUID, embedding []float32, topK int) ([]application.SearchHit, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hits := make([]application.SearchHit, 0)
	for _, record := range s.records {
		if record.UserID != userID {
			continue
		}
		hits = append(hits, application.SearchHit{
			DocumentID: record.DocumentID,
			ChunkIndex: record.ChunkIndex,
			Content:    record.Content,
			Score:      cosine(embedding, record.Embedding),
		})
	}

	sort.Slice(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	if topK < len(hits) {
		hits = hits[:topK]
	}
	return hits, nil
}

func (s *MemoryStore) DeleteByDocument(_ context.Context, userID, documentID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, record := range s.records {
		if record.UserID == userID && record.DocumentID == documentID {
			delete(s.records, id)
		}
	}
	return nil
}

func cosine(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return float32(dot / (math.Sqrt(normA) * math.Sqrt(normB)))
}
