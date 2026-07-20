package meta_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SalehMWS/Muse/internal/publishing/application"
	"github.com/SalehMWS/Muse/internal/publishing/infrastructure/meta"
)

func TestPublishClient_ImageAndPublish(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/123/media", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.FormValue("image_url") == "" || r.FormValue("access_token") != "tok" {
			http.Error(w, "bad", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"id":"container-1"}`))
	})
	mux.HandleFunc("/123/media_publish", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.FormValue("creation_id") != "container-1" {
			http.Error(w, "bad", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"id":"media-1"}`))
	})
	mux.HandleFunc("/media-1", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"permalink":"https://instagram.com/p/abc"}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := meta.NewPublishClient(server.URL, 0, nil)
	cred := application.Credential{InstagramUserID: "123", AccessToken: "tok"}

	container, err := client.CreateImageContainer(context.Background(), cred, "https://cdn/a.jpg", "caption")
	if err != nil {
		t.Fatalf("CreateImageContainer() unexpected error: %v", err)
	}
	if container != "container-1" {
		t.Fatalf("container = %q, want container-1", container)
	}

	published, err := client.Publish(context.Background(), cred, container)
	if err != nil {
		t.Fatalf("Publish() unexpected error: %v", err)
	}
	if published.ID != "media-1" || published.Permalink != "https://instagram.com/p/abc" {
		t.Fatalf("Publish() = %+v, unexpected", published)
	}
}

func TestPublishClient_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/123/media", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"Invalid image_url","type":"OAuthException"}}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := meta.NewPublishClient(server.URL, 0, nil)
	cred := application.Credential{InstagramUserID: "123", AccessToken: "tok"}

	_, err := client.CreateImageContainer(context.Background(), cred, "bad", "caption")
	if !errors.Is(err, meta.ErrGraphAPI) {
		t.Fatalf("error = %v, want ErrGraphAPI", err)
	}
	if !strings.Contains(err.Error(), "Invalid image_url") {
		t.Fatalf("error = %v, want graph message surfaced", err)
	}
}
