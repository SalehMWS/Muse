package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"

	"github.com/SalehMWS/Muse/internal/shared/config"
	"github.com/SalehMWS/Muse/internal/shared/ratelimit"
)

func testClient(t *testing.T) *goredis.Client {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Skipf("ratelimit integration: load config: %v", err)
	}

	client := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		t.Skipf("ratelimit integration: skipping, cannot reach redis: %v", err)
	}

	t.Cleanup(func() {
		_ = client.Close()
	})

	return client
}

func TestRedisLimiterAllowsUpToLimitThenRejects(t *testing.T) {
	client := testClient(t)
	limiter := ratelimit.NewRedisLimiter(client)
	key := "test:" + uuid.NewString()
	ctx := context.Background()

	for i := range 3 {
		decision, err := limiter.Allow(ctx, key, 3, time.Minute)
		if err != nil {
			t.Fatalf("allow %d: %v", i, err)
		}
		if !decision.Allowed {
			t.Fatalf("request %d denied, want allowed", i)
		}
		if want := 2 - i; decision.Remaining != want {
			t.Errorf("request %d remaining = %d, want %d", i, decision.Remaining, want)
		}
	}

	decision, err := limiter.Allow(ctx, key, 3, time.Minute)
	if err != nil {
		t.Fatalf("allow over limit: %v", err)
	}
	if decision.Allowed {
		t.Fatal("fourth request allowed, want denied")
	}
	if decision.RetryAfter <= 0 {
		t.Errorf("RetryAfter = %v, want positive", decision.RetryAfter)
	}
}

func TestRedisLimiterWindowExpires(t *testing.T) {
	client := testClient(t)
	limiter := ratelimit.NewRedisLimiter(client)
	key := "test:" + uuid.NewString()
	ctx := context.Background()

	if _, err := limiter.Allow(ctx, key, 1, 300*time.Millisecond); err != nil {
		t.Fatalf("first allow: %v", err)
	}

	denied, err := limiter.Allow(ctx, key, 1, 300*time.Millisecond)
	if err != nil {
		t.Fatalf("second allow: %v", err)
	}
	if denied.Allowed {
		t.Fatal("second request allowed inside window, want denied")
	}

	time.Sleep(400 * time.Millisecond)

	recovered, err := limiter.Allow(ctx, key, 1, 300*time.Millisecond)
	if err != nil {
		t.Fatalf("third allow: %v", err)
	}
	if !recovered.Allowed {
		t.Fatal("request after window denied, want allowed")
	}
}

func TestRedisLimiterKeysAreIsolated(t *testing.T) {
	client := testClient(t)
	limiter := ratelimit.NewRedisLimiter(client)
	ctx := context.Background()

	first := "test:" + uuid.NewString()
	second := "test:" + uuid.NewString()

	if _, err := limiter.Allow(ctx, first, 1, time.Minute); err != nil {
		t.Fatalf("allow first key: %v", err)
	}

	decision, err := limiter.Allow(ctx, second, 1, time.Minute)
	if err != nil {
		t.Fatalf("allow second key: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("second key denied, want allowed")
	}
}
