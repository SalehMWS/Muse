package meta

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/SalehMWS/Muse/internal/instagram/application"
	"github.com/SalehMWS/Muse/internal/shared/config"
	apperrors "github.com/SalehMWS/Muse/internal/shared/errors"
)

const defaultHTTPTimeout = 10 * time.Second

type OAuthClient struct {
	cfg        config.Instagram
	httpClient *http.Client
}

func NewOAuthClient(cfg config.Instagram) *OAuthClient {
	timeout := cfg.HTTPTimeout
	if timeout <= 0 {
		timeout = defaultHTTPTimeout
	}
	return &OAuthClient{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *OAuthClient) AuthorizationURL(state string) string {
	query := url.Values{}
	query.Set("client_id", c.cfg.ClientID)
	query.Set("redirect_uri", c.cfg.RedirectURI)
	query.Set("response_type", "code")
	query.Set("scope", c.cfg.Scopes)
	query.Set("state", state)
	return strings.TrimRight(c.cfg.AuthBaseURL, "/") + "/oauth/authorize?" + query.Encode()
}

func (c *OAuthClient) ExchangeCode(ctx context.Context, code string) (application.Token, error) {
	form := url.Values{}
	form.Set("client_id", c.cfg.ClientID)
	form.Set("client_secret", c.cfg.ClientSecret)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", c.cfg.RedirectURI)
	form.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(c.cfg.APIBaseURL, "/")+"/oauth/access_token",
		strings.NewReader(form.Encode()))
	if err != nil {
		return application.Token{}, apperrors.Wrap(err, apperrors.CodeExternalAPI, "build instagram token request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var short shortTokenResponse
	if err := c.do(req, &short); err != nil {
		return application.Token{}, err
	}

	exchange := url.Values{}
	exchange.Set("grant_type", "ig_exchange_token")
	exchange.Set("client_secret", c.cfg.ClientSecret)
	exchange.Set("access_token", short.AccessToken)

	longReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		strings.TrimRight(c.cfg.GraphBaseURL, "/")+"/access_token?"+exchange.Encode(), nil)
	if err != nil {
		return application.Token{}, apperrors.Wrap(err, apperrors.CodeExternalAPI, "build instagram long-lived token request")
	}

	var long longTokenResponse
	if err := c.do(longReq, &long); err != nil {
		return application.Token{}, err
	}

	return application.Token{
		AccessToken: long.AccessToken,
		ExpiresIn:   time.Duration(long.ExpiresIn) * time.Second,
	}, nil
}

func (c *OAuthClient) FetchProfile(ctx context.Context, accessToken string) (application.Profile, error) {
	query := url.Values{}
	query.Set("fields", "user_id,username,account_type")
	query.Set("access_token", accessToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		strings.TrimRight(c.cfg.GraphBaseURL, "/")+"/me?"+query.Encode(), nil)
	if err != nil {
		return application.Profile{}, apperrors.Wrap(err, apperrors.CodeExternalAPI, "build instagram profile request")
	}

	var resp profileResponse
	if err := c.do(req, &resp); err != nil {
		return application.Profile{}, err
	}

	userID := resp.UserID
	if userID == "" {
		userID = resp.ID
	}
	return application.Profile{
		UserID:      userID,
		Username:    resp.Username,
		AccountType: resp.AccountType,
	}, nil
}

func (c *OAuthClient) RefreshToken(ctx context.Context, accessToken string) (application.Token, error) {
	query := url.Values{}
	query.Set("grant_type", "ig_refresh_token")
	query.Set("access_token", accessToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		strings.TrimRight(c.cfg.GraphBaseURL, "/")+"/refresh_access_token?"+query.Encode(), nil)
	if err != nil {
		return application.Token{}, apperrors.Wrap(err, apperrors.CodeExternalAPI, "build instagram refresh request")
	}

	var long longTokenResponse
	if err := c.do(req, &long); err != nil {
		return application.Token{}, err
	}

	return application.Token{
		AccessToken: long.AccessToken,
		ExpiresIn:   time.Duration(long.ExpiresIn) * time.Second,
	}, nil
}

func (c *OAuthClient) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeExternalAPI, "instagram request failed")
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeExternalAPI, "read instagram response")
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return apperrors.Wrap(application.ErrInstagramAPI, apperrors.CodeExternalAPI, apiErrorMessage(body, resp.StatusCode))
	}

	if out != nil {
		if err := json.Unmarshal(body, out); err != nil {
			return apperrors.Wrap(err, apperrors.CodeExternalAPI, "decode instagram response")
		}
	}
	return nil
}

func apiErrorMessage(body []byte, status int) string {
	var parsed errorResponse
	_ = json.Unmarshal(body, &parsed)

	if parsed.Error.Message != "" {
		return parsed.Error.Message
	}
	if parsed.ErrorMessage != "" {
		return parsed.ErrorMessage
	}
	return fmt.Sprintf("instagram api returned status %d", status)
}

type shortTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type longTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type profileResponse struct {
	UserID      string `json:"user_id"`
	ID          string `json:"id"`
	Username    string `json:"username"`
	AccountType string `json:"account_type"`
}

type errorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    int    `json:"code"`
	} `json:"error"`
	ErrorType    string `json:"error_type"`
	ErrorMessage string `json:"error_message"`
}
