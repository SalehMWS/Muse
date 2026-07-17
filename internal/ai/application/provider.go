package application

import "context"

type CaptionResult struct {
	Caption  string   `json:"caption"`
	Hashtags []string `json:"hashtags"`
}

type LLMProvider interface {
	GenerateCaptions(ctx context.Context, userPrompt string) (*CaptionResult, error)
}
