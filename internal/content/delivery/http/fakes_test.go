package http_test

import (
	"context"

	"github.com/google/uuid"

	aiapp "github.com/SalehMWS/Muse/internal/ai/application"
	"github.com/SalehMWS/Muse/internal/content/application"
	"github.com/SalehMWS/Muse/internal/content/domain"
)

type fakeLLMProvider struct {
	result *aiapp.CaptionResult
	err    error
}

func (f *fakeLLMProvider) GenerateCaptions(_ context.Context, _ string) (*aiapp.CaptionResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

type fakeContentRepository struct {
	byID map[uuid.UUID]domain.Content
}

func newFakeContentRepository() *fakeContentRepository {
	return &fakeContentRepository{byID: map[uuid.UUID]domain.Content{}}
}

func (f *fakeContentRepository) Create(_ context.Context, content domain.Content) (domain.Content, error) {
	f.byID[content.ID] = content
	return content, nil
}

func (f *fakeContentRepository) FindByIDForUser(_ context.Context, id, userID uuid.UUID) (domain.Content, error) {
	content, ok := f.byID[id]
	if !ok || content.UserID != userID {
		return domain.Content{}, application.ErrContentNotFound
	}
	return content, nil
}

func (f *fakeContentRepository) Update(_ context.Context, content domain.Content) (domain.Content, error) {
	if _, ok := f.byID[content.ID]; !ok {
		return domain.Content{}, application.ErrContentNotFound
	}
	f.byID[content.ID] = content
	return content, nil
}

func (f *fakeContentRepository) List(_ context.Context, filter application.ListFilter) ([]domain.Content, error) {
	items := make([]domain.Content, 0)
	for _, content := range f.byID {
		if content.UserID == filter.UserID {
			items = append(items, content)
		}
	}
	return items, nil
}
