package ai

import (
	"go.uber.org/zap"

	"github.com/SalehMWS/Muse/internal/ai/application"
	"github.com/SalehMWS/Muse/internal/ai/infrastructure/openai"
	"github.com/SalehMWS/Muse/internal/shared/config"
)

func NewProvider(cfg config.AI, logger *zap.Logger) application.LLMProvider {
	return openai.New(openai.Config{
		BaseURL:   cfg.BaseURL,
		APIKey:    cfg.APIKey,
		Model:     cfg.Model,
		Timeout:   cfg.HTTPTimeout,
		MaxTokens: cfg.MaxTokens,
	}, logger)
}
