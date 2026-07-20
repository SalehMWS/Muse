package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/instagram/application"
	httpdelivery "github.com/SalehMWS/Muse/internal/instagram/delivery/http"
	"github.com/SalehMWS/Muse/internal/instagram/domain"
	"github.com/SalehMWS/Muse/internal/shared/authcontext"
)

type deps struct {
	oauth  *fakeOAuthClient
	signer fakeStateSigner
	repo   *fakeAccountRepository
}

func newTestApp(userID uuid.UUID, d deps) *fiber.App {
	connectUC := application.NewConnectUseCase(d.oauth, d.signer)
	callbackUC := application.NewCallbackUseCase(d.oauth, d.signer, fakeTokenCipher{}, d.repo)
	listUC := application.NewListUseCase(d.repo)
	refreshUC := application.NewRefreshUseCase(d.oauth, fakeTokenCipher{}, d.repo)
	disconnectUC := application.NewDisconnectUseCase(d.repo)

	handler := httpdelivery.NewHandler(connectUC, callbackUC, listUC, refreshUC, disconnectUC, nil)

	app := fiber.New()
	requireAuth := func(c *fiber.Ctx) error {
		authcontext.SetUser(c, userID, uuid.New())
		return c.Next()
	}
	httpdelivery.RegisterRoutes(app.Group("/instagram"), handler, requireAuth)
	return app
}

func do(t *testing.T, app *fiber.App, method, path string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequestWithContext(context.Background(), method, path, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var parsed map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &parsed); err != nil {
			t.Fatalf("unmarshal body: %v (%s)", err, raw)
		}
	}
	return resp.StatusCode, parsed
}

func seedHTTPAccount(repo *fakeAccountRepository, userID uuid.UUID) domain.ConnectedAccount {
	accountType := "BUSINESS"
	account := domain.ConnectedAccount{
		ID:              uuid.New(),
		UserID:          userID,
		InstagramUserID: "17841400000000000",
		Username:        "brand",
		AccountType:     &accountType,
		AccessToken:     "enc:token",
		TokenExpiresAt:  time.Now().Add(time.Hour),
		Status:          domain.AccountStatusActive,
	}
	repo.put(account)
	return account
}

func TestHandler_Connect(t *testing.T) {
	userID := uuid.New()
	app := newTestApp(userID, deps{oauth: &fakeOAuthClient{}, signer: fakeStateSigner{userID: userID}, repo: newFakeAccountRepository()})

	status, body := do(t, app, fiber.MethodGet, "/instagram/connect")
	if status != fiber.StatusOK {
		t.Fatalf("status = %d, want %d, body=%v", status, fiber.StatusOK, body)
	}
	data, _ := body["data"].(map[string]any)
	if data["authorization_url"] == "" || data["authorization_url"] == nil {
		t.Fatalf("missing authorization_url: %v", body)
	}
	if data["state"] == "" || data["state"] == nil {
		t.Fatalf("missing state: %v", body)
	}
}

func TestHandler_Callback(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		repo := newFakeAccountRepository()
		d := deps{
			oauth: &fakeOAuthClient{
				exchangeToken: application.Token{AccessToken: "long", ExpiresIn: time.Hour},
				profile:       application.Profile{UserID: "17841400000000000", Username: "brand", AccountType: "BUSINESS"},
			},
			signer: fakeStateSigner{userID: userID},
			repo:   repo,
		}
		app := newTestApp(userID, d)

		status, body := do(t, app, fiber.MethodGet, "/instagram/callback?code=abc&state=xyz")
		if status != fiber.StatusCreated {
			t.Fatalf("status = %d, want %d, body=%v", status, fiber.StatusCreated, body)
		}
		data, _ := body["data"].(map[string]any)
		if data["username"] != "brand" {
			t.Fatalf("username = %v, want brand", data["username"])
		}
		if _, leaked := data["access_token"]; leaked {
			t.Fatal("callback response leaked access_token")
		}
	})

	t.Run("oauth error param", func(t *testing.T) {
		app := newTestApp(userID, deps{oauth: &fakeOAuthClient{}, signer: fakeStateSigner{userID: userID}, repo: newFakeAccountRepository()})
		status, _ := do(t, app, fiber.MethodGet, "/instagram/callback?error=access_denied")
		if status != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want %d", status, fiber.StatusBadRequest)
		}
	})

	t.Run("missing code", func(t *testing.T) {
		app := newTestApp(userID, deps{oauth: &fakeOAuthClient{}, signer: fakeStateSigner{userID: userID}, repo: newFakeAccountRepository()})
		status, _ := do(t, app, fiber.MethodGet, "/instagram/callback?state=xyz")
		if status != fiber.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", status, fiber.StatusUnprocessableEntity)
		}
	})

	t.Run("invalid state", func(t *testing.T) {
		d := deps{oauth: &fakeOAuthClient{}, signer: fakeStateSigner{verifyErr: errors.New("bad")}, repo: newFakeAccountRepository()}
		app := newTestApp(userID, d)
		status, _ := do(t, app, fiber.MethodGet, "/instagram/callback?code=abc&state=xyz")
		if status != fiber.StatusBadRequest {
			t.Fatalf("status = %d, want %d", status, fiber.StatusBadRequest)
		}
	})
}

func TestHandler_List(t *testing.T) {
	userID := uuid.New()
	repo := newFakeAccountRepository()
	seedHTTPAccount(repo, userID)
	app := newTestApp(userID, deps{oauth: &fakeOAuthClient{}, signer: fakeStateSigner{userID: userID}, repo: repo})

	status, body := do(t, app, fiber.MethodGet, "/instagram/accounts")
	if status != fiber.StatusOK {
		t.Fatalf("status = %d, want %d, body=%v", status, fiber.StatusOK, body)
	}
	data, _ := body["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("data length = %d, want 1", len(data))
	}
	item, _ := data[0].(map[string]any)
	if item["status"] != string(domain.AccountStatusActive) {
		t.Fatalf("status = %v, want active", item["status"])
	}
	if _, leaked := item["access_token"]; leaked {
		t.Fatal("list response leaked access_token")
	}
}

func TestHandler_Refresh(t *testing.T) {
	userID := uuid.New()

	t.Run("invalid id", func(t *testing.T) {
		app := newTestApp(userID, deps{oauth: &fakeOAuthClient{}, signer: fakeStateSigner{userID: userID}, repo: newFakeAccountRepository()})
		status, _ := do(t, app, fiber.MethodPost, "/instagram/accounts/not-a-uuid/refresh")
		if status != fiber.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", status, fiber.StatusUnprocessableEntity)
		}
	})

	t.Run("unknown account", func(t *testing.T) {
		app := newTestApp(userID, deps{oauth: &fakeOAuthClient{}, signer: fakeStateSigner{userID: userID}, repo: newFakeAccountRepository()})
		status, _ := do(t, app, fiber.MethodPost, "/instagram/accounts/"+uuid.New().String()+"/refresh")
		if status != fiber.StatusNotFound {
			t.Fatalf("status = %d, want %d", status, fiber.StatusNotFound)
		}
	})
}

func TestHandler_Disconnect(t *testing.T) {
	userID := uuid.New()
	repo := newFakeAccountRepository()
	account := seedHTTPAccount(repo, userID)
	app := newTestApp(userID, deps{oauth: &fakeOAuthClient{}, signer: fakeStateSigner{userID: userID}, repo: repo})

	status, _ := do(t, app, fiber.MethodDelete, "/instagram/accounts/"+account.ID.String())
	if status != fiber.StatusNoContent {
		t.Fatalf("status = %d, want %d", status, fiber.StatusNoContent)
	}
	if len(repo.byID) != 0 {
		t.Fatal("account not deleted")
	}
}
