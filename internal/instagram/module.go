package instagram

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SalehMWS/Muse/internal/instagram/application"
	httpdelivery "github.com/SalehMWS/Muse/internal/instagram/delivery/http"
	"github.com/SalehMWS/Muse/internal/instagram/infrastructure/meta"
	"github.com/SalehMWS/Muse/internal/instagram/infrastructure/postgres"
	"github.com/SalehMWS/Muse/internal/instagram/infrastructure/security"
	"github.com/SalehMWS/Muse/internal/shared/config"
)

type Module struct {
	Handler *httpdelivery.Handler
}

func New(pool *pgxpool.Pool, cfg config.Instagram) (*Module, error) {
	repo := postgres.NewAccountRepository(pool)
	client := meta.NewOAuthClient(cfg)

	cipher, err := security.NewAESTokenCipher(cfg.TokenEncryptionKey)
	if err != nil {
		return nil, err
	}
	signer := security.NewHMACStateSigner(cfg.StateSecret, cfg.StateTTL)

	connectUC := application.NewConnectUseCase(client, signer)
	callbackUC := application.NewCallbackUseCase(client, signer, cipher, repo)
	listUC := application.NewListUseCase(repo)
	refreshUC := application.NewRefreshUseCase(client, cipher, repo)
	disconnectUC := application.NewDisconnectUseCase(repo)

	return &Module{
		Handler: httpdelivery.NewHandler(connectUC, callbackUC, listUC, refreshUC, disconnectUC),
	}, nil
}

func (m *Module) RegisterRoutes(router fiber.Router, requireAuth fiber.Handler) {
	group := router.Group("/instagram")
	httpdelivery.RegisterRoutes(group, m.Handler, requireAuth)
}
