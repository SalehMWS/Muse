package application

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Token struct {
	Value     string
	ExpiresAt time.Time
}

type Claims struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
}

type TokenIssuer interface {
	Issue(ctx context.Context, userID, sessionID uuid.UUID) (Token, error)
	Verify(ctx context.Context, tokenString string) (Claims, error)
}
