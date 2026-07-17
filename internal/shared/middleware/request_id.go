package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const HeaderRequestID = "X-Request-ID"

const LocalsRequestID = "request_id"

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get(HeaderRequestID)
		if requestID == "" {
			id, err := uuid.NewV7()
			if err != nil {
				id = uuid.New()
			}
			requestID = id.String()
		}

		c.Locals(LocalsRequestID, requestID)
		c.Set(HeaderRequestID, requestID)

		return c.Next()
	}
}

func GetRequestID(c *fiber.Ctx) string {
	requestID, _ := c.Locals(LocalsRequestID).(string)
	return requestID
}
