package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	aiapp "github.com/SalehMWS/Muse/internal/ai/application"
	"github.com/SalehMWS/Muse/internal/content/domain"
	"github.com/SalehMWS/Muse/internal/shared/metrics"
)

type GenerateCaptionUseCase struct {
	repo     ContentRepository
	ai       aiapp.LLMProvider
	business *metrics.Business
}

func NewGenerateCaptionUseCase(repo ContentRepository, ai aiapp.LLMProvider, business *metrics.Business) *GenerateCaptionUseCase {
	return &GenerateCaptionUseCase{repo: repo, ai: ai, business: business}
}

func (uc *GenerateCaptionUseCase) Execute(ctx context.Context, userID, contentID uuid.UUID, promptOverride string) (domain.Content, error) {
	content, err := uc.repo.FindByIDForUser(ctx, contentID, userID)
	if err != nil {
		return domain.Content{}, err
	}

	prompt := strings.TrimSpace(promptOverride)
	if prompt == "" {
		prompt = defaultCaptionPrompt(content)
	}

	result, err := uc.ai.GenerateCaptions(ctx, prompt)
	if err != nil {
		return domain.Content{}, fmt.Errorf("%w: %v", ErrCaptionUnavailable, err)
	}

	caption := result.Caption
	hashtags := result.Hashtags
	if err := content.Apply(domain.UpdateContentInput{Caption: &caption, Tags: &hashtags}); err != nil {
		return domain.Content{}, err
	}

	updated, err := uc.repo.Update(ctx, content)
	if err != nil {
		return domain.Content{}, err
	}

	uc.business.Record(metrics.EventCaptionGenerated)

	return updated, nil
}

func defaultCaptionPrompt(content domain.Content) string {
	if title := strings.TrimSpace(content.Title); title != "" {
		return title
	}
	if caption := strings.TrimSpace(content.Caption); caption != "" {
		return caption
	}
	return "an engaging instagram post"
}
