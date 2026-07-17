package domain

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusPublished Status = "published"
	StatusFailed    Status = "failed"
)

type MediaType string

const (
	MediaImage    MediaType = "image"
	MediaCarousel MediaType = "carousel"
	MediaReel     MediaType = "reel"
)

func ValidMediaType(t MediaType) bool {
	switch t {
	case MediaImage, MediaCarousel, MediaReel:
		return true
	default:
		return false
	}
}

type Publication struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	ContentID          uuid.UUID
	InstagramAccountID uuid.UUID
	Platform           string
	PlatformPostID     *string
	Status             Status
	Permalink          *string
	ResponseJSON       []byte
	PublishedAt        *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
