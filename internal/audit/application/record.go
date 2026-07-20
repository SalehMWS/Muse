package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/SalehMWS/Muse/internal/audit/domain"
	"github.com/SalehMWS/Muse/internal/shared/metrics"
	"github.com/SalehMWS/Muse/internal/shared/tracing"
)

type Entry struct {
	UserID       *uuid.UUID
	Action       domain.Action
	Result       domain.Result
	ResourceType string
	ResourceID   string
	IPAddress    string
	UserAgent    string
	Metadata     map[string]string
}

type Recorder struct {
	events   EventRepository
	log      *zap.Logger
	recorder *metrics.Audit
}

func NewRecorder(events EventRepository, log *zap.Logger, recorder *metrics.Audit) *Recorder {
	return &Recorder{events: events, log: log, recorder: recorder}
}

func (r *Recorder) Record(ctx context.Context, entry Entry) {
	if r == nil || r.events == nil {
		return
	}

	ids := tracing.FromContext(ctx)

	event := domain.Event{
		ID:            newEventID(),
		UserID:        entry.UserID,
		Action:        entry.Action,
		Result:        entry.Result,
		ResourceType:  entry.ResourceType,
		ResourceID:    entry.ResourceID,
		IPAddress:     entry.IPAddress,
		UserAgent:     entry.UserAgent,
		RequestID:     ids.RequestID,
		CorrelationID: ids.CorrelationID,
		TraceID:       ids.TraceID,
		Metadata:      entry.Metadata,
		CreatedAt:     time.Now().UTC(),
	}

	if err := r.events.Append(ctx, event); err != nil {
		r.recorder.WriteFailed(string(entry.Action))
		r.log.Error("audit: append event",
			zap.String("action", string(entry.Action)),
			zap.String("result", string(entry.Result)),
			zap.String("request_id", ids.RequestID),
			zap.String("trace_id", ids.TraceID),
			zap.Error(err),
		)
		return
	}

	r.recorder.Recorded(string(entry.Action), string(entry.Result))
}

func newEventID() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}
