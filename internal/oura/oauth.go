package oura

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	authorizeURL = "https://cloud.ouraring.com/oauth/authorize"
	tokenURL     = "https://api.ouraring.com/oauth/token"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

func BuildAuthURL(clientID, redirectURI, state string, scopes []string) (string, error) {
	if clientID == "" {
		return "", errors.New("client_id required")
	}
	if redirectURI == "" {
		return "", errors.New("redirect_uri required")
	}
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	if state != "" {
		q.Set("state", state)
	}
	if len(scopes) > 0 {
		q.Set("scope", strings.Join(scopes, " "))
	}
	return authorizeURL + "?" + q.Encode(), nil
}

func ExchangeCode(ctx context.Context, client *http.Client, clientID, clientSecret, redirectURI, code string) (TokenResponse, error) {
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)
	values.Set("client_id", clientID)
	values.Set("client_secret", clientSecret)
	values.Set("redirect_uri", redirectURI)
	return postToken(ctx, client, values)
}

func RefreshToken(ctx context.Context, client *http.Client, clientID, clientSecret, refreshToken string) (TokenResponse, error) {
	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", refreshToken)
	values.Set("client_id", clientID)
	values.Set("client_secret", clientSecret)
	return postToken(ctx, client, values)
}

func postToken(ctx context.Context, client *http.Client, values url.Values) (TokenResponse, error) {
	var token TokenResponse
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return token, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return token, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return token, err
	}
	if resp.StatusCode >= 400 {
		return token, fmt.Errorf("token exchange failed: %s", strings.TrimSpace(string(body)))
	}
	if err := json.Unmarshal(body, &token); err != nil {
		return token, err
	}
	return token, nil
}
