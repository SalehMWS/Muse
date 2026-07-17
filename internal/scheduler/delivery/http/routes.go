package http

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, h *Handler, requireAuth fiber.Handler) {
	router.Post("/:id/schedule", requireAuth, h.Create)
	router.Get("/:id/schedules", requireAuth, h.List)
	router.Delete("/:id/schedules/:scheduleId", requireAuth, h.Cancel)
}
