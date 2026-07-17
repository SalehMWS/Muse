package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/content/domain"
)

type AttachMediaUseCase struct {
	content ContentRepository
	media   MediaRepository
}

func NewAttachMediaUseCase(content ContentRepository, media MediaRepository) *AttachMediaUseCase {
	return &AttachMediaUseCase{content: content, media: media}
}

func (uc *AttachMediaUseCase) Execute(ctx context.Context, userID, contentID uuid.UUID, url, mediaType string, position int) (domain.Media, error) {
	if _, err := uc.content.FindByIDForUser(ctx, contentID, userID); err != nil {
		return domain.Media{}, err
	}

	media, err := domain.NewMedia(contentID, url, mediaType, position)
	if err != nil {
		return domain.Media{}, err
	}
	return uc.media.Create(ctx, media)
}

type ListMediaUseCase struct {
	content ContentRepository
	media   MediaRepository
}

func NewListMediaUseCase(content ContentRepository, media MediaRepository) *ListMediaUseCase {
	return &ListMediaUseCase{content: content, media: media}
}

func (uc *ListMediaUseCase) Execute(ctx context.Context, userID, contentID uuid.UUID) ([]domain.Media, error) {
	if _, err := uc.content.FindByIDForUser(ctx, contentID, userID); err != nil {
		return nil, err
	}
	return uc.media.ListByContent(ctx, contentID)
}

type DeleteMediaUseCase struct {
	content ContentRepository
	media   MediaRepository
}

func NewDeleteMediaUseCase(content ContentRepository, media MediaRepository) *DeleteMediaUseCase {
	return &DeleteMediaUseCase{content: content, media: media}
}

func (uc *DeleteMediaUseCase) Execute(ctx context.Context, userID, contentID, mediaID uuid.UUID) error {
	if _, err := uc.content.FindByIDForUser(ctx, contentID, userID); err != nil {
		return err
	}
	if _, err := uc.media.FindByIDForContent(ctx, mediaID, contentID); err != nil {
		return err
	}
	return uc.media.DeleteForContent(ctx, mediaID, contentID)
}
