package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	aiapp "github.com/SalehMWS/Muse/internal/ai/application"
	"github.com/SalehMWS/Muse/internal/content/application"
	"github.com/SalehMWS/Muse/internal/content/domain"
)

func TestGenerateCaptionUseCase_Execute(t *testing.T) {
	t.Run("success applies caption and normalized tags", func(t *testing.T) {
		repo := newFakeContentRepository()
		userID := uuid.New()
		created, _ := application.NewCreateUseCase(repo).Execute(context.Background(), userID, domain.NewContentInput{Title: "Coffee shop"})

		provider := &fakeLLMProvider{result: &aiapp.CaptionResult{
			Caption:  "Slow mornings, strong espresso.",
			Hashtags: []string{"#coffee", "espresso"},
		}}
		uc := application.NewGenerateCaptionUseCase(repo, provider)

		out, err := uc.Execute(context.Background(), userID, created.ID, "")
		if err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}
		if out.Caption != "Slow mornings, strong espresso." {
			t.Fatalf("Caption = %q, unexpected", out.Caption)
		}
		if len(out.Tags) != 2 || out.Tags[0] != "coffee" || out.Tags[1] != "espresso" {
			t.Fatalf("Tags = %v, want normalized [coffee espresso]", out.Tags)
		}
		if provider.prompt != "Coffee shop" {
			t.Fatalf("prompt = %q, want fallback to content title", provider.prompt)
		}
	})

	t.Run("explicit prompt override is used", func(t *testing.T) {
		repo := newFakeContentRepository()
		userID := uuid.New()
		created, _ := application.NewCreateUseCase(repo).Execute(context.Background(), userID, domain.NewContentInput{Title: "Ignored"})
		provider := &fakeLLMProvider{result: &aiapp.CaptionResult{Caption: "c", Hashtags: []string{"#a"}}}
		uc := application.NewGenerateCaptionUseCase(repo, provider)

		if _, err := uc.Execute(context.Background(), userID, created.ID, "custom prompt"); err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}
		if provider.prompt != "custom prompt" {
			t.Fatalf("prompt = %q, want custom prompt", provider.prompt)
		}
	})

	t.Run("provider error maps to caption unavailable", func(t *testing.T) {
		repo := newFakeContentRepository()
		userID := uuid.New()
		created, _ := application.NewCreateUseCase(repo).Execute(context.Background(), userID, domain.NewContentInput{Title: "x"})
		uc := application.NewGenerateCaptionUseCase(repo, &fakeLLMProvider{err: errors.New("provider down")})

		if _, err := uc.Execute(context.Background(), userID, created.ID, ""); !errors.Is(err, application.ErrCaptionUnavailable) {
			t.Fatalf("Execute() error = %v, want %v", err, application.ErrCaptionUnavailable)
		}
	})

	t.Run("unknown content", func(t *testing.T) {
		uc := application.NewGenerateCaptionUseCase(newFakeContentRepository(), &fakeLLMProvider{})
		if _, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), ""); !errors.Is(err, application.ErrContentNotFound) {
			t.Fatalf("Execute() error = %v, want %v", err, application.ErrContentNotFound)
		}
	})
}
