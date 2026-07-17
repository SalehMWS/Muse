package content

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	contentapp "github.com/SalehMWS/Muse/internal/content/application"
	contentpg "github.com/SalehMWS/Muse/internal/content/infrastructure/postgres"
	"github.com/SalehMWS/Muse/internal/publishing/application"
)

type ContentReader struct {
	content *contentpg.ContentRepository
	media   *contentpg.MediaRepository
}

func NewContentReader(pool *pgxpool.Pool) *ContentReader {
	return &ContentReader{
		content: contentpg.NewContentRepository(pool),
		media:   contentpg.NewMediaRepository(pool),
	}
}

func (r *ContentReader) ContentForUser(ctx context.Context, userID, contentID uuid.UUID) (application.PublishableContent, error) {
	content, err := r.content.FindByIDForUser(ctx, contentID, userID)
	if err != nil {
		if errors.Is(err, contentapp.ErrContentNotFound) {
			return application.PublishableContent{}, application.ErrContentNotFound
		}
		return application.PublishableContent{}, err
	}

	mediaRows, err := r.media.ListByContent(ctx, contentID)
	if err != nil {
		return application.PublishableContent{}, err
	}

	items := make([]application.MediaItem, 0, len(mediaRows))
	for _, media := range mediaRows {
		items = append(items, application.MediaItem{URL: media.URL, MediaType: string(media.MediaType)})
	}

	return application.PublishableContent{
		ContentID:   content.ID,
		Caption:     content.Caption,
		Hashtags:    content.Tags,
		ContentType: string(content.ContentType),
		Media:       items,
	}, nil
}
