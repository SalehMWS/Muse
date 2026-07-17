package application

import (
	"context"

	"github.com/google/uuid"
)

type Account struct {
	ID              uuid.UUID
	InstagramUserID string
	AccessToken     string
}

type AccountReader interface {
	AccountForUser(ctx context.Context, userID, accountID uuid.UUID) (Account, error)
}

type MediaItem struct {
	URL       string
	MediaType string
}

type PublishableContent struct {
	ContentID   uuid.UUID
	Caption     string
	Hashtags    []string
	ContentType string
	Media       []MediaItem
}

type ContentReader interface {
	ContentForUser(ctx context.Context, userID, contentID uuid.UUID) (PublishableContent, error)
}

type Credential struct {
	InstagramUserID string
	AccessToken     string
}

type PublishedMedia struct {
	ID        string
	Permalink string
}

type PublishClient interface {
	CreateImageContainer(ctx context.Context, cred Credential, imageURL, caption string) (string, error)
	CreateReelContainer(ctx context.Context, cred Credential, videoURL, caption string) (string, error)
	CreateCarouselItem(ctx context.Context, cred Credential, url string, isVideo bool) (string, error)
	CreateCarouselContainer(ctx context.Context, cred Credential, childIDs []string, caption string) (string, error)
	Publish(ctx context.Context, cred Credential, containerID string) (PublishedMedia, error)
}
