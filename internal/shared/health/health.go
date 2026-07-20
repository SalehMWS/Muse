package health

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const pingTimeout = 2 * time.Second

type Pinger interface {
	Ping(ctx context.Context) error
}

type Options struct {
	Version     string
	VectorStore string
	Knowledge   Pinger
	Queue       Pinger
}

type Checker struct {
	db      *pgxpool.Pool
	redis   *redis.Client
	opts    Options
	started bool
}

func NewChecker(db *pgxpool.Pool, redisClient *redis.Client, opts Options) *Checker {
	return &Checker{db: db, redis: redisClient, opts: opts}
}

type componentStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type readyResponse struct {
	Status     string                     `json:"status"`
	Version    string                     `json:"version,omitempty"`
	Components map[string]componentStatus `json:"components"`
}

func (c *Checker) MarkStarted() {
	c.started = true
}

func (c *Checker) RegisterRoutes(app *fiber.App) {
	group := app.Group("/health")
	group.Get("/live", c.live)
	group.Get("/ready", c.ready)
	group.Get("/startup", c.startup)
}

func (c *Checker) live(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{"status": "ok", "version": c.opts.Version})
}

func (c *Checker) ready(ctx *fiber.Ctx) error {
	timeoutCtx, cancel := context.WithTimeout(ctx.UserContext(), pingTimeout)
	defer cancel()

	components := map[string]componentStatus{
		"postgres": checkComponent(func() error { return c.db.Ping(timeoutCtx) }),
		"redis":    checkComponent(func() error { return c.redis.Ping(timeoutCtx).Err() }),
	}

	if c.opts.Queue != nil {
		components["queue"] = checkComponent(func() error { return c.opts.Queue.Ping(timeoutCtx) })
	}
	if c.opts.Knowledge != nil && c.opts.VectorStore == "milvus" {
		components["milvus"] = checkComponent(func() error { return c.opts.Knowledge.Ping(timeoutCtx) })
	}

	return c.respond(ctx, components)
}

func (c *Checker) startup(ctx *fiber.Ctx) error {
	timeoutCtx, cancel := context.WithTimeout(ctx.UserContext(), pingTimeout)
	defer cancel()

	components := map[string]componentStatus{
		"configuration": checkComponent(func() error { return nil }),
		"postgres":      checkComponent(func() error { return c.db.Ping(timeoutCtx) }),
		"redis":         checkComponent(func() error { return c.redis.Ping(timeoutCtx).Err() }),
		"dependencies":  {Status: "ok"},
	}

	if !c.started {
		components["dependencies"] = componentStatus{Status: "starting"}
	}

	return c.respond(ctx, components)
}

func (c *Checker) respond(ctx *fiber.Ctx, components map[string]componentStatus) error {
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
		Version:    c.opts.Version,
		Components: components,
	})
}

func checkComponent(ping func() error) componentStatus {
	if err := ping(); err != nil {
		return componentStatus{Status: "unavailable", Error: err.Error()}
	}
	return componentStatus{Status: "ok"}
}
