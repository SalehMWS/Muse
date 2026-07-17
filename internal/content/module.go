package content

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SalehMWS/Muse/internal/content/application"
	httpdelivery "github.com/SalehMWS/Muse/internal/content/delivery/http"
	"github.com/SalehMWS/Muse/internal/content/infrastructure/postgres"
)

type Module struct {
	Handler *httpdelivery.Handler
}

func New(pool *pgxpool.Pool) *Module {
	repo := postgres.NewContentRepository(pool)

	return &Module{
		Handler: httpdelivery.NewHandler(
			application.NewCreateUseCase(repo),
			application.NewGetUseCase(repo),
			application.NewUpdateUseCase(repo),
			application.NewArchiveUseCase(repo),
			application.NewDuplicateUseCase(repo),
			application.NewListUseCase(repo),
		),
	}
}

func (m *Module) RegisterRoutes(router fiber.Router, requireAuth fiber.Handler) {
	group := router.Group("/contents")
	httpdelivery.RegisterRoutes(group, m.Handler, requireAuth)
}
