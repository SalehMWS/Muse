package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SalehMWS/Muse/internal/auth/application"
	httpdelivery "github.com/SalehMWS/Muse/internal/auth/delivery/http"
	"github.com/SalehMWS/Muse/internal/auth/infrastructure/postgres"
	"github.com/SalehMWS/Muse/internal/auth/infrastructure/security"
	"github.com/SalehMWS/Muse/internal/shared/config"
)

type Module struct {
	Handler    *httpdelivery.Handler
	Middleware fiber.Handler
}

func New(pool *pgxpool.Pool, jwtCfg config.JWT, argonCfg config.Argon2) *Module {
	users := postgres.NewUserRepository(pool)
	sessions := postgres.NewSessionRepository(pool)
	hasher := security.NewArgon2Hasher(argonCfg)
	issuer := security.NewJWTIssuer(jwtCfg)

	registerUC := application.NewRegisterUseCase(users, hasher)
	loginUC := application.NewLoginUseCase(users, sessions, hasher, issuer, jwtCfg.RefreshTokenTTL)
	refreshUC := application.NewRefreshUseCase(sessions, issuer, jwtCfg.RefreshTokenTTL)
	logoutUC := application.NewLogoutUseCase(sessions)
	getCurrentUserUC := application.NewGetCurrentUserUseCase(users)

	return &Module{
		Handler:    httpdelivery.NewHandler(registerUC, loginUC, refreshUC, logoutUC, getCurrentUserUC),
		Middleware: httpdelivery.RequireAuth(issuer),
	}
}

func (m *Module) RegisterRoutes(router fiber.Router) {
	group := router.Group("/auth")
	httpdelivery.RegisterRoutes(group, m.Handler, m.Middleware)
}
