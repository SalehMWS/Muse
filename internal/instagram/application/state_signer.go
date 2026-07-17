package application

import "github.com/google/uuid"

type StateSigner interface {
	Sign(userID uuid.UUID) (string, error)
	Verify(state string) (uuid.UUID, error)
}
