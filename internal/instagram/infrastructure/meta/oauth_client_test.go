package meta_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SalehMWS/Muse/internal/instagram/infrastructure/meta"
	"github.com/SalehMWS/Muse/internal/shared/config"
)

func newTestClient(handler http.Handler) (*meta.OAuthClient, *httptest.Server) {
	server := httptest.NewServer(handler)
	cfg := config.Instagram{
		ClientID:     "client-123",
		ClientSecret: "secret-456",
		RedirectURI:  "https://app.example.com/callback",
		Scopes:       "instagram_business_basic",
		AuthBaseURL:  "https://www.instagram.com",
		APIBaseURL:   server.URL,
		GraphBaseURL: server.URL,
	}
	return meta.NewOAuthClient(cfg), server
}

func TestOAuthClient_AuthorizationURL(t *testing.T) {
	client, server := newTestClient(http.NewServeMux())
	defer server.Close()

	url := client.AuthorizationURL("state-token")
	for _, want := range []string{"client_id=client-123", "state=state-token", "response_type=code", "scope=instagram_business_basic"} {
		if !strings.Contains(url, want) {
			t.Fatalf("AuthorizationURL() = %q, missing %q", url, want)
		}
	}
}

func TestOAuthClient_ExchangeCode(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.FormValue("code") != "the-code" || r.FormValue("grant_type") != "authorization_code" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"short-token","user_id":123}`))
	})
	mux.HandleFunc("/access_token", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("grant_type") != "ig_exchange_token" || r.URL.Query().Get("access_token") != "short-token" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"long-token","token_type":"bearer","expires_in":5184000}`))
	})

	client, server := newTestClient(mux)
	defer server.Close()

	token, err := client.ExchangeCode(context.Background(), "the-code")
	if err != nil {
		t.Fatalf("ExchangeCode() unexpected error: %v", err)
	}
	if token.AccessToken != "long-token" {
		t.Fatalf("ExchangeCode() token = %q, want long-token", token.AccessToken)
	}
	if token.ExpiresIn.Seconds() != 5184000 {
		t.Fatalf("ExchangeCode() expiresIn = %v, want 5184000s", token.ExpiresIn)
	}
}

func TestOAuthClient_FetchProfile(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/me", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"user_id":"17841400000000000","username":"brand","account_type":"BUSINESS"}`))
	})

	client, server := newTestClient(mux)
	defer server.Close()

	profile, err := client.FetchProfile(context.Background(), "long-token")
	if err != nil {
		t.Fatalf("FetchProfile() unexpected error: %v", err)
	}
	if profile.UserID != "17841400000000000" || profile.Username != "brand" || profile.AccountType != "BUSINESS" {
		t.Fatalf("FetchProfile() = %+v, unexpected", profile)
	}
}

func TestOAuthClient_RefreshToken(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/refresh_access_token", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("grant_type") != "ig_refresh_token" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"refreshed-token","expires_in":5184000}`))
	})

	client, server := newTestClient(mux)
	defer server.Close()

	token, err := client.RefreshToken(context.Background(), "long-token")
	if err != nil {
		t.Fatalf("RefreshToken() unexpected error: %v", err)
	}
	if token.AccessToken != "refreshed-token" {
		t.Fatalf("RefreshToken() token = %q, want refreshed-token", token.AccessToken)
	}
}

func TestOAuthClient_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/me", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"Invalid OAuth access token","type":"OAuthException","code":190}}`))
	})

	client, server := newTestClient(mux)
	defer server.Close()

	_, err := client.FetchProfile(context.Background(), "bad-token")
	if err == nil {
		t.Fatal("FetchProfile() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "Invalid OAuth access token") {
		t.Fatalf("FetchProfile() error = %v, want to contain meta message", err)
	}
}
