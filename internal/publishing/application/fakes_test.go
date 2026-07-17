package application_test

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/publishing/application"
	"github.com/SalehMWS/Muse/internal/publishing/domain"
)

type fakeAccountReader struct {
	account application.Account
	err     error
}

func (f fakeAccountReader) AccountForUser(_ context.Context, _, _ uuid.UUID) (application.Account, error) {
	if f.err != nil {
		return application.Account{}, f.err
	}
	return f.account, nil
}

type fakeContentReader struct {
	content application.PublishableContent
	err     error
}

func (f fakeContentReader) ContentForUser(_ context.Context, _, _ uuid.UUID) (application.PublishableContent, error) {
	if f.err != nil {
		return application.PublishableContent{}, f.err
	}
	return f.content, nil
}

type recordingPublishClient struct {
	failStage    string
	imageCalls   int
	reelCalls    int
	itemCalls    int
	carousel     int
	publishCalls int
	lastCaption  string
	lastChildren []string
}

func (c *recordingPublishClient) CreateImageContainer(_ context.Context, _ application.Credential, _, caption string) (string, error) {
	c.imageCalls++
	c.lastCaption = caption
	if c.failStage == "container" {
		return "", errors.New("container failed")
	}
	return "container-image", nil
}

func (c *recordingPublishClient) CreateReelContainer(_ context.Context, _ application.Credential, _, caption string) (string, error) {
	c.reelCalls++
	c.lastCaption = caption
	return "container-reel", nil
}

func (c *recordingPublishClient) CreateCarouselItem(_ context.Context, _ application.Credential, _ string, _ bool) (string, error) {
	c.itemCalls++
	return "child", nil
}

func (c *recordingPublishClient) CreateCarouselContainer(_ context.Context, _ application.Credential, childIDs []string, caption string) (string, error) {
	c.carousel++
	c.lastChildren = childIDs
	c.lastCaption = caption
	return "container-carousel", nil
}

func (c *recordingPublishClient) Publish(_ context.Context, _ application.Credential, _ string) (application.PublishedMedia, error) {
	c.publishCalls++
	if c.failStage == "publish" {
		return application.PublishedMedia{}, errors.New("publish failed")
	}
	return application.PublishedMedia{ID: "media-123", Permalink: "https://instagram.com/p/xyz"}, nil
}

type fakePublicationRepository struct {
	created []domain.Publication
}

func (f *fakePublicationRepository) Create(_ context.Context, publication domain.Publication) (domain.Publication, error) {
	f.created = append(f.created, publication)
	return publication, nil
}

func (f *fakePublicationRepository) ListByContentForUser(_ context.Context, _, _ uuid.UUID) ([]domain.Publication, error) {
	return f.created, nil
}
