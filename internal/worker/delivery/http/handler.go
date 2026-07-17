package http

import (
	"github.com/gofiber/fiber/v2"

	"github.com/SalehMWS/Muse/internal/shared/response"
	"github.com/SalehMWS/Muse/internal/worker/application"
)

type Handler struct {
	pool *application.Pool
}

func NewHandler(pool *application.Pool) *Handler {
	return &Handler{pool: pool}
}

func (h *Handler) Stats(c *fiber.Ctx) error {
	return response.OK(c, h.pool.Stats())
}

func RegisterRoutes(router fiber.Router, h *Handler, requireAuth fiber.Handler) {
	router.Get("/worker/stats", requireAuth, h.Stats)
}
