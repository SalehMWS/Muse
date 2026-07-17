package domain

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusDisabled  Status = "disabled"
	StatusDeleted   Status = "deleted"
)

type User struct {
	ID            uuid.UUID
	Email         Email
	PasswordHash  string
	DisplayName   string
	AvatarURL     *string
	Status        Status
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

func (u User) CanAuthenticate() bool {
	return u.Status == StatusActive
}
