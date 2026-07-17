package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

const refreshTokenBytes = 32

type Session struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RefreshTokenHash string
	Device           *string
	IPAddress        *string
	UserAgent        *string
	CreatedAt        time.Time
	ExpiresAt        time.Time
	LastActivityAt   time.Time
}

func GenerateRefreshToken() (string, error) {
	buf := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (s Session) IsExpired(now time.Time) bool {
	return !now.Before(s.ExpiresAt)
}
