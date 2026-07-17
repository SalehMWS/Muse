package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/scheduler/application"
	httpdelivery "github.com/SalehMWS/Muse/internal/scheduler/delivery/http"
	"github.com/SalehMWS/Muse/internal/shared/authcontext"
)

func newTestApp(userID uuid.UUID, checker fakeContentChecker) (*fiber.App, *fakeScheduleRepository) {
	repo := newFakeScheduleRepository()
	handler := httpdelivery.NewHandler(
		application.NewCreateScheduleUseCase(repo, fakeCronParser{}, checker),
		application.NewListSchedulesUseCase(repo),
		application.NewCancelScheduleUseCase(repo),
	)

	app := fiber.New()
	requireAuth := func(c *fiber.Ctx) error {
		authcontext.SetUser(c, userID, uuid.New())
		return c.Next()
	}
	httpdelivery.RegisterRoutes(app.Group("/contents"), handler, requireAuth)
	return app, repo
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

func TestHandler_CreateOneTime(t *testing.T) {
	userID := uuid.New()
	app, _ := newTestApp(userID, fakeContentChecker{})

	future := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	status, body := doJSON(t, app, fiber.MethodPost, "/contents/"+uuid.New().String()+"/schedule", map[string]any{
		"instagram_account_id": uuid.New().String(),
		"scheduled_for":        future,
	})
	if status != fiber.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%v", status, fiber.StatusCreated, body)
	}
	data, _ := body["data"].(map[string]any)
	if data["status"] != "scheduled" {
		t.Fatalf("status field = %v, want scheduled", data["status"])
	}
}

func TestHandler_CreateValidation(t *testing.T) {
	userID := uuid.New()
	app, _ := newTestApp(userID, fakeContentChecker{})

	t.Run("missing account", func(t *testing.T) {
		status, _ := doJSON(t, app, fiber.MethodPost, "/contents/"+uuid.New().String()+"/schedule", map[string]any{
			"scheduled_for": time.Now().Add(time.Hour).Format(time.RFC3339),
		})
		if status != fiber.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", status, fiber.StatusUnprocessableEntity)
		}
	})

	t.Run("past time", func(t *testing.T) {
		status, _ := doJSON(t, app, fiber.MethodPost, "/contents/"+uuid.New().String()+"/schedule", map[string]any{
			"instagram_account_id": uuid.New().String(),
			"scheduled_for":        time.Now().Add(-time.Hour).Format(time.RFC3339),
		})
		if status != fiber.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", status, fiber.StatusUnprocessableEntity)
		}
	})
}

func TestHandler_CreateContentNotOwned(t *testing.T) {
	userID := uuid.New()
	app, _ := newTestApp(userID, fakeContentChecker{err: application.ErrContentNotFound})

	status, _ := doJSON(t, app, fiber.MethodPost, "/contents/"+uuid.New().String()+"/schedule", map[string]any{
		"instagram_account_id": uuid.New().String(),
		"scheduled_for":        time.Now().Add(time.Hour).Format(time.RFC3339),
	})
	if status != fiber.StatusNotFound {
		t.Fatalf("status = %d, want %d", status, fiber.StatusNotFound)
	}
}

func TestHandler_ListAndCancel(t *testing.T) {
	userID := uuid.New()
	app, repo := newTestApp(userID, fakeContentChecker{})

	contentID := uuid.New().String()
	_, createBody := doJSON(t, app, fiber.MethodPost, "/contents/"+contentID+"/schedule", map[string]any{
		"instagram_account_id": uuid.New().String(),
		"cron_expression":      "0 12 * * *",
	})
	data, _ := createBody["data"].(map[string]any)
	scheduleID, _ := data["id"].(string)

	status, listBody := doJSON(t, app, fiber.MethodGet, "/contents/"+contentID+"/schedules", nil)
	if status != fiber.StatusOK {
		t.Fatalf("list status = %d, want %d", status, fiber.StatusOK)
	}
	items, _ := listBody["data"].([]any)
	if len(items) != 1 {
		t.Fatalf("schedules = %d, want 1", len(items))
	}

	status, _ = doJSON(t, app, fiber.MethodDelete, "/contents/"+contentID+"/schedules/"+scheduleID, nil)
	if status != fiber.StatusNoContent {
		t.Fatalf("cancel status = %d, want %d", status, fiber.StatusNoContent)
	}
	if len(repo.byID) != 1 {
		t.Fatalf("repo schedules = %d, want 1 (cancelled, not deleted)", len(repo.byID))
	}
}
