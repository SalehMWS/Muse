package application_test

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
	prompt string
}

func (f *fakeLLMProvider) GenerateCaptions(_ context.Context, prompt string) (*aiapp.CaptionResult, error) {
	f.prompt = prompt
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

type fakeContentRepository struct {
	byID       map[uuid.UUID]domain.Content
	listReturn []domain.Content
	listErr    error
	lastFilter application.ListFilter
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
	f.lastFilter = filter
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listReturn, nil
}
