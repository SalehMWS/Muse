package bootstrap

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/SalehMWS/Muse/internal/ai"
	"github.com/SalehMWS/Muse/internal/audit"
	"github.com/SalehMWS/Muse/internal/auth"
	"github.com/SalehMWS/Muse/internal/content"
	"github.com/SalehMWS/Muse/internal/instagram"
	"github.com/SalehMWS/Muse/internal/knowledge"
	"github.com/SalehMWS/Muse/internal/publishing"
	pubcontent "github.com/SalehMWS/Muse/internal/publishing/infrastructure/content"
	pubinstagram "github.com/SalehMWS/Muse/internal/publishing/infrastructure/instagram"
	"github.com/SalehMWS/Muse/internal/scheduler"
	"github.com/SalehMWS/Muse/internal/shared/cache"
	"github.com/SalehMWS/Muse/internal/shared/config"
	"github.com/SalehMWS/Muse/internal/shared/database"
	"github.com/SalehMWS/Muse/internal/shared/health"
	applogger "github.com/SalehMWS/Muse/internal/shared/logger"
	"github.com/SalehMWS/Muse/internal/shared/metrics"
	"github.com/SalehMWS/Muse/internal/shared/middleware"
	"github.com/SalehMWS/Muse/internal/shared/ratelimit"
	"github.com/SalehMWS/Muse/internal/shared/response"
	"github.com/SalehMWS/Muse/internal/worker"
)

type Container struct {
	Config         *config.Config
	Logger         *zap.Logger
	DB             *pgxpool.Pool
	Redis          *redis.Client
	App            *fiber.App
	AuthMiddleware fiber.Handler
	Metrics        *metrics.Metrics
	Health         *health.Checker
	Scheduler      *scheduler.Module
	Worker         *worker.Module
	Knowledge      *knowledge.Module
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

	recorder := metrics.New(cfg.App.Name, cfg.App.Version, string(cfg.App.Env))
	if err := recorder.Register(metrics.NewPoolCollector(db)); err != nil {
		_ = redisClient.Close()
		db.Close()
		return nil, fmt.Errorf("bootstrap: register pool collector: %w", err)
	}

	app := fiber.New(fiber.Config{
		AppName:                 cfg.App.Name,
		DisableStartupMessage:   true,
		ErrorHandler:            fiberErrorHandler,
		BodyLimit:               cfg.HTTP.BodyLimit,
		ReadTimeout:             cfg.HTTP.ReadTimeout,
		WriteTimeout:            cfg.HTTP.WriteTimeout,
		IdleTimeout:             cfg.HTTP.IdleTimeout,
		EnableTrustedProxyCheck: len(cfg.HTTP.TrustedProxies) > 0,
		TrustedProxies:          cfg.HTTP.TrustedProxies,
		ProxyHeader:             proxyHeader(cfg.HTTP),
	})

	app.Use(middleware.RequestID())
	app.Use(middleware.Recover(log))
	app.Use(middleware.Metrics(recorder.HTTP))
	app.Use(middleware.RequestLogger(log))
	app.Use(middleware.SecurityHeaders(cfg.Security))

	if cfg.Security.CORSEnabled() {
		app.Use(middleware.CORS(cfg.Security))
	}

	if cfg.RateLimit.Enabled {
		limiter := ratelimit.NewRedisLimiter(redisClient)

		app.Use("/api/v1/auth", middleware.RateLimit(limiter, middleware.RateLimitRule{
			Scope:    "auth",
			Limit:    cfg.RateLimit.AuthRequests,
			Window:   cfg.RateLimit.AuthWindow,
			FailOpen: cfg.RateLimit.FailOpen,
		}, recorder.RateLimit, log))

		app.Use("/api/v1", middleware.RateLimit(limiter, middleware.RateLimitRule{
			Scope:    "api",
			Limit:    cfg.RateLimit.Requests,
			Window:   cfg.RateLimit.Window,
			FailOpen: cfg.RateLimit.FailOpen,
		}, recorder.RateLimit, log))
	}

	if cfg.Observability.MetricsEnabled {
		recorder.RegisterRoutes(app, cfg.Observability.MetricsPath)
	}

	auditModule := audit.New(db, log, recorder)

	authModule := auth.New(db, cfg.JWT, cfg.Argon2, auditModule.Recorder)
	apiV1 := app.Group("/api/v1")
	authModule.RegisterRoutes(apiV1)
	auditModule.RegisterRoutes(apiV1, authModule.Middleware)

	instagramModule, err := instagram.New(db, cfg.Instagram, auditModule.Recorder)
	if err != nil {
		_ = redisClient.Close()
		db.Close()
		return nil, fmt.Errorf("bootstrap: %w", err)
	}
	instagramModule.RegisterRoutes(apiV1, authModule.Middleware)

	aiProvider := ai.NewProvider(cfg.AI, log, recorder.AI)
	contentModule := content.New(db, aiProvider, recorder)
	contentModule.RegisterRoutes(apiV1, authModule.Middleware)

	tokenService, err := instagram.NewTokenService(db, cfg.Instagram)
	if err != nil {
		_ = redisClient.Close()
		db.Close()
		return nil, fmt.Errorf("bootstrap: %w", err)
	}
	publishingModule := publishing.New(
		db,
		cfg.Instagram,
		pubinstagram.NewAccountReader(tokenService),
		pubcontent.NewContentReader(db),
		recorder,
		auditModule.Recorder,
	)
	publishingModule.RegisterRoutes(apiV1, authModule.Middleware)

	workerModule, err := worker.New(
		redisClient,
		publishingModule.Publish,
		log,
		cfg.Worker.Concurrency,
		recorder.Worker,
		cfg.Observability.QueueDepthInterval,
	)
	if err != nil {
		_ = redisClient.Close()
		db.Close()
		return nil, fmt.Errorf("bootstrap: %w", err)
	}
	workerModule.RegisterRoutes(apiV1, authModule.Middleware)

	schedulerModule := scheduler.New(db, workerModule.Enqueuer, log, cfg.Scheduler.PollInterval, recorder)
	schedulerModule.RegisterRoutes(apiV1, authModule.Middleware)

	knowledgeModule, err := knowledge.New(db, cfg.Knowledge, cfg.AI, recorder)
	if err != nil {
		_ = redisClient.Close()
		db.Close()
		return nil, fmt.Errorf("bootstrap: %w", err)
	}
	knowledgeModule.RegisterRoutes(apiV1, authModule.Middleware)

	checker := health.NewChecker(db, redisClient, health.Options{
		Version:     cfg.App.Version,
		VectorStore: cfg.Knowledge.VectorStore,
		Knowledge:   knowledgeModule,
		Queue:       workerModule,
	})
	checker.RegisterRoutes(app)

	return &Container{
		Config:         cfg,
		Logger:         log,
		DB:             db,
		Redis:          redisClient,
		App:            app,
		AuthMiddleware: authModule.Middleware,
		Metrics:        recorder,
		Health:         checker,
		Scheduler:      schedulerModule,
		Worker:         workerModule,
		Knowledge:      knowledgeModule,
	}, nil
}

func (c *Container) Shutdown(ctx context.Context) error {
	if c.Knowledge != nil {
		c.Knowledge.Close()
	}

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

func proxyHeader(cfg config.HTTP) string {
	if len(cfg.TrustedProxies) == 0 {
		return ""
	}
	return fiber.HeaderXForwardedFor
}
