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

	"github.com/SalehMWS/Muse/internal/auth/application"
	httpdelivery "github.com/SalehMWS/Muse/internal/auth/delivery/http"
	"github.com/SalehMWS/Muse/internal/auth/domain"
)

func newTestApp() (*fiber.App, *fakeUserRepository) {
	users := newFakeUserRepository()
	sessions := newFakeSessionRepository()
	hasher := fakePasswordHasher{}
	issuer := fakeTokenIssuer{}

	registerUC := application.NewRegisterUseCase(users, hasher)
	loginUC := application.NewLoginUseCase(users, sessions, hasher, issuer, 30*24*time.Hour)
	refreshUC := application.NewRefreshUseCase(sessions, issuer, 30*24*time.Hour)
	logoutUC := application.NewLogoutUseCase(sessions)
	meUC := application.NewGetCurrentUserUseCase(users)

	handler := httpdelivery.NewHandler(registerUC, loginUC, refreshUC, logoutUC, meUC)

	app := fiber.New()
	httpdelivery.RegisterRoutes(app, handler, httpdelivery.RequireAuth(issuer))

	return app, users
}

func doJSON(t *testing.T, app *fiber.App, method, path string, body any, headers map[string]string) (int, map[string]any) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequestWithContext(context.Background(), method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	var parsed map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &parsed); err != nil {
			t.Fatalf("unmarshal response body: %v (%s)", err, raw)
		}
	}

	return resp.StatusCode, parsed
}

func TestHandler_Register(t *testing.T) {
	app, _ := newTestApp()

	t.Run("success", func(t *testing.T) {
		status, body := doJSON(t, app, fiber.MethodPost, "/register", map[string]string{
			"email": "new@example.com", "password": "Str0ng!Passw0rd", "display_name": "New User",
		}, nil)
		if status != fiber.StatusCreated {
			t.Fatalf("status = %d, want %d, body=%v", status, fiber.StatusCreated, body)
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		status, body := doJSON(t, app, fiber.MethodPost, "/register", map[string]string{
			"email": "new@example.com", "password": "An0ther!Passw0rd", "display_name": "Dup User",
		}, nil)
		if status != fiber.StatusConflict {
			t.Fatalf("status = %d, want %d, body=%v", status, fiber.StatusConflict, body)
		}
	})

	t.Run("missing fields", func(t *testing.T) {
		status, _ := doJSON(t, app, fiber.MethodPost, "/register", map[string]string{"email": "onlyemail@example.com"}, nil)
		if status != fiber.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", status, fiber.StatusUnprocessableEntity)
		}
	})
}

func TestHandler_LoginAndMe(t *testing.T) {
	app, users := newTestApp()
	userID, _ := uuid.NewV7()
	email, _ := domain.NewEmail("user@example.com")
	users.put(domain.User{ID: userID, Email: email, PasswordHash: "hashed:Str0ng!Passw0rd", Status: domain.StatusActive})

	t.Run("me without token is unauthorized", func(t *testing.T) {
		status, _ := doJSON(t, app, fiber.MethodGet, "/me", nil, nil)
		if status != fiber.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", status, fiber.StatusUnauthorized)
		}
	})

	t.Run("wrong password is unauthorized", func(t *testing.T) {
		status, _ := doJSON(t, app, fiber.MethodPost, "/login", map[string]string{
			"email": "user@example.com", "password": "wrong-password",
		}, nil)
		if status != fiber.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", status, fiber.StatusUnauthorized)
		}
	})

	t.Run("login then me", func(t *testing.T) {
		status, body := doJSON(t, app, fiber.MethodPost, "/login", map[string]string{
			"email": "user@example.com", "password": "Str0ng!Passw0rd",
		}, nil)
		if status != fiber.StatusOK {
			t.Fatalf("login status = %d, want %d, body=%v", status, fiber.StatusOK, body)
		}

		data, _ := body["data"].(map[string]any)
		accessToken, _ := data["access_token"].(string)
		if accessToken == "" {
			t.Fatalf("login response missing access_token: %v", body)
		}

		meStatus, meBody := doJSON(t, app, fiber.MethodGet, "/me", nil, map[string]string{
			"Authorization": "Bearer " + accessToken,
		})
		if meStatus != fiber.StatusOK {
			t.Fatalf("me status = %d, want %d, body=%v", meStatus, fiber.StatusOK, meBody)
		}
	})
}

func TestHandler_RefreshAndLogout(t *testing.T) {
	app, users := newTestApp()
	userID, _ := uuid.NewV7()
	email, _ := domain.NewEmail("refresh@example.com")
	users.put(domain.User{ID: userID, Email: email, PasswordHash: "hashed:Str0ng!Passw0rd", Status: domain.StatusActive})

	loginStatus, loginBody := doJSON(t, app, fiber.MethodPost, "/login", map[string]string{
		"email": "refresh@example.com", "password": "Str0ng!Passw0rd",
	}, nil)
	if loginStatus != fiber.StatusOK {
		t.Fatalf("login status = %d, body=%v", loginStatus, loginBody)
	}
	loginData, _ := loginBody["data"].(map[string]any)
	refreshToken, _ := loginData["refresh_token"].(string)

	t.Run("refresh rotates the token", func(t *testing.T) {
		status, body := doJSON(t, app, fiber.MethodPost, "/refresh", map[string]string{"refresh_token": refreshToken}, nil)
		if status != fiber.StatusOK {
			t.Fatalf("status = %d, want %d, body=%v", status, fiber.StatusOK, body)
		}
		data, _ := body["data"].(map[string]any)
		newRefreshToken, _ := data["refresh_token"].(string)
		if newRefreshToken == "" || newRefreshToken == refreshToken {
			t.Fatalf("refresh did not rotate the token: %v", body)
		}
		refreshToken = newRefreshToken
	})

	t.Run("logout then refresh fails", func(t *testing.T) {
		logoutStatus, _ := doJSON(t, app, fiber.MethodPost, "/logout", map[string]string{"refresh_token": refreshToken}, nil)
		if logoutStatus != fiber.StatusNoContent {
			t.Fatalf("logout status = %d, want %d", logoutStatus, fiber.StatusNoContent)
		}

		status, _ := doJSON(t, app, fiber.MethodPost, "/refresh", map[string]string{"refresh_token": refreshToken}, nil)
		if status != fiber.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", status, fiber.StatusUnauthorized)
		}
	})
}
