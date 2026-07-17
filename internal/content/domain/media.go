package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type MediaType string

const (
	MediaImage MediaType = "image"
	MediaVideo MediaType = "video"
)

func ValidMediaType(t MediaType) bool {
	return t == MediaImage || t == MediaVideo
}

type Media struct {
	ID        uuid.UUID
	ContentID uuid.UUID
	URL       string
	MediaType MediaType
	Position  int
	CreatedAt time.Time
}

func NewMedia(contentID uuid.UUID, url, mediaType string, position int) (Media, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return Media{}, ErrMediaURLRequired
	}

	mediaValue := MediaType(defaultString(strings.TrimSpace(strings.ToLower(mediaType)), string(MediaImage)))
	if !ValidMediaType(mediaValue) {
		return Media{}, ErrInvalidMediaType
	}

	if position < 0 {
		position = 0
	}

	return Media{
		ID:        uuid.New(),
		ContentID: contentID,
		URL:       url,
		MediaType: mediaValue,
		Position:  position,
	}, nil
}
