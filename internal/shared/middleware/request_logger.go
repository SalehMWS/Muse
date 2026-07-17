package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/SalehMWS/Muse/internal/shared/logger"
	"github.com/SalehMWS/Muse/internal/shared/response"
)

func RequestLogger(base *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		requestID := GetRequestID(c)
		scoped := base.With(
			zap.String("request_id", requestID),
			zap.String("module", "http"),
		)

		ctx := logger.WithContext(c.UserContext(), scoped)
		c.SetUserContext(ctx)

		err := c.Next()

		status := c.Response().StatusCode()
		if err != nil {
			status = response.StatusFromError(err)
		}

		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", status),
			zap.Duration("duration", time.Since(start)),
		}

		switch {
		case status >= fiber.StatusInternalServerError:
			base.Error("request completed", fields...)
		case status >= fiber.StatusBadRequest:
			base.Warn("request completed", fields...)
		default:
			base.Info("request completed", fields...)
		}

		return err
	}
}
