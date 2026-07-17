package health

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const pingTimeout = 2 * time.Second

type Checker struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewChecker(db *pgxpool.Pool, redisClient *redis.Client) *Checker {
	return &Checker{db: db, redis: redisClient}
}

type componentStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type readyResponse struct {
	Status     string                     `json:"status"`
	Components map[string]componentStatus `json:"components"`
}

func (c *Checker) RegisterRoutes(app *fiber.App) {
	group := app.Group("/health")
	group.Get("/live", c.live)
	group.Get("/ready", c.ready)
}

func (c *Checker) live(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{"status": "ok"})
}

func (c *Checker) ready(ctx *fiber.Ctx) error {
	timeoutCtx, cancel := context.WithTimeout(ctx.UserContext(), pingTimeout)
	defer cancel()

	components := map[string]componentStatus{
		"postgres": checkComponent(func() error { return c.db.Ping(timeoutCtx) }),
		"redis":    checkComponent(func() error { return c.redis.Ping(timeoutCtx).Err() }),
	}

	overall := "ok"
	httpStatus := fiber.StatusOK
	for _, component := range components {
		if component.Status != "ok" {
			overall = "unavailable"
			httpStatus = fiber.StatusServiceUnavailable
			break
		}
	}

	return ctx.Status(httpStatus).JSON(readyResponse{
		Status:     overall,
		Components: components,
	})
}

func checkComponent(ping func() error) componentStatus {
	if err := ping(); err != nil {
		return componentStatus{Status: "unavailable", Error: err.Error()}
	}
	return componentStatus{Status: "ok"}
}
