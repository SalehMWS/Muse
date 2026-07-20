package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/SalehMWS/Muse/internal/shared/metrics"
	"github.com/SalehMWS/Muse/internal/shared/response"
)

func Metrics(recorder *metrics.HTTP) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		recorder.RequestStarted()

		err := c.Next()

		status := c.Response().StatusCode()
		if err != nil {
			status = response.StatusFromError(err)
		}

		recorder.RequestFinished(
			c.Method(),
			routeLabel(c),
			status,
			len(c.Response().Body()),
			time.Since(start),
		)

		return err
	}
}

func routeLabel(c *fiber.Ctx) string {
	route := c.Route()
	if route == nil || route.Path == "" {
		return "unmatched"
	}
	if route.Path == "/" && c.Path() != "/" {
		return "unmatched"
	}
	return route.Path
}
