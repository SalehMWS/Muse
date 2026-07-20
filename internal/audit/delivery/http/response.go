package http

import (
	"time"

	"github.com/SalehMWS/Muse/internal/audit/domain"
)

type EventResponse struct {
	ID            string            `json:"id"`
	Action        string            `json:"action"`
	Result        string            `json:"result"`
	ResourceType  string            `json:"resource_type,omitempty"`
	ResourceID    string            `json:"resource_id,omitempty"`
	IPAddress     string            `json:"ip_address,omitempty"`
	UserAgent     string            `json:"user_agent,omitempty"`
	RequestID     string            `json:"request_id,omitempty"`
	CorrelationID string            `json:"correlation_id,omitempty"`
	TraceID       string            `json:"trace_id,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	CreatedAt     string            `json:"created_at"`
}

func newEventResponse(event domain.Event) EventResponse {
	return EventResponse{
		ID:            event.ID.String(),
		Action:        string(event.Action),
		Result:        string(event.Result),
		ResourceType:  event.ResourceType,
		ResourceID:    event.ResourceID,
		IPAddress:     event.IPAddress,
		UserAgent:     event.UserAgent,
		RequestID:     event.RequestID,
		CorrelationID: event.CorrelationID,
		TraceID:       event.TraceID,
		Metadata:      event.Metadata,
		CreatedAt:     event.CreatedAt.Format(time.RFC3339),
	}
}

func newEventResponses(events []domain.Event) []EventResponse {
	items := make([]EventResponse, 0, len(events))
	for _, event := range events {
		items = append(items, newEventResponse(event))
	}
	return items
}
