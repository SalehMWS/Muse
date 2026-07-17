package embedding

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
)

var ErrEmbeddingAPI = errors.New("embedding api error")

type OpenAIEmbedder struct {
	baseURL    string
	apiKey     string
	model      string
	dim        int
	httpClient *http.Client
}

func NewOpenAIEmbedder(baseURL, apiKey, model string, dim int, timeout time.Duration) *OpenAIEmbedder {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if dim <= 0 {
		dim = 1536
	}
	return &OpenAIEmbedder{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		model:      model,
		dim:        dim,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (e *OpenAIEmbedder) Dimension() int {
	return e.dim
}

func (e *OpenAIEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	payload, err := json.Marshal(embeddingRequest{Model: e.model, Input: texts})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/embeddings", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrEmbeddingAPI, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEmbeddingAPI, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: read response: %v", ErrEmbeddingAPI, err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%w: status %d: %s", ErrEmbeddingAPI, resp.StatusCode, string(body))
	}

	var parsed embeddingResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("%w: decode response: %v", ErrEmbeddingAPI, err)
	}
	if len(parsed.Data) != len(texts) {
		return nil, fmt.Errorf("%w: expected %d embeddings, got %d", ErrEmbeddingAPI, len(texts), len(parsed.Data))
	}

	out := make([][]float32, len(parsed.Data))
	for i, item := range parsed.Data {
		out[i] = item.Embedding
	}
	return out, nil
}

type embeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}
