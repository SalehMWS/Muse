package bootstrap

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/SalehMWS/Muse/internal/auth"
	"github.com/SalehMWS/Muse/internal/shared/cache"
	"github.com/SalehMWS/Muse/internal/shared/config"
	"github.com/SalehMWS/Muse/internal/shared/database"
	"github.com/SalehMWS/Muse/internal/shared/health"
	applogger "github.com/SalehMWS/Muse/internal/shared/logger"
	"github.com/SalehMWS/Muse/internal/shared/middleware"
	"github.com/SalehMWS/Muse/internal/shared/response"
)

type Container struct {
	Config         *config.Config
	Logger         *zap.Logger
	DB             *pgxpool.Pool
	Redis          *redis.Client
	App            *fiber.App
	AuthMiddleware fiber.Handler
}

func New(ctx context.Context) (*Container, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("bootstrap: %w", err)
	}

	log, err := applogger.New(string(cfg.App.Env), cfg.Log.Level)
	if err != nil {
		return nil, fmt.Errorf("bootstrap: build logger: %w", err)
	}

	db, err := database.NewPool(ctx, cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("bootstrap: %w", err)
	}

	redisClient, err := cache.NewClient(ctx, cfg.Redis)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("bootstrap: %w", err)
	}

	app := fiber.New(fiber.Config{
		AppName:               cfg.App.Name,
		DisableStartupMessage: true,
		ErrorHandler:          fiberErrorHandler,
	})

	app.Use(middleware.RequestID())
	app.Use(middleware.Recover(log))
	app.Use(middleware.RequestLogger(log))

	health.NewChecker(db, redisClient).RegisterRoutes(app)

	authModule := auth.New(db, cfg.JWT, cfg.Argon2)
	apiV1 := app.Group("/api/v1")
	authModule.RegisterRoutes(apiV1)

	return &Container{
		Config:         cfg,
		Logger:         log,
		DB:             db,
		Redis:          redisClient,
		App:            app,
		AuthMiddleware: authModule.Middleware,
	}, nil
}

func (c *Container) Shutdown(ctx context.Context) error {
	if err := c.Redis.Close(); err != nil {
		c.Logger.Error("shutdown: close redis", zap.Error(err))
	}

	c.DB.Close()

	if err := c.Logger.Sync(); err != nil {
		_ = err
	}

	return nil
}

func fiberErrorHandler(c *fiber.Ctx, err error) error {
	return response.Fail(c, err)
}
