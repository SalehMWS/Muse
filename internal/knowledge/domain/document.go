package domain

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusIndexed Status = "indexed"
	StatusFailed  Status = "failed"
)

type Document struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Title      string
	Source     string
	Status     Status
	ChunkCount int
	LastError  *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
