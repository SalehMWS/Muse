package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"

	"github.com/SalehMWS/Muse/internal/shared/config"
)

func SecurityHeaders(cfg config.Security) fiber.Handler {
	hstsMaxAge := 0
	if cfg.HSTSEnabled {
		hstsMaxAge = int(cfg.HSTSMaxAge.Seconds())
	}

	return helmet.New(helmet.Config{
		XFrameOptions:         cfg.FrameOptions,
		ReferrerPolicy:        cfg.ReferrerPolicy,
		ContentSecurityPolicy: cfg.ContentSecurityPolicy,
		PermissionPolicy:      cfg.PermissionsPolicy,
		HSTSMaxAge:            hstsMaxAge,
	})
}

func CORS(cfg config.Security) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     strings.Join(cfg.CORSAllowedOrigins, ","),
		AllowMethods:     cfg.CORSAllowedMethods,
		AllowHeaders:     cfg.CORSAllowedHeaders,
		AllowCredentials: cfg.CORSAllowCredentials,
		MaxAge:           int(cfg.CORSMaxAge.Seconds()),
	})
}
