package publishing

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	auditapp "github.com/SalehMWS/Muse/internal/audit/application"
	"github.com/SalehMWS/Muse/internal/publishing/application"
	httpdelivery "github.com/SalehMWS/Muse/internal/publishing/delivery/http"
	"github.com/SalehMWS/Muse/internal/publishing/infrastructure/meta"
	"github.com/SalehMWS/Muse/internal/publishing/infrastructure/postgres"
	"github.com/SalehMWS/Muse/internal/shared/config"
	"github.com/SalehMWS/Muse/internal/shared/metrics"
)

type Module struct {
	Handler *httpdelivery.Handler
	Publish *application.PublishUseCase
}

func New(pool *pgxpool.Pool, cfg config.Instagram, accounts application.AccountReader, contents application.ContentReader, recorder *metrics.Metrics, audit *auditapp.Recorder) *Module {
	var (
		publishingRecorder *metrics.Publishing
		businessRecorder   *metrics.Business
	)
	if recorder != nil {
		publishingRecorder = recorder.Publishing
		businessRecorder = recorder.Business
	}

	repo := postgres.NewPublicationRepository(pool)
	client := meta.NewPublishClient(cfg.GraphBaseURL, cfg.HTTPTimeout, publishingRecorder)

	publishUC := application.NewPublishUseCase(accounts, contents, client, repo, publishingRecorder, businessRecorder, audit)
	listUC := application.NewListPublicationsUseCase(repo)

	return &Module{
		Handler: httpdelivery.NewHandler(publishUC, listUC),
		Publish: publishUC,
	}
}

func (m *Module) RegisterRoutes(router fiber.Router, requireAuth fiber.Handler) {
	group := router.Group("/contents")
	httpdelivery.RegisterRoutes(group, m.Handler, requireAuth)
}
