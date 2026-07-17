package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/content/application"
	httpdelivery "github.com/SalehMWS/Muse/internal/content/delivery/http"
	"github.com/SalehMWS/Muse/internal/content/domain"
	"github.com/SalehMWS/Muse/internal/shared/authcontext"
)

func newTestApp(userID uuid.UUID) *fiber.App {
	repo := newFakeContentRepository()
	handler := httpdelivery.NewHandler(
		application.NewCreateUseCase(repo),
		application.NewGetUseCase(repo),
		application.NewUpdateUseCase(repo),
		application.NewArchiveUseCase(repo),
		application.NewDuplicateUseCase(repo),
		application.NewListUseCase(repo),
	)

	app := fiber.New()
	requireAuth := func(c *fiber.Ctx) error {
		authcontext.SetUser(c, userID, uuid.New())
		return c.Next()
	}
	httpdelivery.RegisterRoutes(app.Group("/contents"), handler, requireAuth)
	return app
}

func doJSON(t *testing.T, app *fiber.App, method, path string, body any) (int, map[string]any) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequestWithContext(context.Background(), method, path, reader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var parsed map[string]any
	if len(rawResp) > 0 {
		if err := json.Unmarshal(rawResp, &parsed); err != nil {
			t.Fatalf("unmarshal body: %v (%s)", err, rawResp)
		}
	}
	return resp.StatusCode, parsed
}

func createContent(t *testing.T, app *fiber.App, title string) string {
	t.Helper()
	status, body := doJSON(t, app, fiber.MethodPost, "/contents", map[string]any{"title": title})
	if status != fiber.StatusCreated {
		t.Fatalf("create status = %d, want %d, body=%v", status, fiber.StatusCreated, body)
	}
	data, _ := body["data"].(map[string]any)
	id, _ := data["id"].(string)
	if id == "" {
		t.Fatalf("create returned no id: %v", body)
	}
	return id
}

func TestHandler_CreateAndGet(t *testing.T) {
	app := newTestApp(uuid.New())

	id := createContent(t, app, "My Post")

	status, body := doJSON(t, app, fiber.MethodGet, "/contents/"+id, nil)
	if status != fiber.StatusOK {
		t.Fatalf("get status = %d, want %d", status, fiber.StatusOK)
	}
	data, _ := body["data"].(map[string]any)
	if data["status"] != string(domain.StatusDraft) {
		t.Fatalf("status = %v, want draft", data["status"])
	}
}

func TestHandler_CreateInvalid(t *testing.T) {
	app := newTestApp(uuid.New())
	status, _ := doJSON(t, app, fiber.MethodPost, "/contents", map[string]any{"title": strings.Repeat("x", domain.MaxTitleLength+1)})
	if status != fiber.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", status, fiber.StatusUnprocessableEntity)
	}
}

func TestHandler_GetErrors(t *testing.T) {
	app := newTestApp(uuid.New())

	status, _ := doJSON(t, app, fiber.MethodGet, "/contents/not-a-uuid", nil)
	if status != fiber.StatusUnprocessableEntity {
		t.Fatalf("bad uuid status = %d, want %d", status, fiber.StatusUnprocessableEntity)
	}

	status, _ = doJSON(t, app, fiber.MethodGet, "/contents/"+uuid.New().String(), nil)
	if status != fiber.StatusNotFound {
		t.Fatalf("unknown status = %d, want %d", status, fiber.StatusNotFound)
	}
}

func TestHandler_Update(t *testing.T) {
	app := newTestApp(uuid.New())
	id := createContent(t, app, "Editable")

	archived := string(domain.StatusArchived)
	status, body := doJSON(t, app, fiber.MethodPatch, "/contents/"+id, map[string]any{"status": archived})
	if status != fiber.StatusOK {
		t.Fatalf("update status = %d, want %d, body=%v", status, fiber.StatusOK, body)
	}
	data, _ := body["data"].(map[string]any)
	if data["status"] != archived {
		t.Fatalf("status = %v, want archived", data["status"])
	}

	status, _ = doJSON(t, app, fiber.MethodPatch, "/contents/"+id, map[string]any{"status": "published"})
	if status != fiber.StatusUnprocessableEntity {
		t.Fatalf("invalid status update = %d, want %d", status, fiber.StatusUnprocessableEntity)
	}
}

func TestHandler_ArchiveAndDuplicate(t *testing.T) {
	app := newTestApp(uuid.New())
	id := createContent(t, app, "Campaign")

	status, body := doJSON(t, app, fiber.MethodDelete, "/contents/"+id, nil)
	if status != fiber.StatusOK {
		t.Fatalf("archive status = %d, want %d", status, fiber.StatusOK)
	}
	data, _ := body["data"].(map[string]any)
	if data["status"] != string(domain.StatusArchived) {
		t.Fatalf("archive status field = %v, want archived", data["status"])
	}

	status, body = doJSON(t, app, fiber.MethodPost, "/contents/"+id+"/duplicate", nil)
	if status != fiber.StatusCreated {
		t.Fatalf("duplicate status = %d, want %d", status, fiber.StatusCreated)
	}
	data, _ = body["data"].(map[string]any)
	if data["title"] != "Copy of Campaign" || data["status"] != string(domain.StatusDraft) {
		t.Fatalf("duplicate data = %v, unexpected", data)
	}
}

func TestHandler_List(t *testing.T) {
	app := newTestApp(uuid.New())
	createContent(t, app, "one")
	createContent(t, app, "two")

	status, body := doJSON(t, app, fiber.MethodGet, "/contents", nil)
	if status != fiber.StatusOK {
		t.Fatalf("list status = %d, want %d", status, fiber.StatusOK)
	}
	data, _ := body["data"].(map[string]any)
	items, _ := data["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("list items = %d, want 2", len(items))
	}
}
