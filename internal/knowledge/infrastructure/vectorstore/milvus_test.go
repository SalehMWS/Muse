package vectorstore_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	milvus "github.com/milvus-io/milvus-sdk-go/v2/client"

	"github.com/SalehMWS/Muse/internal/knowledge/application"
	"github.com/SalehMWS/Muse/internal/knowledge/infrastructure/vectorstore"
)

func testMilvusStore(t *testing.T) (*vectorstore.MilvusStore, func()) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := milvus.NewClient(ctx, milvus.Config{Address: "localhost:19530"})
	if err != nil {
		t.Skipf("milvus integration: skipping, cannot reach milvus: %v", err)
	}

	collection := "test_kb_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	store := vectorstore.NewMilvusStore(client, collection)
	cleanup := func() {
		dropCtx, dropCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer dropCancel()
		_ = client.DropCollection(dropCtx, collection)
		_ = client.Close()
	}
	return store, cleanup
}

func searchUntil(t *testing.T, store *vectorstore.MilvusStore, userID uuid.UUID, vec []float32, want int) []application.SearchHit {
	t.Helper()
	ctx := context.Background()
	var hits []application.SearchHit
	for i := 0; i < 15; i++ {
		var err error
		hits, err = store.Search(ctx, userID, vec, 5)
		if err != nil {
			t.Fatalf("Search() unexpected error: %v", err)
		}
		if len(hits) == want {
			return hits
		}
		time.Sleep(300 * time.Millisecond)
	}
	return hits
}

func TestMilvusStore_Integration(t *testing.T) {
	store, cleanup := testMilvusStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := store.EnsureReady(ctx, 8); err != nil {
		t.Fatalf("EnsureReady() unexpected error: %v", err)
	}

	userID := uuid.New()
	docID := uuid.New()
	vec := []float32{0.5, 0.5, 0.5, 0.5, 0, 0, 0, 0}

	if err := store.Upsert(ctx, []application.VectorRecord{{
		ID: docID.String() + "-0", UserID: userID, DocumentID: docID, ChunkIndex: 0,
		Content: "coffee beans roasted fresh", Embedding: vec,
	}}); err != nil {
		t.Fatalf("Upsert() unexpected error: %v", err)
	}

	hits := searchUntil(t, store, userID, vec, 1)
	if len(hits) != 1 {
		t.Fatalf("Search() hits = %d, want 1", len(hits))
	}
	if hits[0].DocumentID != docID || hits[0].Content != "coffee beans roasted fresh" {
		t.Fatalf("hit = %+v, unexpected", hits[0])
	}

	other, err := store.Search(ctx, uuid.New(), vec, 5)
	if err != nil {
		t.Fatalf("Search() other user error: %v", err)
	}
	if len(other) != 0 {
		t.Fatalf("other user hits = %d, want 0 (tenant isolation)", len(other))
	}

	if err := store.DeleteByDocument(ctx, userID, docID); err != nil {
		t.Fatalf("DeleteByDocument() unexpected error: %v", err)
	}
	remaining := searchUntil(t, store, userID, vec, 0)
	if len(remaining) != 0 {
		t.Fatalf("after delete, hits = %d, want 0", len(remaining))
	}
}
