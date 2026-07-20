package ai

import (
	"go.uber.org/zap"

	"github.com/SalehMWS/Muse/internal/ai/application"
	"github.com/SalehMWS/Muse/internal/ai/infrastructure/openai"
	"github.com/SalehMWS/Muse/internal/shared/config"
	"github.com/SalehMWS/Muse/internal/shared/metrics"
)

func NewProvider(cfg config.AI, logger *zap.Logger, recorder *metrics.AI) application.LLMProvider {
	return openai.New(openai.Config{
		Provider:  cfg.Provider,
		BaseURL:   cfg.BaseURL,
		APIKey:    cfg.APIKey,
		Model:     cfg.Model,
		Timeout:   cfg.HTTPTimeout,
		MaxTokens: cfg.MaxTokens,
	}, logger, recorder)
}
