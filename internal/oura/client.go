package oura

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mattjefferson/oura-cli/internal/config"
	"github.com/mattjefferson/oura-cli/internal/output"
)

const apiBaseURL = "https://api.ouraring.com"

type Client struct {
	cfg          *config.Config
	cfgPath      string
	envOverrides config.EnvOverrides
	httpClient   *http.Client
	printer      *output.Printer
}

type Response struct {
	Status int
	Body   []byte
}

func NewClient(cfg *config.Config, cfgPath string, envOverrides config.EnvOverrides, timeout time.Duration, printer *output.Printer) *Client {
	return &Client{
		cfg:          cfg,
		cfgPath:      cfgPath,
		envOverrides: envOverrides,
		httpClient:   &http.Client{Timeout: timeout},
		printer:      printer,
	}
}

func (c *Client) Get(ctx context.Context, path string, query url.Values) (Response, error) {
	return c.do(ctx, http.MethodGet, path, query)
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values) (Response, error) {
	var respData Response
	if c.cfg.Token == nil || c.cfg.Token.AccessToken == "" {
		return respData, errors.New("missing access token")
	}
	u := apiBaseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return respData, err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.Token.AccessToken)
	c.printer.Debugf("%s %s", method, u)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return respData, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return respData, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		if refreshed, err := c.tryRefresh(ctx); err == nil && refreshed {
			return c.do(ctx, method, path, query)
		}
	}

	respData.Status = resp.StatusCode
	respData.Body = body
	return respData, nil
}

func (c *Client) tryRefresh(ctx context.Context) (bool, error) {
	if c.cfg.Token == nil || c.cfg.Token.RefreshToken == "" {
		return false, nil
	}
	if c.cfg.ClientID == "" || c.cfg.ClientSecret == "" {
		return false, errors.New("missing client credentials for token refresh")
	}

	tokenResp, err := RefreshToken(ctx, c.httpClient, c.cfg.ClientID, c.cfg.ClientSecret, c.cfg.Token.RefreshToken)
	if err != nil {
		return false, err
	}
	c.cfg.Token.AccessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		c.cfg.Token.RefreshToken = tokenResp.RefreshToken
	}
	if tokenResp.ExpiresIn > 0 {
		expiresAt := time.Now().UTC().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		c.cfg.Token.ExpiresAt = expiresAt.Format(time.RFC3339)
	}
	if tokenResp.TokenType != "" {
		c.cfg.Token.TokenType = tokenResp.TokenType
	}

	if c.persistAllowed() {
		if err := config.Save(c.cfgPath, *c.cfg); err != nil {
			c.printer.Debugf("token refresh save failed: %v", err)
		}
	}

	return true, nil
}

func (c *Client) persistAllowed() bool {
	if c.envOverrides.AccessToken || c.envOverrides.RefreshToken {
		return false
	}
	return true
}

func BuildPath(sandbox bool, segment string) string {
	if sandbox {
		return fmt.Sprintf("/v2/sandbox/usercollection/%s", segment)
	}
	return fmt.Sprintf("/v2/usercollection/%s", segment)
}

func BuildDocumentPath(sandbox bool, segment, documentID string) string {
	escaped := url.PathEscape(documentID)
	if sandbox {
		return fmt.Sprintf("/v2/sandbox/usercollection/%s/%s", segment, escaped)
	}
	return fmt.Sprintf("/v2/usercollection/%s/%s", segment, escaped)
}

func BuildQuery(params map[string]string) url.Values {
	values := url.Values{}
	for k, v := range params {
		if strings.TrimSpace(v) == "" {
			continue
		}
		values.Set(k, v)
	}
	return values
}
