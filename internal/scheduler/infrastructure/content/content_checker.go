package content

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	contentapp "github.com/SalehMWS/Muse/internal/content/application"
	contentpg "github.com/SalehMWS/Muse/internal/content/infrastructure/postgres"
	"github.com/SalehMWS/Muse/internal/scheduler/application"
)

type ContentChecker struct {
	repo *contentpg.ContentRepository
}

func NewContentChecker(pool *pgxpool.Pool) *ContentChecker {
	return &ContentChecker{repo: contentpg.NewContentRepository(pool)}
}

func (c *ContentChecker) EnsureOwned(ctx context.Context, userID, contentID uuid.UUID) error {
	if _, err := c.repo.FindByIDForUser(ctx, contentID, userID); err != nil {
		if errors.Is(err, contentapp.ErrContentNotFound) {
			return application.ErrContentNotFound
		}
		return err
	}
	return nil
}
