package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	auditapp "github.com/SalehMWS/Muse/internal/audit/application"
	auditdomain "github.com/SalehMWS/Muse/internal/audit/domain"
	"github.com/SalehMWS/Muse/internal/publishing/domain"
	"github.com/SalehMWS/Muse/internal/shared/metrics"
)

type PublishUseCase struct {
	accounts AccountReader
	contents ContentReader
	client   PublishClient
	repo     PublicationRepository
	recorder *metrics.Publishing
	business *metrics.Business
	audit    *auditapp.Recorder
}

func NewPublishUseCase(accounts AccountReader, contents ContentReader, client PublishClient, repo PublicationRepository, recorder *metrics.Publishing, business *metrics.Business, audit *auditapp.Recorder) *PublishUseCase {
	return &PublishUseCase{
		accounts: accounts,
		contents: contents,
		client:   client,
		repo:     repo,
		recorder: recorder,
		business: business,
		audit:    audit,
	}
}

func (uc *PublishUseCase) recordAudit(ctx context.Context, in PublishInput, result auditdomain.Result, metadata map[string]string) {
	userID := in.UserID
	uc.audit.Record(ctx, auditapp.Entry{
		UserID:       &userID,
		Action:       auditdomain.ActionContentPublished,
		Result:       result,
		ResourceType: "content",
		ResourceID:   in.ContentID.String(),
		Metadata:     metadata,
	})
}

type PublishInput struct {
	UserID             uuid.UUID
	ContentID          uuid.UUID
	InstagramAccountID uuid.UUID
	MediaType          string
}

func (uc *PublishUseCase) Execute(ctx context.Context, in PublishInput) (domain.Publication, error) {
	account, err := uc.accounts.AccountForUser(ctx, in.UserID, in.InstagramAccountID)
	if err != nil {
		return domain.Publication{}, err
	}

	content, err := uc.contents.ContentForUser(ctx, in.UserID, in.ContentID)
	if err != nil {
		return domain.Publication{}, err
	}
	if len(content.Media) == 0 {
		return domain.Publication{}, ErrNoMedia
	}

	mediaType, err := resolveMediaType(in.MediaType, content)
	if err != nil {
		return domain.Publication{}, err
	}

	caption := buildCaption(content.Caption, content.Hashtags)
	cred := Credential{InstagramUserID: account.InstagramUserID, AccessToken: account.AccessToken}

	start := time.Now()
	published, publishErr := uc.publish(ctx, cred, mediaType, content, caption)
	uc.recorder.Published(string(mediaType), metrics.Outcome(publishErr), time.Since(start))

	publication := domain.Publication{
		ID:                 uuid.New(),
		UserID:             in.UserID,
		ContentID:          in.ContentID,
		InstagramAccountID: account.ID,
		Platform:           "instagram",
	}

	if publishErr != nil {
		publication.Status = domain.StatusFailed
		publication.ResponseJSON = marshalJSON(map[string]string{"error": publishErr.Error()})
		if _, err := uc.repo.Create(ctx, publication); err != nil {
			return domain.Publication{}, err
		}
		uc.recordAudit(ctx, in, auditdomain.ResultFailure, map[string]string{"media_type": string(mediaType)})
		return domain.Publication{}, fmt.Errorf("%w: %v", ErrPublishFailed, publishErr)
	}

	now := time.Now()
	publication.Status = domain.StatusPublished
	publication.PlatformPostID = &published.ID
	publication.PublishedAt = &now
	if published.Permalink != "" {
		permalink := published.Permalink
		publication.Permalink = &permalink
	}
	publication.ResponseJSON = marshalJSON(map[string]string{"media_id": published.ID, "permalink": published.Permalink})

	created, err := uc.repo.Create(ctx, publication)
	if err != nil {
		return domain.Publication{}, err
	}

	uc.business.Record(metrics.EventPostPublished)
	uc.recordAudit(ctx, in, auditdomain.ResultSuccess, map[string]string{
		"media_type":  string(mediaType),
		"platform_id": published.ID,
	})

	return created, nil
}

func (uc *PublishUseCase) publish(ctx context.Context, cred Credential, mediaType domain.MediaType, content PublishableContent, caption string) (PublishedMedia, error) {
	switch mediaType {
	case domain.MediaImage:
		container, err := uc.client.CreateImageContainer(ctx, cred, content.Media[0].URL, caption)
		if err != nil {
			return PublishedMedia{}, err
		}
		return uc.client.Publish(ctx, cred, container)

	case domain.MediaReel:
		container, err := uc.client.CreateReelContainer(ctx, cred, content.Media[0].URL, caption)
		if err != nil {
			return PublishedMedia{}, err
		}
		return uc.client.Publish(ctx, cred, container)

	case domain.MediaCarousel:
		childIDs := make([]string, 0, len(content.Media))
		for _, item := range content.Media {
			child, err := uc.client.CreateCarouselItem(ctx, cred, item.URL, item.MediaType == "video")
			if err != nil {
				return PublishedMedia{}, err
			}
			childIDs = append(childIDs, child)
		}
		container, err := uc.client.CreateCarouselContainer(ctx, cred, childIDs, caption)
		if err != nil {
			return PublishedMedia{}, err
		}
		return uc.client.Publish(ctx, cred, container)

	default:
		return PublishedMedia{}, ErrInvalidMediaType
	}
}

func resolveMediaType(override string, content PublishableContent) (domain.MediaType, error) {
	if override != "" {
		mediaType := domain.MediaType(strings.ToLower(strings.TrimSpace(override)))
		if !domain.ValidMediaType(mediaType) {
			return "", ErrInvalidMediaType
		}
		return mediaType, nil
	}

	if len(content.Media) > 1 {
		return domain.MediaCarousel, nil
	}

	switch strings.ToLower(content.ContentType) {
	case "reel", "video":
		return domain.MediaReel, nil
	}

	if len(content.Media) == 1 && content.Media[0].MediaType == "video" {
		return domain.MediaReel, nil
	}
	return domain.MediaImage, nil
}

func buildCaption(caption string, hashtags []string) string {
	caption = strings.TrimSpace(caption)

	tags := make([]string, 0, len(hashtags))
	for _, tag := range hashtags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if !strings.HasPrefix(tag, "#") {
			tag = "#" + tag
		}
		tags = append(tags, tag)
	}

	if len(tags) == 0 {
		return caption
	}
	tagline := strings.Join(tags, " ")
	if caption == "" {
		return tagline
	}
	return caption + "\n\n" + tagline
}

func marshalJSON(value any) []byte {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return raw
}
