package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/knowledge/application"
	httpdelivery "github.com/SalehMWS/Muse/internal/knowledge/delivery/http"
	"github.com/SalehMWS/Muse/internal/knowledge/domain"
	"github.com/SalehMWS/Muse/internal/knowledge/infrastructure/embedding"
	"github.com/SalehMWS/Muse/internal/knowledge/infrastructure/vectorstore"
	"github.com/SalehMWS/Muse/internal/shared/authcontext"
)

type fakeDocumentRepository struct {
	byID map[uuid.UUID]domain.Document
}

func newFakeDocumentRepository() *fakeDocumentRepository {
	return &fakeDocumentRepository{byID: map[uuid.UUID]domain.Document{}}
}

func (f *fakeDocumentRepository) Create(_ context.Context, d domain.Document) (domain.Document, error) {
	f.byID[d.ID] = d
	return d, nil
}
func (f *fakeDocumentRepository) FindByIDForUser(_ context.Context, id, userID uuid.UUID) (domain.Document, error) {
	d, ok := f.byID[id]
	if !ok || d.UserID != userID {
		return domain.Document{}, application.ErrDocumentNotFound
	}
	return d, nil
}
func (f *fakeDocumentRepository) ListByUser(_ context.Context, userID uuid.UUID) ([]domain.Document, error) {
	out := make([]domain.Document, 0)
	for _, d := range f.byID {
		if d.UserID == userID {
			out = append(out, d)
		}
	}
	return out, nil
}
func (f *fakeDocumentRepository) UpdateStatus(_ context.Context, id uuid.UUID, status domain.Status, chunkCount int, lastError *string) (domain.Document, error) {
	d := f.byID[id]
	d.Status = status
	d.ChunkCount = chunkCount
	d.LastError = lastError
	f.byID[id] = d
	return d, nil
}
func (f *fakeDocumentRepository) DeleteForUser(_ context.Context, id, userID uuid.UUID) error {
	delete(f.byID, id)
	return nil
}

func newTestApp(userID uuid.UUID) *fiber.App {
	repo := newFakeDocumentRepository()
	embedder := embedding.NewLocalEmbedder(256)
	store := vectorstore.NewMemoryStore()
	handler := httpdelivery.NewHandler(
		application.NewIngestUseCase(repo, embedder, store, 50, 10, nil, nil),
		application.NewQueryUseCase(embedder, store, 5, nil),
		application.NewListDocumentsUseCase(repo),
		application.NewDeleteDocumentUseCase(repo, store),
	)

	app := fiber.New()
	requireAuth := func(c *fiber.Ctx) error {
		authcontext.SetUser(c, userID, uuid.New())
		return c.Next()
	}
	httpdelivery.RegisterRoutes(app, handler, requireAuth)
	return app
}

func doJSON(t *testing.T, app *fiber.App, method, path string, body any) (int, map[string]any) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequestWithContext(context.Background(), method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	rawResp, _ := io.ReadAll(resp.Body)
	var parsed map[string]any
	if len(rawResp) > 0 {
		if err := json.Unmarshal(rawResp, &parsed); err != nil {
			t.Fatalf("unmarshal body: %v (%s)", err, rawResp)
		}
	}
	return resp.StatusCode, parsed
}

func TestHandler_IngestAndQuery(t *testing.T) {
	app := newTestApp(uuid.New())

	status, body := doJSON(t, app, fiber.MethodPost, "/knowledge/documents", map[string]any{
		"title":   "Brand",
		"content": "Our coffee is roasted fresh daily in small batches.",
	})
	if status != fiber.StatusCreated {
		t.Fatalf("ingest status = %d, want %d, body=%v", status, fiber.StatusCreated, body)
	}
	data, _ := body["data"].(map[string]any)
	if data["status"] != string(domain.StatusIndexed) {
		t.Fatalf("document status = %v, want indexed", data["status"])
	}

	status, qbody := doJSON(t, app, fiber.MethodPost, "/knowledge/query", map[string]any{"query": "fresh roasted coffee"})
	if status != fiber.StatusOK {
		t.Fatalf("query status = %d, want %d", status, fiber.StatusOK)
	}
	qdata, _ := qbody["data"].(map[string]any)
	hits, _ := qdata["hits"].([]any)
	if len(hits) == 0 {
		t.Fatalf("query hits = 0, want >= 1")
	}
	if ctxStr, _ := qdata["context"].(string); ctxStr == "" {
		t.Fatal("query context is empty")
	}
}

func TestHandler_IngestEmptyContent(t *testing.T) {
	app := newTestApp(uuid.New())
	status, _ := doJSON(t, app, fiber.MethodPost, "/knowledge/documents", map[string]any{"title": "x"})
	if status != fiber.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", status, fiber.StatusUnprocessableEntity)
	}
}

func TestHandler_ListAndDelete(t *testing.T) {
	app := newTestApp(uuid.New())
	_, body := doJSON(t, app, fiber.MethodPost, "/knowledge/documents", map[string]any{"content": "hello world content"})
	data, _ := body["data"].(map[string]any)
	id, _ := data["id"].(string)

	status, listBody := doJSON(t, app, fiber.MethodGet, "/knowledge/documents", nil)
	if status != fiber.StatusOK {
		t.Fatalf("list status = %d, want %d", status, fiber.StatusOK)
	}
	items, _ := listBody["data"].([]any)
	if len(items) != 1 {
		t.Fatalf("documents = %d, want 1", len(items))
	}

	status, _ = doJSON(t, app, fiber.MethodDelete, "/knowledge/documents/"+id, nil)
	if status != fiber.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", status, fiber.StatusNoContent)
	}
}
