package middleware

import (
	"math"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	apperrors "github.com/SalehMWS/Muse/internal/shared/errors"
	"github.com/SalehMWS/Muse/internal/shared/metrics"
	"github.com/SalehMWS/Muse/internal/shared/ratelimit"
)

const (
	headerRateLimitLimit     = "X-RateLimit-Limit"
	headerRateLimitRemaining = "X-RateLimit-Remaining"
	headerRateLimitReset     = "X-RateLimit-Reset"
)

type RateLimitRule struct {
	Scope    string
	Limit    int
	Window   time.Duration
	FailOpen bool
}

func RateLimit(limiter ratelimit.Limiter, rule RateLimitRule, recorder *metrics.RateLimit, log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		decision, err := limiter.Allow(c.UserContext(), rule.Scope+":"+c.IP(), rule.Limit, rule.Window)
		if err != nil {
			recorder.Failed(rule.Scope)
			log.Error("rate limiter unavailable",
				zap.String("scope", rule.Scope),
				zap.Bool("fail_open", rule.FailOpen),
				zap.Error(err),
			)

			if rule.FailOpen {
				return c.Next()
			}
			return apperrors.New(apperrors.CodeUnavailable, "rate limiter unavailable")
		}

		if tighterThanCurrent(c, decision.Remaining) {
			c.Set(headerRateLimitLimit, strconv.Itoa(decision.Limit))
			c.Set(headerRateLimitRemaining, strconv.Itoa(decision.Remaining))
			c.Set(headerRateLimitReset, strconv.Itoa(ceilSeconds(resetWindow(decision, rule))))
		}

		if !decision.Allowed {
			retryAfter := ceilSeconds(decision.RetryAfter)
			c.Set(fiber.HeaderRetryAfter, strconv.Itoa(retryAfter))
			recorder.Limited(rule.Scope)

			return apperrors.New(apperrors.CodeRateLimited, "rate limit exceeded")
		}

		recorder.Allowed(rule.Scope)

		return c.Next()
	}
}

func tighterThanCurrent(c *fiber.Ctx, remaining int) bool {
	current := c.GetRespHeader(headerRateLimitRemaining)
	if current == "" {
		return true
	}

	parsed, err := strconv.Atoi(current)
	if err != nil {
		return true
	}
	return remaining < parsed
}

func resetWindow(decision ratelimit.Decision, rule RateLimitRule) time.Duration {
	if !decision.Allowed {
		return decision.RetryAfter
	}
	return rule.Window
}

func ceilSeconds(d time.Duration) int {
	if d <= 0 {
		return 1
	}
	return int(math.Ceil(d.Seconds()))
}
