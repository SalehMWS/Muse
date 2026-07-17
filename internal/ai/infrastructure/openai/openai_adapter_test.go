package openai_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/SalehMWS/Muse/internal/ai/infrastructure/openai"
)

func completionServer(t *testing.T, content string) (*openai.OpenAICompatibleProvider, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("path = %q, want /chat/completions", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization = %q, want Bearer test-key", got)
		}

		var body map[string]any
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Errorf("request body not json: %v", err)
		}
		if format, _ := body["response_format"].(map[string]any); format["type"] != "json_object" {
			t.Errorf("response_format = %v, want json_object", body["response_format"])
		}
		if body["model"] != "test-model" {
			t.Errorf("model = %v, want test-model", body["model"])
		}

		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":` + strconv.Quote(content) + `}}]}`))
	}))

	provider := openai.New(openai.Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "test-model",
	}, nil)
	return provider, server
}

func TestGenerateCaptions_Success(t *testing.T) {
	provider, server := completionServer(t, `{"caption":"Golden hour magic","hashtags":["#sunset","beach"]}`)
	defer server.Close()

	result, err := provider.GenerateCaptions(context.Background(), "a sunset at the beach")
	if err != nil {
		t.Fatalf("GenerateCaptions() unexpected error: %v", err)
	}
	if result.Caption != "Golden hour magic" {
		t.Fatalf("Caption = %q, want Golden hour magic", result.Caption)
	}
	if len(result.Hashtags) != 2 || result.Hashtags[0] != "#sunset" || result.Hashtags[1] != "#beach" {
		t.Fatalf("Hashtags = %v, want normalized [#sunset #beach]", result.Hashtags)
	}
}

func TestGenerateCaptions_StripsCodeFences(t *testing.T) {
	fenced := "```json\n{\"caption\":\"Fenced\",\"hashtags\":[\"#x\"]}\n```"
	provider, server := completionServer(t, fenced)
	defer server.Close()

	result, err := provider.GenerateCaptions(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("GenerateCaptions() unexpected error: %v", err)
	}
	if result.Caption != "Fenced" {
		t.Fatalf("Caption = %q, want Fenced", result.Caption)
	}
}

func TestGenerateCaptions_EmptyPrompt(t *testing.T) {
	provider := openai.New(openai.Config{BaseURL: "http://127.0.0.1:0", APIKey: "k", Model: "m"}, nil)
	if _, err := provider.GenerateCaptions(context.Background(), "   "); !errors.Is(err, openai.ErrEmptyPrompt) {
		t.Fatalf("error = %v, want ErrEmptyPrompt", err)
	}
}

func TestGenerateCaptions_NonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"Invalid API Key","type":"authentication_error"}}`))
	}))
	defer server.Close()

	provider := openai.New(openai.Config{BaseURL: server.URL, APIKey: "bad", Model: "m"}, nil)
	_, err := provider.GenerateCaptions(context.Background(), "prompt")
	if !errors.Is(err, openai.ErrProvider) {
		t.Fatalf("error = %v, want ErrProvider", err)
	}
	if !strings.Contains(err.Error(), "Invalid API Key") {
		t.Fatalf("error = %v, want provider message surfaced", err)
	}
}

func TestGenerateCaptions_ErrorObjectWithOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"error":{"message":"rate limited"}}`))
	}))
	defer server.Close()

	provider := openai.New(openai.Config{BaseURL: server.URL, APIKey: "k", Model: "m"}, nil)
	if _, err := provider.GenerateCaptions(context.Background(), "prompt"); !errors.Is(err, openai.ErrProvider) {
		t.Fatalf("error = %v, want ErrProvider", err)
	}
}

func TestGenerateCaptions_InvalidCompletionJSON(t *testing.T) {
	provider, server := completionServer(t, "sorry, I cannot help with that")
	defer server.Close()

	if _, err := provider.GenerateCaptions(context.Background(), "prompt"); !errors.Is(err, openai.ErrInvalidResponse) {
		t.Fatalf("error = %v, want ErrInvalidResponse", err)
	}
}
