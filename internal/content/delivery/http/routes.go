package http

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, h *Handler, requireAuth fiber.Handler) {
	router.Post("/", requireAuth, h.Create)
	router.Get("/", requireAuth, h.List)
	router.Get("/:id", requireAuth, h.Get)
	router.Patch("/:id", requireAuth, h.Update)
	router.Delete("/:id", requireAuth, h.Archive)
	router.Post("/:id/duplicate", requireAuth, h.Duplicate)
}
