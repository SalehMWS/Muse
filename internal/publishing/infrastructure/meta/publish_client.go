package meta

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/SalehMWS/Muse/internal/publishing/application"
)

const defaultTimeout = 30 * time.Second

var ErrGraphAPI = errors.New("instagram graph api error")

type PublishClient struct {
	baseURL    string
	httpClient *http.Client
}

var _ application.PublishClient = (*PublishClient)(nil)

func NewPublishClient(baseURL string, timeout time.Duration) *PublishClient {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	return &PublishClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *PublishClient) CreateImageContainer(ctx context.Context, cred application.Credential, imageURL, caption string) (string, error) {
	form := url.Values{}
	form.Set("image_url", imageURL)
	form.Set("caption", caption)
	form.Set("access_token", cred.AccessToken)
	return c.createContainer(ctx, cred, form)
}

func (c *PublishClient) CreateReelContainer(ctx context.Context, cred application.Credential, videoURL, caption string) (string, error) {
	form := url.Values{}
	form.Set("media_type", "REELS")
	form.Set("video_url", videoURL)
	form.Set("caption", caption)
	form.Set("access_token", cred.AccessToken)
	return c.createContainer(ctx, cred, form)
}

func (c *PublishClient) CreateCarouselItem(ctx context.Context, cred application.Credential, mediaURL string, isVideo bool) (string, error) {
	form := url.Values{}
	form.Set("is_carousel_item", "true")
	if isVideo {
		form.Set("media_type", "VIDEO")
		form.Set("video_url", mediaURL)
	} else {
		form.Set("image_url", mediaURL)
	}
	form.Set("access_token", cred.AccessToken)
	return c.createContainer(ctx, cred, form)
}

func (c *PublishClient) CreateCarouselContainer(ctx context.Context, cred application.Credential, childIDs []string, caption string) (string, error) {
	form := url.Values{}
	form.Set("media_type", "CAROUSEL")
	form.Set("children", strings.Join(childIDs, ","))
	form.Set("caption", caption)
	form.Set("access_token", cred.AccessToken)
	return c.createContainer(ctx, cred, form)
}

func (c *PublishClient) Publish(ctx context.Context, cred application.Credential, containerID string) (application.PublishedMedia, error) {
	form := url.Values{}
	form.Set("creation_id", containerID)
	form.Set("access_token", cred.AccessToken)

	var resp idResponse
	if err := c.postForm(ctx, c.baseURL+"/"+cred.InstagramUserID+"/media_publish", form, &resp); err != nil {
		return application.PublishedMedia{}, err
	}
	if resp.ID == "" {
		return application.PublishedMedia{}, fmt.Errorf("%w: empty media id", ErrGraphAPI)
	}

	permalink, _ := c.permalink(ctx, cred, resp.ID)
	return application.PublishedMedia{ID: resp.ID, Permalink: permalink}, nil
}

func (c *PublishClient) createContainer(ctx context.Context, cred application.Credential, form url.Values) (string, error) {
	var resp idResponse
	if err := c.postForm(ctx, c.baseURL+"/"+cred.InstagramUserID+"/media", form, &resp); err != nil {
		return "", err
	}
	if resp.ID == "" {
		return "", fmt.Errorf("%w: empty container id", ErrGraphAPI)
	}
	return resp.ID, nil
}

func (c *PublishClient) permalink(ctx context.Context, cred application.Credential, mediaID string) (string, error) {
	query := url.Values{}
	query.Set("fields", "permalink")
	query.Set("access_token", cred.AccessToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/"+mediaID+"?"+query.Encode(), nil)
	if err != nil {
		return "", err
	}

	var resp permalinkResponse
	if err := c.do(req, &resp); err != nil {
		return "", err
	}
	return resp.Permalink, nil
}

func (c *PublishClient) postForm(ctx context.Context, endpoint string, form url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("%w: build request: %v", ErrGraphAPI, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.do(req, out)
}

func (c *PublishClient) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrGraphAPI, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: read response: %v", ErrGraphAPI, err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%w: %s", ErrGraphAPI, graphErrorMessage(body, resp.StatusCode))
	}

	if out != nil {
		if err := json.Unmarshal(body, out); err != nil {
			return fmt.Errorf("%w: decode response: %v", ErrGraphAPI, err)
		}
	}
	return nil
}

func graphErrorMessage(body []byte, status int) string {
	var parsed errorResponse
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Error.Message != "" {
		return parsed.Error.Message
	}
	return fmt.Sprintf("graph api returned status %d", status)
}

type idResponse struct {
	ID string `json:"id"`
}

type permalinkResponse struct {
	Permalink string `json:"permalink"`
}

type errorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    int    `json:"code"`
	} `json:"error"`
}
