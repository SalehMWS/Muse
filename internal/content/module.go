package content

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	aiapp "github.com/SalehMWS/Muse/internal/ai/application"
	"github.com/SalehMWS/Muse/internal/content/application"
	httpdelivery "github.com/SalehMWS/Muse/internal/content/delivery/http"
	"github.com/SalehMWS/Muse/internal/content/infrastructure/postgres"
	"github.com/SalehMWS/Muse/internal/shared/metrics"
)

type Module struct {
	Handler *httpdelivery.Handler
}

func New(pool *pgxpool.Pool, aiProvider aiapp.LLMProvider, recorder *metrics.Metrics) *Module {
	var business *metrics.Business
	if recorder != nil {
		business = recorder.Business
	}

	repo := postgres.NewContentRepository(pool)
	mediaRepo := postgres.NewMediaRepository(pool)

	return &Module{
		Handler: httpdelivery.NewHandler(
			application.NewCreateUseCase(repo),
			application.NewGetUseCase(repo),
			application.NewUpdateUseCase(repo),
			application.NewArchiveUseCase(repo),
			application.NewDuplicateUseCase(repo),
			application.NewListUseCase(repo),
			application.NewGenerateCaptionUseCase(repo, aiProvider, business),
			application.NewAttachMediaUseCase(repo, mediaRepo),
			application.NewListMediaUseCase(repo, mediaRepo),
			application.NewDeleteMediaUseCase(repo, mediaRepo),
		),
	}
}

func (m *Module) RegisterRoutes(router fiber.Router, requireAuth fiber.Handler) {
	group := router.Group("/contents")
	httpdelivery.RegisterRoutes(group, m.Handler, requireAuth)
}
