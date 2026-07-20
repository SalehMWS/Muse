package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/SalehMWS/Muse/internal/shared/config"
	"github.com/SalehMWS/Muse/internal/shared/middleware"
)

func securityConfig() config.Security {
	return config.Security{
		CORSAllowedOrigins:    []string{"https://app.novaflow.dev"},
		CORSAllowedMethods:    "GET,POST",
		CORSAllowedHeaders:    "Authorization,Content-Type",
		CORSMaxAge:            12 * time.Hour,
		HSTSEnabled:           true,
		HSTSMaxAge:            365 * 24 * time.Hour,
		ContentSecurityPolicy: "default-src 'none'; frame-ancestors 'none'",
		FrameOptions:          "DENY",
		ReferrerPolicy:        "no-referrer",
		PermissionsPolicy:     "geolocation=(), microphone=(), camera=()",
	}
}

func TestSecurityHeadersAreSet(t *testing.T) {
	cfg := securityConfig()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(middleware.SecurityHeaders(cfg))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	resp, err := app.Test(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	want := map[string]string{
		fiber.HeaderXFrameOptions:         "DENY",
		fiber.HeaderXContentTypeOptions:   "nosniff",
		fiber.HeaderReferrerPolicy:        "no-referrer",
		fiber.HeaderContentSecurityPolicy: "default-src 'none'; frame-ancestors 'none'",
		fiber.HeaderPermissionsPolicy:     "geolocation=(), microphone=(), camera=()",
	}

	for header, expected := range want {
		if got := resp.Header.Get(header); got != expected {
			t.Errorf("%s = %q, want %q", header, got, expected)
		}
	}

	if got := resp.Header.Get(fiber.HeaderStrictTransportSecurity); got != "" {
		t.Errorf("Strict-Transport-Security = %q over plain HTTP, want empty", got)
	}
}

func TestSecurityHeadersSetHSTSOverTLS(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage:   true,
		EnableTrustedProxyCheck: true,
		TrustedProxies:          []string{"0.0.0.0/0"},
	})
	app.Use(middleware.SecurityHeaders(securityConfig()))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	req.Header.Set(fiber.HeaderXForwardedProto, "https")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if got := resp.Header.Get(fiber.HeaderStrictTransportSecurity); got != "max-age=31536000; includeSubDomains" {
		t.Errorf("Strict-Transport-Security = %q, want %q", got, "max-age=31536000; includeSubDomains")
	}
}

func TestSecurityHeadersOmitHSTSWhenDisabled(t *testing.T) {
	cfg := securityConfig()
	cfg.HSTSEnabled = false

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(middleware.SecurityHeaders(cfg))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	resp, err := app.Test(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if got := resp.Header.Get(fiber.HeaderStrictTransportSecurity); got != "" {
		t.Errorf("Strict-Transport-Security = %q, want empty", got)
	}
}

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(middleware.CORS(securityConfig()))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	req.Header.Set(fiber.HeaderOrigin, "https://app.novaflow.dev")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if got := resp.Header.Get(fiber.HeaderAccessControlAllowOrigin); got != "https://app.novaflow.dev" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "https://app.novaflow.dev")
	}
}

func TestCORSRejectsUnknownOrigin(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(middleware.CORS(securityConfig()))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	req.Header.Set(fiber.HeaderOrigin, "https://evil.example")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if got := resp.Header.Get(fiber.HeaderAccessControlAllowOrigin); got == "https://evil.example" {
		t.Error("Access-Control-Allow-Origin echoed an origin outside the allowlist")
	}
}
