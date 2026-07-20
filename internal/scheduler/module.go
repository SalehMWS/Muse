package scheduler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/SalehMWS/Muse/internal/scheduler/application"
	httpdelivery "github.com/SalehMWS/Muse/internal/scheduler/delivery/http"
	schedcontent "github.com/SalehMWS/Muse/internal/scheduler/infrastructure/content"
	"github.com/SalehMWS/Muse/internal/scheduler/infrastructure/cron"
	"github.com/SalehMWS/Muse/internal/scheduler/infrastructure/postgres"
	schedqueue "github.com/SalehMWS/Muse/internal/scheduler/infrastructure/queue"
	"github.com/SalehMWS/Muse/internal/shared/metrics"
	workerapp "github.com/SalehMWS/Muse/internal/worker/application"
)

type Module struct {
	Handler *httpdelivery.Handler
	Runner  *application.Runner
}

func New(pool *pgxpool.Pool, enqueuer workerapp.Enqueuer, logger *zap.Logger, interval time.Duration, recorder *metrics.Metrics) *Module {
	var (
		schedulerRecorder *metrics.Scheduler
		businessRecorder  *metrics.Business
	)
	if recorder != nil {
		schedulerRecorder = recorder.Scheduler
		businessRecorder = recorder.Business
	}

	repo := postgres.NewScheduleRepository(pool)
	cronParser := cron.New()
	publisher := schedqueue.NewPublisher(enqueuer, 3)
	contentChecker := schedcontent.NewContentChecker(pool)

	createUC := application.NewCreateScheduleUseCase(repo, cronParser, contentChecker, businessRecorder)
	listUC := application.NewListSchedulesUseCase(repo)
	cancelUC := application.NewCancelScheduleUseCase(repo)
	runner := application.NewRunner(repo, publisher, cronParser, logger, interval, 0, schedulerRecorder)

	return &Module{
		Handler: httpdelivery.NewHandler(createUC, listUC, cancelUC),
		Runner:  runner,
	}
}

func (m *Module) RegisterRoutes(router fiber.Router, requireAuth fiber.Handler) {
	group := router.Group("/contents")
	httpdelivery.RegisterRoutes(group, m.Handler, requireAuth)
}
