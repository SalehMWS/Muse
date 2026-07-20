package domain

import (
	"time"

	"github.com/google/uuid"
)

type Action string

const (
	ActionUserRegistered        Action = "user.registered"
	ActionUserLoggedIn          Action = "user.logged_in"
	ActionUserLoginFailed       Action = "user.login_failed"
	ActionUserLoggedOut         Action = "user.logged_out"
	ActionInstagramConnected    Action = "instagram.connected"
	ActionInstagramDisconnected Action = "instagram.disconnected"
	ActionContentPublished      Action = "content.published"
)

type Result string

const (
	ResultSuccess Result = "success"
	ResultFailure Result = "failure"
)

type Event struct {
	ID            uuid.UUID
	UserID        *uuid.UUID
	Action        Action
	Result        Result
	ResourceType  string
	ResourceID    string
	IPAddress     string
	UserAgent     string
	RequestID     string
	CorrelationID string
	TraceID       string
	Metadata      map[string]string
	CreatedAt     time.Time
}

func (a Action) Valid() bool {
	switch a {
	case ActionUserRegistered,
		ActionUserLoggedIn,
		ActionUserLoginFailed,
		ActionUserLoggedOut,
		ActionInstagramConnected,
		ActionInstagramDisconnected,
		ActionContentPublished:
		return true
	default:
		return false
	}
}
