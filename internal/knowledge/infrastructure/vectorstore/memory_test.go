package vectorstore_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/knowledge/application"
	"github.com/SalehMWS/Muse/internal/knowledge/infrastructure/vectorstore"
)

func TestMemoryStore_UpsertSearchDelete(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	ctx := context.Background()
	userID := uuid.New()
	docID := uuid.New()

	records := []application.VectorRecord{
		{ID: "a", UserID: userID, DocumentID: docID, ChunkIndex: 0, Content: "coffee beans", Embedding: []float32{1, 0, 0}},
		{ID: "b", UserID: userID, DocumentID: docID, ChunkIndex: 1, Content: "tax report", Embedding: []float32{0, 1, 0}},
		{ID: "c", UserID: uuid.New(), DocumentID: uuid.New(), ChunkIndex: 0, Content: "other user", Embedding: []float32{1, 0, 0}},
	}
	if err := store.Upsert(ctx, records); err != nil {
		t.Fatalf("Upsert() unexpected error: %v", err)
	}

	hits, err := store.Search(ctx, userID, []float32{1, 0, 0}, 5)
	if err != nil {
		t.Fatalf("Search() unexpected error: %v", err)
	}
	if len(hits) != 2 {
		t.Fatalf("Search() hits = %d, want 2 (tenant-scoped)", len(hits))
	}
	if hits[0].Content != "coffee beans" {
		t.Fatalf("top hit = %q, want coffee beans (closest vector)", hits[0].Content)
	}

	if err := store.DeleteByDocument(ctx, userID, docID); err != nil {
		t.Fatalf("DeleteByDocument() unexpected error: %v", err)
	}
	remaining, _ := store.Search(ctx, userID, []float32{1, 0, 0}, 5)
	if len(remaining) != 0 {
		t.Fatalf("after delete, hits = %d, want 0", len(remaining))
	}
}
