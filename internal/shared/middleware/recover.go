package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	apperrors "github.com/SalehMWS/Muse/internal/shared/errors"
	"github.com/SalehMWS/Muse/internal/shared/response"
)

func Recover(base *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		start := time.Now()

		defer func() {
			if r := recover(); r != nil {
				panicErr, ok := r.(error)
				if !ok {
					panicErr = fmt.Errorf("%v", r)
				}

				base.Error("panic recovered",
					zap.String("request_id", GetRequestID(c)),
					zap.String("method", c.Method()),
					zap.String("path", c.Path()),
					zap.Duration("duration", time.Since(start)),
					zap.Error(panicErr),
					zap.Stack("stack"),
				)

				err = response.Fail(c, apperrors.NewInternal(panicErr))
			}
		}()

		return c.Next()
	}
}
