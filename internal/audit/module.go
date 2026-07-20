package audit

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/SalehMWS/Muse/internal/audit/application"
	httpdelivery "github.com/SalehMWS/Muse/internal/audit/delivery/http"
	"github.com/SalehMWS/Muse/internal/audit/infrastructure/postgres"
	"github.com/SalehMWS/Muse/internal/shared/metrics"
)

type Module struct {
	Handler  *httpdelivery.Handler
	Recorder *application.Recorder
}

func New(pool *pgxpool.Pool, log *zap.Logger, recorder *metrics.Metrics) *Module {
	var auditRecorder *metrics.Audit
	if recorder != nil {
		auditRecorder = recorder.Audit
	}

	repo := postgres.NewEventRepository(pool)

	return &Module{
		Handler:  httpdelivery.NewHandler(application.NewListEventsUseCase(repo)),
		Recorder: application.NewRecorder(repo, log, auditRecorder),
	}
}

func (m *Module) RegisterRoutes(router fiber.Router, requireAuth fiber.Handler) {
	httpdelivery.RegisterRoutes(router, m.Handler, requireAuth)
}
