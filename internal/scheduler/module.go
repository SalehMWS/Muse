package scheduler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	pubapp "github.com/SalehMWS/Muse/internal/publishing/application"
	"github.com/SalehMWS/Muse/internal/scheduler/application"
	httpdelivery "github.com/SalehMWS/Muse/internal/scheduler/delivery/http"
	schedcontent "github.com/SalehMWS/Muse/internal/scheduler/infrastructure/content"
	"github.com/SalehMWS/Muse/internal/scheduler/infrastructure/cron"
	"github.com/SalehMWS/Muse/internal/scheduler/infrastructure/postgres"
	schedpublishing "github.com/SalehMWS/Muse/internal/scheduler/infrastructure/publishing"
)

type Module struct {
	Handler *httpdelivery.Handler
	Runner  *application.Runner
}

func New(pool *pgxpool.Pool, publish *pubapp.PublishUseCase, logger *zap.Logger, interval time.Duration) *Module {
	repo := postgres.NewScheduleRepository(pool)
	cronParser := cron.New()
	publisher := schedpublishing.NewPublisher(publish)
	contentChecker := schedcontent.NewContentChecker(pool)

	createUC := application.NewCreateScheduleUseCase(repo, cronParser, contentChecker)
	listUC := application.NewListSchedulesUseCase(repo)
	cancelUC := application.NewCancelScheduleUseCase(repo)
	runner := application.NewRunner(repo, publisher, cronParser, logger, interval, 0)

	return &Module{
		Handler: httpdelivery.NewHandler(createUC, listUC, cancelUC),
		Runner:  runner,
	}
}

func (m *Module) RegisterRoutes(router fiber.Router, requireAuth fiber.Handler) {
	group := router.Group("/contents")
	httpdelivery.RegisterRoutes(group, m.Handler, requireAuth)
}
