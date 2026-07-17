package redis_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	goredis "github.com/redis/go-redis/v9"

	"github.com/SalehMWS/Muse/internal/shared/config"
	"github.com/SalehMWS/Muse/internal/worker/domain"
	wredis "github.com/SalehMWS/Muse/internal/worker/infrastructure/redis"
)

func repoRoot() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..")
}

func testClient(t *testing.T) *goredis.Client {
	t.Helper()
	_ = godotenv.Load(filepath.Join(repoRoot(), "configs", ".env"))
	cfg, err := config.Load()
	if err != nil {
		t.Skipf("redis integration: load config: %v", err)
	}
	client := goredis.NewClient(&goredis.Options{
		Addr: cfg.Redis.Addr(), Password: cfg.Redis.Password, DB: cfg.Redis.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		t.Skipf("redis integration: skipping, cannot reach redis: %v", err)
	}
	return client
}

func TestBroker_EnqueueReadAck(t *testing.T) {
	client := testClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	stream := "test:jobs:" + uuid.NewString()
	defer client.Del(ctx, stream, stream+":dead")

	broker := wredis.NewBroker(client, stream, "workers", "consumer-1")
	if err := broker.EnsureGroup(ctx); err != nil {
		t.Fatalf("EnsureGroup() unexpected error: %v", err)
	}

	created, err := domain.NewJob(domain.TypeInstagramPublish, domain.PublishPayload{UserID: uuid.New()}, 3)
	if err != nil {
		t.Fatalf("NewJob() unexpected error: %v", err)
	}
	if err := broker.Enqueue(ctx, created); err != nil {
		t.Fatalf("Enqueue() unexpected error: %v", err)
	}

	deliveries, err := broker.Read(ctx, 10, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("Read() unexpected error: %v", err)
	}
	if len(deliveries) != 1 || deliveries[0].Job.ID != created.ID {
		t.Fatalf("Read() = %+v, want the enqueued job", deliveries)
	}

	if err := broker.Ack(ctx, deliveries[0]); err != nil {
		t.Fatalf("Ack() unexpected error: %v", err)
	}

	again, err := broker.Read(ctx, 10, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("Read() second unexpected error: %v", err)
	}
	if len(again) != 0 {
		t.Fatalf("Read() second = %d, want 0 after ack", len(again))
	}
}

func TestBroker_DeadLetter(t *testing.T) {
	client := testClient(t)
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	stream := "test:jobs:" + uuid.NewString()
	defer client.Del(ctx, stream, stream+":dead")

	broker := wredis.NewBroker(client, stream, "workers", "consumer-1")
	if err := broker.EnsureGroup(ctx); err != nil {
		t.Fatalf("EnsureGroup() unexpected error: %v", err)
	}

	job, _ := domain.NewJob(domain.TypeInstagramPublish, domain.PublishPayload{}, 1)
	if err := broker.DeadLetter(ctx, job, "exhausted"); err != nil {
		t.Fatalf("DeadLetter() unexpected error: %v", err)
	}

	length, err := client.XLen(ctx, stream+":dead").Result()
	if err != nil {
		t.Fatalf("XLen() unexpected error: %v", err)
	}
	if length != 1 {
		t.Fatalf("dead-letter stream length = %d, want 1", length)
	}
}
