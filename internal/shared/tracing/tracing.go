package tracing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"

	"go.uber.org/zap"
)

const (
	HeaderRequestID     = "X-Request-ID"
	HeaderCorrelationID = "X-Correlation-ID"
	HeaderTraceID       = "X-Trace-ID"
	HeaderTraceparent   = "traceparent"
)

type contextKey struct{}

var idsContextKey = contextKey{}

type IDs struct {
	RequestID     string
	CorrelationID string
	TraceID       string
	SpanID        string
}

func (i IDs) Fields() []zap.Field {
	return []zap.Field{
		zap.String("request_id", i.RequestID),
		zap.String("correlation_id", i.CorrelationID),
		zap.String("trace_id", i.TraceID),
		zap.String("span_id", i.SpanID),
	}
}

func (i IDs) Traceparent() string {
	if i.TraceID == "" || i.SpanID == "" {
		return ""
	}
	return "00-" + i.TraceID + "-" + i.SpanID + "-01"
}

func NewTraceID() string {
	return randomHex(16)
}

func NewSpanID() string {
	return randomHex(8)
}

func ParseTraceparent(header string) (traceID, parentSpanID string, ok bool) {
	parts := strings.Split(strings.TrimSpace(header), "-")
	if len(parts) != 4 {
		return "", "", false
	}
	if !isHex(parts[1], 32) || !isHex(parts[2], 16) {
		return "", "", false
	}
	if strings.Trim(parts[1], "0") == "" || strings.Trim(parts[2], "0") == "" {
		return "", "", false
	}
	return parts[1], parts[2], true
}

func WithIDs(ctx context.Context, ids IDs) context.Context {
	return context.WithValue(ctx, idsContextKey, ids)
}

func FromContext(ctx context.Context) IDs {
	ids, _ := ctx.Value(idsContextKey).(IDs)
	return ids
}

func randomHex(size int) string {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return strings.Repeat("0", size*2)
	}
	return hex.EncodeToString(buf)
}

func isHex(value string, length int) bool {
	if len(value) != length {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
