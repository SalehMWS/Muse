package worker

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	pubapp "github.com/SalehMWS/Muse/internal/publishing/application"
	"github.com/SalehMWS/Muse/internal/worker/application"
	httpdelivery "github.com/SalehMWS/Muse/internal/worker/delivery/http"
	"github.com/SalehMWS/Muse/internal/worker/domain"
	workerpublishing "github.com/SalehMWS/Muse/internal/worker/infrastructure/publishing"
	redisbroker "github.com/SalehMWS/Muse/internal/worker/infrastructure/redis"
)

const (
	jobStream = "novaflow:jobs"
	jobGroup  = "workers"
)

type Module struct {
	Enqueuer application.Enqueuer
	Pool     *application.Pool
	Handler  *httpdelivery.Handler
}

func New(client *goredis.Client, publish *pubapp.PublishUseCase, logger *zap.Logger, workers int) (*Module, error) {
	broker := redisbroker.NewBroker(client, jobStream, jobGroup, "worker-"+uuid.NewString())
	if err := broker.EnsureGroup(context.Background()); err != nil {
		return nil, err
	}

	dispatcher := application.NewDispatcher()
	dispatcher.Register(domain.TypeInstagramPublish, workerpublishing.NewPublishHandler(publish))

	pool := application.NewPool(broker, dispatcher, logger, workers, 2*time.Second)

	return &Module{
		Enqueuer: broker,
		Pool:     pool,
		Handler:  httpdelivery.NewHandler(pool),
	}, nil
}

func (m *Module) RegisterRoutes(router fiber.Router, requireAuth fiber.Handler) {
	httpdelivery.RegisterRoutes(router, m.Handler, requireAuth)
}
