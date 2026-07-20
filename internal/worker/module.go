package worker

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	pubapp "github.com/SalehMWS/Muse/internal/publishing/application"
	"github.com/SalehMWS/Muse/internal/shared/metrics"
	"github.com/SalehMWS/Muse/internal/worker/application"
	httpdelivery "github.com/SalehMWS/Muse/internal/worker/delivery/http"
	"github.com/SalehMWS/Muse/internal/worker/domain"
	workerpublishing "github.com/SalehMWS/Muse/internal/worker/infrastructure/publishing"
	redisbroker "github.com/SalehMWS/Muse/internal/worker/infrastructure/redis"
)

const (
	jobStream            = "novaflow:jobs"
	jobGroup             = "workers"
	defaultDepthInterval = 15 * time.Second
)

type Module struct {
	Enqueuer application.Enqueuer
	Pool     *application.Pool
	Handler  *httpdelivery.Handler

	broker   *redisbroker.Broker
	recorder *metrics.Worker
	interval time.Duration
}

func New(client *goredis.Client, publish *pubapp.PublishUseCase, logger *zap.Logger, workers int, recorder *metrics.Worker, depthInterval time.Duration) (*Module, error) {
	broker := redisbroker.NewBroker(client, jobStream, jobGroup, "worker-"+uuid.NewString())
	if err := broker.EnsureGroup(context.Background()); err != nil {
		return nil, err
	}

	dispatcher := application.NewDispatcher()
	dispatcher.Register(domain.TypeInstagramPublish, workerpublishing.NewPublishHandler(publish))

	pool := application.NewPool(broker, dispatcher, logger, application.PoolOptions{
		Workers:  workers,
		Block:    2 * time.Second,
		Queue:    jobStream,
		Recorder: recorder,
	})

	if depthInterval <= 0 {
		depthInterval = defaultDepthInterval
	}

	return &Module{
		Enqueuer: broker,
		Pool:     pool,
		Handler:  httpdelivery.NewHandler(pool),
		broker:   broker,
		recorder: recorder,
		interval: depthInterval,
	}, nil
}

func (m *Module) RegisterRoutes(router fiber.Router, requireAuth fiber.Handler) {
	httpdelivery.RegisterRoutes(router, m.Handler, requireAuth)
}

func (m *Module) Ping(ctx context.Context) error {
	_, err := m.broker.Depth(ctx)
	return err
}

func (m *Module) ReportQueueDepth(ctx context.Context) {
	if m.recorder == nil {
		return
	}

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	m.sampleDepth(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.sampleDepth(ctx)
		}
	}
}

func (m *Module) sampleDepth(ctx context.Context) {
	if depth, err := m.broker.Depth(ctx); err == nil {
		m.recorder.SetQueueDepth(m.broker.Stream(), float64(depth))
	}
	if depth, err := m.broker.DeadLetterDepth(ctx); err == nil {
		m.recorder.SetQueueDepth(m.broker.DeadLetterStream(), float64(depth))
	}
}
