package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/knowledge/application"
	"github.com/SalehMWS/Muse/internal/knowledge/domain"
	"github.com/SalehMWS/Muse/internal/knowledge/infrastructure/embedding"
	"github.com/SalehMWS/Muse/internal/knowledge/infrastructure/vectorstore"
)

type fakeDocumentRepository struct {
	byID map[uuid.UUID]domain.Document
}

func newFakeDocumentRepository() *fakeDocumentRepository {
	return &fakeDocumentRepository{byID: map[uuid.UUID]domain.Document{}}
}

func (f *fakeDocumentRepository) Create(_ context.Context, document domain.Document) (domain.Document, error) {
	f.byID[document.ID] = document
	return document, nil
}

func (f *fakeDocumentRepository) FindByIDForUser(_ context.Context, id, userID uuid.UUID) (domain.Document, error) {
	document, ok := f.byID[id]
	if !ok || document.UserID != userID {
		return domain.Document{}, application.ErrDocumentNotFound
	}
	return document, nil
}

func (f *fakeDocumentRepository) ListByUser(_ context.Context, userID uuid.UUID) ([]domain.Document, error) {
	items := make([]domain.Document, 0)
	for _, document := range f.byID {
		if document.UserID == userID {
			items = append(items, document)
		}
	}
	return items, nil
}

func (f *fakeDocumentRepository) UpdateStatus(_ context.Context, id uuid.UUID, status domain.Status, chunkCount int, lastError *string) (domain.Document, error) {
	document, ok := f.byID[id]
	if !ok {
		return domain.Document{}, application.ErrDocumentNotFound
	}
	document.Status = status
	document.ChunkCount = chunkCount
	document.LastError = lastError
	f.byID[id] = document
	return document, nil
}

func (f *fakeDocumentRepository) DeleteForUser(_ context.Context, id, userID uuid.UUID) error {
	if document, ok := f.byID[id]; ok && document.UserID == userID {
		delete(f.byID, id)
	}
	return nil
}

type errorEmbedder struct{}

func (errorEmbedder) Dimension() int { return 256 }
func (errorEmbedder) Embed(context.Context, []string) ([][]float32, error) {
	return nil, errors.New("provider down")
}

func TestIngestAndQuery_RealPipeline(t *testing.T) {
	repo := newFakeDocumentRepository()
	embedder := embedding.NewLocalEmbedder(256)
	store := vectorstore.NewMemoryStore()
	userID := uuid.New()
	ctx := context.Background()

	ingest := application.NewIngestUseCase(repo, embedder, store, 50, 10, nil, nil)
	doc, err := ingest.Execute(ctx, application.IngestInput{
		UserID:  userID,
		Title:   "Brand voice",
		Content: "Our coffee is roasted fresh daily in small batches with great care and attention.",
	})
	if err != nil {
		t.Fatalf("Ingest Execute() unexpected error: %v", err)
	}
	if doc.Status != domain.StatusIndexed || doc.ChunkCount < 1 {
		t.Fatalf("ingested doc = %+v, want indexed with chunks", doc)
	}

	query := application.NewQueryUseCase(embedder, store, 5, nil)
	out, err := query.Execute(ctx, application.QueryInput{UserID: userID, Query: "fresh roasted coffee"})
	if err != nil {
		t.Fatalf("Query Execute() unexpected error: %v", err)
	}
	if len(out.Hits) == 0 {
		t.Fatal("Query returned no hits")
	}
	if !strings.Contains(strings.ToLower(out.Hits[0].Content), "coffee") {
		t.Fatalf("top hit = %q, want it to mention coffee", out.Hits[0].Content)
	}
	if strings.TrimSpace(out.Context) == "" {
		t.Fatal("Query built an empty context")
	}

	// tenant isolation: another user sees nothing
	other, _ := query.Execute(ctx, application.QueryInput{UserID: uuid.New(), Query: "coffee"})
	if len(other.Hits) != 0 {
		t.Fatalf("other user hits = %d, want 0", len(other.Hits))
	}
}

func TestIngest_EmbeddingFailureMarksFailed(t *testing.T) {
	repo := newFakeDocumentRepository()
	userID := uuid.New()
	uc := application.NewIngestUseCase(repo, errorEmbedder{}, vectorstore.NewMemoryStore(), 50, 10, nil, nil)

	if _, err := uc.Execute(context.Background(), application.IngestInput{UserID: userID, Content: "some text"}); !errors.Is(err, application.ErrEmbedding) {
		t.Fatalf("Execute() error = %v, want %v", err, application.ErrEmbedding)
	}

	docs, _ := repo.ListByUser(context.Background(), userID)
	if len(docs) != 1 || docs[0].Status != domain.StatusFailed {
		t.Fatalf("document should be marked failed, got %+v", docs)
	}
}

func TestIngest_EmptyContent(t *testing.T) {
	uc := application.NewIngestUseCase(newFakeDocumentRepository(), embedding.NewLocalEmbedder(64), vectorstore.NewMemoryStore(), 50, 10, nil, nil)
	if _, err := uc.Execute(context.Background(), application.IngestInput{UserID: uuid.New(), Content: "   "}); !errors.Is(err, application.ErrEmptyContent) {
		t.Fatalf("Execute() error = %v, want %v", err, application.ErrEmptyContent)
	}
}
