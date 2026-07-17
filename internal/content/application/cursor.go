package application

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func encodeCursor(createdAt time.Time, id uuid.UUID) string {
	raw := fmt.Sprintf("%d|%s", createdAt.UnixNano(), id.String())
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(cursor string) (time.Time, uuid.UUID, error) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, uuid.Nil, ErrInvalidCursor
	}

	createdRaw, idRaw, ok := strings.Cut(string(raw), "|")
	if !ok {
		return time.Time{}, uuid.Nil, ErrInvalidCursor
	}

	nanos, err := strconv.ParseInt(createdRaw, 10, 64)
	if err != nil {
		return time.Time{}, uuid.Nil, ErrInvalidCursor
	}

	id, err := uuid.Parse(idRaw)
	if err != nil {
		return time.Time{}, uuid.Nil, ErrInvalidCursor
	}

	return time.Unix(0, nanos), id, nil
}
