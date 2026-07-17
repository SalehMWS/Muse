package http

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, h *Handler, requireAuth fiber.Handler) {
	router.Post("/:id/publish", requireAuth, h.Publish)
	router.Get("/:id/publications", requireAuth, h.ListPublications)
}
