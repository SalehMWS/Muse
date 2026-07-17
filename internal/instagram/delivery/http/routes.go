package http

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, h *Handler, requireAuth fiber.Handler) {
	router.Get("/connect", requireAuth, h.Connect)
	router.Get("/callback", h.Callback)
	router.Get("/accounts", requireAuth, h.List)
	router.Post("/accounts/:id/refresh", requireAuth, h.Refresh)
	router.Delete("/accounts/:id", requireAuth, h.Disconnect)
}
