package http_test

import (
	"context"

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

type stubPublishClient struct{}

func (stubPublishClient) CreateImageContainer(context.Context, application.Credential, string, string) (string, error) {
	return "container", nil
}
func (stubPublishClient) CreateReelContainer(context.Context, application.Credential, string, string) (string, error) {
	return "container", nil
}
func (stubPublishClient) CreateCarouselItem(context.Context, application.Credential, string, bool) (string, error) {
	return "child", nil
}
func (stubPublishClient) CreateCarouselContainer(context.Context, application.Credential, []string, string) (string, error) {
	return "container", nil
}
func (stubPublishClient) Publish(context.Context, application.Credential, string) (application.PublishedMedia, error) {
	return application.PublishedMedia{ID: "media-1", Permalink: "https://instagram.com/p/x"}, nil
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
