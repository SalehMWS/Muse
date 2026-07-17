package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/publishing/application"
	"github.com/SalehMWS/Muse/internal/publishing/domain"
)

func newPublishInput() application.PublishInput {
	return application.PublishInput{
		UserID:             uuid.New(),
		ContentID:          uuid.New(),
		InstagramAccountID: uuid.New(),
	}
}

func activeAccount() application.Account {
	return application.Account{ID: uuid.New(), InstagramUserID: "17841400000000000", AccessToken: "tok"}
}

func TestPublishUseCase_Image(t *testing.T) {
	client := &recordingPublishClient{}
	repo := &fakePublicationRepository{}
	uc := application.NewPublishUseCase(
		fakeAccountReader{account: activeAccount()},
		fakeContentReader{content: application.PublishableContent{
			Caption:     "Golden hour",
			Hashtags:    []string{"sunset", "#beach"},
			ContentType: "image",
			Media:       []application.MediaItem{{URL: "https://cdn/a.jpg", MediaType: "image"}},
		}},
		client, repo,
	)

	pub, err := uc.Execute(context.Background(), newPublishInput())
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if pub.Status != domain.StatusPublished || pub.PlatformPostID == nil || *pub.PlatformPostID != "media-123" {
		t.Fatalf("Execute() = %+v, unexpected", pub)
	}
	if client.imageCalls != 1 || client.publishCalls != 1 {
		t.Fatalf("client calls: image=%d publish=%d", client.imageCalls, client.publishCalls)
	}
	if !strings.Contains(client.lastCaption, "Golden hour") || !strings.Contains(client.lastCaption, "#sunset") || !strings.Contains(client.lastCaption, "#beach") {
		t.Fatalf("caption = %q, want caption + normalized hashtags", client.lastCaption)
	}
	if len(repo.created) != 1 {
		t.Fatalf("publications recorded = %d, want 1", len(repo.created))
	}
}

func TestPublishUseCase_Carousel(t *testing.T) {
	client := &recordingPublishClient{}
	uc := application.NewPublishUseCase(
		fakeAccountReader{account: activeAccount()},
		fakeContentReader{content: application.PublishableContent{
			Media: []application.MediaItem{
				{URL: "https://cdn/a.jpg", MediaType: "image"},
				{URL: "https://cdn/b.mp4", MediaType: "video"},
			},
		}},
		client, &fakePublicationRepository{},
	)

	if _, err := uc.Execute(context.Background(), newPublishInput()); err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if client.itemCalls != 2 || client.carousel != 1 || client.publishCalls != 1 {
		t.Fatalf("carousel calls: item=%d carousel=%d publish=%d", client.itemCalls, client.carousel, client.publishCalls)
	}
	if len(client.lastChildren) != 2 {
		t.Fatalf("children = %v, want 2", client.lastChildren)
	}
}

func TestPublishUseCase_ReelByContentType(t *testing.T) {
	client := &recordingPublishClient{}
	uc := application.NewPublishUseCase(
		fakeAccountReader{account: activeAccount()},
		fakeContentReader{content: application.PublishableContent{
			ContentType: "reel",
			Media:       []application.MediaItem{{URL: "https://cdn/a.mp4", MediaType: "video"}},
		}},
		client, &fakePublicationRepository{},
	)

	if _, err := uc.Execute(context.Background(), newPublishInput()); err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if client.reelCalls != 1 {
		t.Fatalf("reelCalls = %d, want 1", client.reelCalls)
	}
}

func TestPublishUseCase_NoMedia(t *testing.T) {
	uc := application.NewPublishUseCase(
		fakeAccountReader{account: activeAccount()},
		fakeContentReader{content: application.PublishableContent{ContentType: "image"}},
		&recordingPublishClient{}, &fakePublicationRepository{},
	)
	if _, err := uc.Execute(context.Background(), newPublishInput()); !errors.Is(err, application.ErrNoMedia) {
		t.Fatalf("Execute() error = %v, want %v", err, application.ErrNoMedia)
	}
}

func TestPublishUseCase_AccountNotFound(t *testing.T) {
	uc := application.NewPublishUseCase(
		fakeAccountReader{err: application.ErrAccountNotFound},
		fakeContentReader{},
		&recordingPublishClient{}, &fakePublicationRepository{},
	)
	if _, err := uc.Execute(context.Background(), newPublishInput()); !errors.Is(err, application.ErrAccountNotFound) {
		t.Fatalf("Execute() error = %v, want %v", err, application.ErrAccountNotFound)
	}
}

func TestPublishUseCase_PublishFailureRecordsFailedPublication(t *testing.T) {
	repo := &fakePublicationRepository{}
	uc := application.NewPublishUseCase(
		fakeAccountReader{account: activeAccount()},
		fakeContentReader{content: application.PublishableContent{
			ContentType: "image",
			Media:       []application.MediaItem{{URL: "https://cdn/a.jpg", MediaType: "image"}},
		}},
		&recordingPublishClient{failStage: "publish"}, repo,
	)

	if _, err := uc.Execute(context.Background(), newPublishInput()); !errors.Is(err, application.ErrPublishFailed) {
		t.Fatalf("Execute() error = %v, want %v", err, application.ErrPublishFailed)
	}
	if len(repo.created) != 1 || repo.created[0].Status != domain.StatusFailed {
		t.Fatalf("expected one failed publication recorded, got %+v", repo.created)
	}
}

func TestPublishUseCase_InvalidOverride(t *testing.T) {
	uc := application.NewPublishUseCase(
		fakeAccountReader{account: activeAccount()},
		fakeContentReader{content: application.PublishableContent{
			Media: []application.MediaItem{{URL: "https://cdn/a.jpg", MediaType: "image"}},
		}},
		&recordingPublishClient{}, &fakePublicationRepository{},
	)
	in := newPublishInput()
	in.MediaType = "story"
	if _, err := uc.Execute(context.Background(), in); !errors.Is(err, application.ErrInvalidMediaType) {
		t.Fatalf("Execute() error = %v, want %v", err, application.ErrInvalidMediaType)
	}
}
