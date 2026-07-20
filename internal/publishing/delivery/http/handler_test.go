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

	"github.com/SalehMWS/Muse/internal/publishing/application"
	httpdelivery "github.com/SalehMWS/Muse/internal/publishing/delivery/http"
	"github.com/SalehMWS/Muse/internal/publishing/domain"
	"github.com/SalehMWS/Muse/internal/shared/authcontext"
)

func newTestApp(userID uuid.UUID, accounts application.AccountReader, contents application.ContentReader) *fiber.App {
	repo := &fakePublicationRepository{}
	handler := httpdelivery.NewHandler(
		application.NewPublishUseCase(accounts, contents, stubPublishClient{}, repo, nil, nil),
		application.NewListPublicationsUseCase(repo),
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

func imageContent() application.PublishableContent {
	return application.PublishableContent{
		Caption:     "hi",
		ContentType: "image",
		Media:       []application.MediaItem{{URL: "https://cdn/a.jpg", MediaType: "image"}},
	}
}

func TestHandler_Publish(t *testing.T) {
	userID := uuid.New()
	accounts := fakeAccountReader{account: application.Account{ID: uuid.New(), InstagramUserID: "123", AccessToken: "tok"}}
	app := newTestApp(userID, accounts, fakeContentReader{content: imageContent()})

	status, body := doJSON(t, app, fiber.MethodPost, "/contents/"+uuid.New().String()+"/publish", map[string]any{"instagram_account_id": uuid.New().String()})
	if status != fiber.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%v", status, fiber.StatusCreated, body)
	}
	data, _ := body["data"].(map[string]any)
	if data["status"] != string(domain.StatusPublished) {
		t.Fatalf("status field = %v, want published", data["status"])
	}
}

func TestHandler_PublishMissingAccount(t *testing.T) {
	userID := uuid.New()
	app := newTestApp(userID, fakeAccountReader{}, fakeContentReader{content: imageContent()})

	status, _ := doJSON(t, app, fiber.MethodPost, "/contents/"+uuid.New().String()+"/publish", map[string]any{})
	if status != fiber.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", status, fiber.StatusUnprocessableEntity)
	}
}

func TestHandler_PublishNoMedia(t *testing.T) {
	userID := uuid.New()
	accounts := fakeAccountReader{account: application.Account{ID: uuid.New(), InstagramUserID: "123", AccessToken: "tok"}}
	app := newTestApp(userID, accounts, fakeContentReader{content: application.PublishableContent{ContentType: "image"}})

	status, _ := doJSON(t, app, fiber.MethodPost, "/contents/"+uuid.New().String()+"/publish", map[string]any{"instagram_account_id": uuid.New().String()})
	if status != fiber.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", status, fiber.StatusUnprocessableEntity)
	}
}

func TestHandler_ListPublications(t *testing.T) {
	userID := uuid.New()
	accounts := fakeAccountReader{account: application.Account{ID: uuid.New(), InstagramUserID: "123", AccessToken: "tok"}}
	app := newTestApp(userID, accounts, fakeContentReader{content: imageContent()})

	contentID := uuid.New().String()
	doJSON(t, app, fiber.MethodPost, "/contents/"+contentID+"/publish", map[string]any{"instagram_account_id": uuid.New().String()})

	status, body := doJSON(t, app, fiber.MethodGet, "/contents/"+contentID+"/publications", nil)
	if status != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", status, fiber.StatusOK)
	}
	items, _ := body["data"].([]any)
	if len(items) != 1 {
		t.Fatalf("publications = %d, want 1", len(items))
	}
}
