package http

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, h *Handler, requireAuth fiber.Handler) {
	router.Post("/", requireAuth, h.Create)
	router.Get("/", requireAuth, h.List)
	router.Get("/:id", requireAuth, h.Get)
	router.Patch("/:id", requireAuth, h.Update)
	router.Delete("/:id", requireAuth, h.Archive)
	router.Post("/:id/duplicate", requireAuth, h.Duplicate)
	router.Post("/:id/caption", requireAuth, h.GenerateCaption)
	router.Post("/:id/media", requireAuth, h.AttachMedia)
	router.Get("/:id/media", requireAuth, h.ListMedia)
	router.Delete("/:id/media/:mediaId", requireAuth, h.DeleteMedia)
}
