package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/shared/tracing"
)

const HeaderRequestID = tracing.HeaderRequestID

const (
	LocalsRequestID     = "request_id"
	LocalsCorrelationID = "correlation_id"
	LocalsTraceID       = "trace_id"
	LocalsSpanID        = "span_id"
)

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ids := resolveIDs(c)

		c.Locals(LocalsRequestID, ids.RequestID)
		c.Locals(LocalsCorrelationID, ids.CorrelationID)
		c.Locals(LocalsTraceID, ids.TraceID)
		c.Locals(LocalsSpanID, ids.SpanID)

		c.Set(tracing.HeaderRequestID, ids.RequestID)
		c.Set(tracing.HeaderCorrelationID, ids.CorrelationID)
		c.Set(tracing.HeaderTraceID, ids.TraceID)

		c.SetUserContext(tracing.WithIDs(c.UserContext(), ids))

		return c.Next()
	}
}

func resolveIDs(c *fiber.Ctx) tracing.IDs {
	requestID := c.Get(tracing.HeaderRequestID)
	if requestID == "" {
		id, err := uuid.NewV7()
		if err != nil {
			id = uuid.New()
		}
		requestID = id.String()
	}

	correlationID := c.Get(tracing.HeaderCorrelationID)
	if correlationID == "" {
		correlationID = requestID
	}

	traceID := c.Get(tracing.HeaderTraceID)
	if parsed, _, ok := tracing.ParseTraceparent(c.Get(tracing.HeaderTraceparent)); ok {
		traceID = parsed
	}
	if traceID == "" {
		traceID = tracing.NewTraceID()
	}

	return tracing.IDs{
		RequestID:     requestID,
		CorrelationID: correlationID,
		TraceID:       traceID,
		SpanID:        tracing.NewSpanID(),
	}
}

func GetRequestID(c *fiber.Ctx) string {
	requestID, _ := c.Locals(LocalsRequestID).(string)
	return requestID
}

func GetIDs(c *fiber.Ctx) tracing.IDs {
	return tracing.FromContext(c.UserContext())
}
