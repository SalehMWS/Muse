package http

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, h *Handler, requireAuth fiber.Handler) {
	router.Post("/register", h.Register)
	router.Post("/login", h.Login)
	router.Post("/refresh", h.Refresh)
	router.Post("/logout", h.Logout)
	router.Get("/me", requireAuth, h.Me)
}
