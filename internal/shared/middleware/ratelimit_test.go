package middleware_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/SalehMWS/Muse/internal/shared/middleware"
	"github.com/SalehMWS/Muse/internal/shared/ratelimit"
	"github.com/SalehMWS/Muse/internal/shared/response"
)

type stubLimiter struct {
	decision ratelimit.Decision
	err      error
	calls    int
}

func (s *stubLimiter) Allow(_ context.Context, _ string, _ int, _ time.Duration) (ratelimit.Decision, error) {
	s.calls++
	if s.err != nil {
		return ratelimit.Decision{}, s.err
	}
	return s.decision, nil
}

func newTestApp(limiter ratelimit.Limiter, rule middleware.RateLimitRule) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return response.Fail(c, err)
		},
	})
	app.Use(middleware.RateLimit(limiter, rule, nil, zap.NewNop()))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
	return app
}

func defaultRule() middleware.RateLimitRule {
	return middleware.RateLimitRule{
		Scope:    "api",
		Limit:    100,
		Window:   time.Minute,
		FailOpen: true,
	}
}

func errorCodeOf(t *testing.T, resp *http.Response) string {
	t.Helper()

	var envelope response.Envelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error == nil {
		t.Fatal("expected error body, got none")
	}
	return envelope.Error.Code
}

func TestRateLimitAllowsAndSetsHeaders(t *testing.T) {
	limiter := &stubLimiter{decision: ratelimit.Decision{Allowed: true, Limit: 100, Remaining: 99}}
	app := newTestApp(limiter, defaultRule())

	resp, err := app.Test(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if got := resp.Header.Get("X-RateLimit-Limit"); got != "100" {
		t.Errorf("X-RateLimit-Limit = %q, want %q", got, "100")
	}
	if got := resp.Header.Get("X-RateLimit-Remaining"); got != "99" {
		t.Errorf("X-RateLimit-Remaining = %q, want %q", got, "99")
	}
	if got := resp.Header.Get("X-RateLimit-Reset"); got != "60" {
		t.Errorf("X-RateLimit-Reset = %q, want %q", got, "60")
	}
}

func TestRateLimitRejectsWhenExhausted(t *testing.T) {
	limiter := &stubLimiter{decision: ratelimit.Decision{
		Allowed:    false,
		Limit:      100,
		Remaining:  0,
		RetryAfter: 1500 * time.Millisecond,
	}}
	app := newTestApp(limiter, defaultRule())

	resp, err := app.Test(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusTooManyRequests)
	}
	if got := resp.Header.Get(fiber.HeaderRetryAfter); got != "2" {
		t.Errorf("Retry-After = %q, want %q", got, "2")
	}
	if got := errorCodeOf(t, resp); got != "RATE_LIMITED" {
		t.Errorf("error code = %q, want %q", got, "RATE_LIMITED")
	}
}

func TestRateLimitFailsOpenWhenBackendErrors(t *testing.T) {
	limiter := &stubLimiter{err: errors.New("redis down")}
	app := newTestApp(limiter, defaultRule())

	resp, err := app.Test(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRateLimitFailsClosedWhenConfigured(t *testing.T) {
	limiter := &stubLimiter{err: errors.New("redis down")}
	rule := defaultRule()
	rule.FailOpen = false
	app := newTestApp(limiter, rule)

	resp, err := app.Test(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusServiceUnavailable)
	}
	if got := errorCodeOf(t, resp); got != "SERVICE_UNAVAILABLE" {
		t.Errorf("error code = %q, want %q", got, "SERVICE_UNAVAILABLE")
	}
}

func TestRateLimitStackedTiersReportTighterRemaining(t *testing.T) {
	strict := &stubLimiter{decision: ratelimit.Decision{Allowed: true, Limit: 10, Remaining: 1}}
	general := &stubLimiter{decision: ratelimit.Decision{Allowed: true, Limit: 100, Remaining: 42}}

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return response.Fail(c, err)
		},
	})
	app.Use(middleware.RateLimit(strict, middleware.RateLimitRule{
		Scope: "auth", Limit: 10, Window: time.Minute, FailOpen: true,
	}, nil, zap.NewNop()))
	app.Use(middleware.RateLimit(general, defaultRule(), nil, zap.NewNop()))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	resp, err := app.Test(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if got := resp.Header.Get("X-RateLimit-Remaining"); got != "1" {
		t.Errorf("X-RateLimit-Remaining = %q, want %q from the stricter tier", got, "1")
	}
	if got := resp.Header.Get("X-RateLimit-Limit"); got != "10" {
		t.Errorf("X-RateLimit-Limit = %q, want %q from the stricter tier", got, "10")
	}
}

func TestRateLimitNilRecorderIsSafe(t *testing.T) {
	limiter := &stubLimiter{decision: ratelimit.Decision{Allowed: true, Limit: 1, Remaining: 0}}
	app := newTestApp(limiter, defaultRule())

	resp, err := app.Test(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if limiter.calls != 1 {
		t.Fatalf("limiter calls = %d, want 1", limiter.calls)
	}
}
