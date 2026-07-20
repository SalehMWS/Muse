package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/SalehMWS/Muse/internal/ai/application"
	"github.com/SalehMWS/Muse/internal/shared/metrics"
)

const (
	defaultTimeout    = 30 * time.Second
	defaultMaxTokens  = 1024
	defaultTemp       = 0.7
	completionsPath   = "/chat/completions"
	responseFormatKey = "json_object"
)

const systemPrompt = `You are a caption engine for Instagram posts.

You MUST respond with a single raw JSON object and NOTHING else. Do not wrap it in markdown code fences. Do not add explanations, greetings, or any text before or after the JSON.

The JSON object MUST match this exact schema:
{"caption": <string>, "hashtags": <array of strings>}

Rules:
- "caption": one engaging Instagram caption for the user's request. Emoji are allowed. Do NOT put hashtags inside the caption.
- "hashtags": between 5 and 15 relevant hashtags. Every element is a single token that starts with "#", contains no spaces, and is lowercase.
- Output ONLY the JSON object. No prose. No code fences. No trailing commentary.

Example input: a cozy coffee shop on a rainy morning
Example output: {"caption":"Rainy mornings and fresh espresso are the perfect slow start.","hashtags":["#coffee","#rainyday","#coffeeshop","#morningvibes","#espresso"]}`

var (
	ErrEmptyPrompt     = errors.New("user prompt is required")
	ErrProvider        = errors.New("ai provider returned an error")
	ErrEmptyCompletion = errors.New("ai provider returned an empty completion")
	ErrInvalidResponse = errors.New("ai provider returned an unparseable completion")
)

type Config struct {
	Provider  string
	BaseURL   string
	APIKey    string
	Model     string
	Timeout   time.Duration
	MaxTokens int
}

type OpenAICompatibleProvider struct {
	provider   string
	baseURL    string
	apiKey     string
	modelName  string
	maxTokens  int
	logger     *zap.Logger
	httpClient *http.Client
	recorder   *metrics.AI
}

var _ application.LLMProvider = (*OpenAICompatibleProvider)(nil)

func New(cfg Config, logger *zap.Logger, recorder *metrics.AI) *OpenAICompatibleProvider {
	if logger == nil {
		logger = zap.NewNop()
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	maxTokens := cfg.MaxTokens
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}

	provider := cfg.Provider
	if provider == "" {
		provider = "openai-compatible"
	}

	return &OpenAICompatibleProvider{
		provider:   provider,
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:     cfg.APIKey,
		modelName:  cfg.Model,
		maxTokens:  maxTokens,
		logger:     logger,
		httpClient: &http.Client{Timeout: timeout},
		recorder:   recorder,
	}
}

func (p *OpenAICompatibleProvider) GenerateCaptions(ctx context.Context, userPrompt string) (*application.CaptionResult, error) {
	start := time.Now()
	result, err := p.generateCaptions(ctx, userPrompt)
	p.recorder.Request(p.provider, p.modelName, metrics.Outcome(err), time.Since(start))
	return result, err
}

func (p *OpenAICompatibleProvider) generateCaptions(ctx context.Context, userPrompt string) (*application.CaptionResult, error) {
	if strings.TrimSpace(userPrompt) == "" {
		return nil, ErrEmptyPrompt
	}

	payload, err := json.Marshal(chatRequest{
		Model: p.modelName,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		ResponseFormat: responseFormat{Type: responseFormatKey},
		Temperature:    defaultTemp,
		MaxTokens:      p.maxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("encode ai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+completionsPath, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build ai request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.logger.Warn("ai request failed", zap.String("model", p.modelName), zap.Error(err))
		return nil, fmt.Errorf("%w: %v", ErrProvider, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read ai response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		message := providerErrorMessage(body, resp.StatusCode)
		p.logger.Warn("ai provider returned a non-success status",
			zap.Int("status", resp.StatusCode),
			zap.String("model", p.modelName),
			zap.String("message", message),
		)
		return nil, fmt.Errorf("%w: %s", ErrProvider, message)
	}

	var parsed chatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("%w: decode envelope: %v", ErrInvalidResponse, err)
	}
	if parsed.Usage != nil {
		p.recorder.Tokens(p.provider, p.modelName, parsed.Usage.PromptTokens, parsed.Usage.CompletionTokens)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		p.logger.Warn("ai provider returned an error object",
			zap.String("model", p.modelName),
			zap.String("message", parsed.Error.Message),
		)
		return nil, fmt.Errorf("%w: %s", ErrProvider, parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("%w: no choices", ErrEmptyCompletion)
	}

	content := extractJSONObject(parsed.Choices[0].Message.Content)
	if content == "" {
		return nil, fmt.Errorf("%w: blank content", ErrEmptyCompletion)
	}

	var result application.CaptionResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		p.logger.Warn("ai completion was not valid caption json",
			zap.String("model", p.modelName),
			zap.String("content", content),
		)
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	if strings.TrimSpace(result.Caption) == "" {
		return nil, fmt.Errorf("%w: missing caption", ErrEmptyCompletion)
	}
	result.Hashtags = normalizeHashtags(result.Hashtags)

	return &result, nil
}

func extractJSONObject(content string) string {
	trimmed := strings.TrimSpace(content)

	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
		if newline := strings.IndexByte(trimmed, '\n'); newline >= 0 && !strings.HasPrefix(trimmed, "{") {
			trimmed = trimmed[newline+1:]
		}
		trimmed = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(trimmed), "```"))
	}

	start := strings.IndexByte(trimmed, '{')
	end := strings.LastIndexByte(trimmed, '}')
	if start >= 0 && end > start {
		return trimmed[start : end+1]
	}
	return trimmed
}

func normalizeHashtags(tags []string) []string {
	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if !strings.HasPrefix(tag, "#") {
			tag = "#" + tag
		}
		normalized = append(normalized, tag)
	}
	return normalized
}

func providerErrorMessage(body []byte, status int) string {
	var parsed chatResponse
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Error != nil && parsed.Error.Message != "" {
		return parsed.Error.Message
	}
	return fmt.Sprintf("ai provider returned status %d", status)
}

type chatRequest struct {
	Model          string         `json:"model"`
	Messages       []chatMessage  `json:"messages"`
	ResponseFormat responseFormat `json:"response_format"`
	Temperature    float64        `json:"temperature,omitempty"`
	MaxTokens      int            `json:"max_tokens,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message      chatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}
