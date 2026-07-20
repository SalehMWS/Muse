package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/SalehMWS/Muse/internal/audit/domain"
	"github.com/SalehMWS/Muse/internal/shared/database/sqlc"
)

type EventRepository struct {
	queries *sqlc.Queries
}

func NewEventRepository(db sqlc.DBTX) *EventRepository {
	return &EventRepository{queries: sqlc.New(db)}
}

func (r *EventRepository) Append(ctx context.Context, event domain.Event) error {
	metadata, err := encodeMetadata(event.Metadata)
	if err != nil {
		return err
	}

	return r.queries.AppendAuditLog(ctx, sqlc.AppendAuditLogParams{
		ID:            event.ID,
		UserID:        toPgUUID(event.UserID),
		Action:        string(event.Action),
		Result:        string(event.Result),
		ResourceType:  event.ResourceType,
		ResourceID:    event.ResourceID,
		IpAddress:     event.IPAddress,
		UserAgent:     event.UserAgent,
		RequestID:     event.RequestID,
		CorrelationID: event.CorrelationID,
		TraceID:       event.TraceID,
		Metadata:      metadata,
		CreatedAt:     pgtype.Timestamptz{Time: event.CreatedAt, Valid: true},
	})
}

func (r *EventRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Event, error) {
	rows, err := r.queries.ListAuditLogsByUser(ctx, sqlc.ListAuditLogsByUserParams{
		UserID: pgtype.UUID{Bytes: userID, Valid: true},
		Limit:  safeLimit(limit),
	})
	if err != nil {
		return nil, err
	}

	events := make([]domain.Event, 0, len(rows))
	for _, row := range rows {
		event, err := toDomainEvent(row)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func toDomainEvent(row sqlc.AuditLog) (domain.Event, error) {
	metadata, err := decodeMetadata(row.Metadata)
	if err != nil {
		return domain.Event{}, err
	}

	event := domain.Event{
		ID:            row.ID,
		Action:        domain.Action(row.Action),
		Result:        domain.Result(row.Result),
		ResourceType:  row.ResourceType,
		ResourceID:    row.ResourceID,
		IPAddress:     row.IpAddress,
		UserAgent:     row.UserAgent,
		RequestID:     row.RequestID,
		CorrelationID: row.CorrelationID,
		TraceID:       row.TraceID,
		Metadata:      metadata,
		CreatedAt:     row.CreatedAt.Time,
	}

	if row.UserID.Valid {
		id := uuid.UUID(row.UserID.Bytes)
		event.UserID = &id
	}

	return event, nil
}

func safeLimit(limit int) int32 {
	const maxLimit = 200
	if limit <= 0 || limit > maxLimit {
		return maxLimit
	}
	return int32(limit)
}

func toPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func encodeMetadata(metadata map[string]string) ([]byte, error) {
	if len(metadata) == 0 {
		return []byte("{}"), nil
	}

	encoded, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("audit: encode metadata: %w", err)
	}
	return encoded, nil
}

func decodeMetadata(raw []byte) (map[string]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	metadata := map[string]string{}
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return nil, fmt.Errorf("audit: decode metadata: %w", err)
	}
	if len(metadata) == 0 {
		return nil, nil
	}
	return metadata, nil
}
