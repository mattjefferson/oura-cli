package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Token struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresAt    string `json:"expires_at,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
}

type Config struct {
	ClientID     string   `json:"client_id,omitempty"`
	ClientSecret string   `json:"client_secret,omitempty"`
	RedirectURI  string   `json:"redirect_uri,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
	Token        *Token   `json:"token,omitempty"`
}

type EnvOverrides struct {
	AccessToken  bool
	RefreshToken bool
	ClientID     bool
	ClientSecret bool
	RedirectURI  bool
	Scopes       bool
}

func DefaultPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "oura", "config.json"), nil
}

func Load(path string) (Config, bool, error) {
	var cfg Config
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, false, nil
		}
		return cfg, false, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return cfg, true, err
	}
	return cfg, true, nil
}

func Save(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
}

func ApplyEnv(cfg *Config) EnvOverrides {
	over := EnvOverrides{}
	if v := os.Getenv("OURA_CLIENT_ID"); v != "" {
		cfg.ClientID = v
		over.ClientID = true
	}
	if v := os.Getenv("OURA_CLIENT_SECRET"); v != "" {
		cfg.ClientSecret = v
		over.ClientSecret = true
	}
	if v := os.Getenv("OURA_REDIRECT_URI"); v != "" {
		cfg.RedirectURI = v
		over.RedirectURI = true
	}
	if v := os.Getenv("OURA_SCOPES"); v != "" {
		cfg.Scopes = splitScopes(v)
		over.Scopes = true
	}
	if v := os.Getenv("OURA_ACCESS_TOKEN"); v != "" {
		if cfg.Token == nil {
			cfg.Token = &Token{}
		}
		cfg.Token.AccessToken = v
		over.AccessToken = true
	}
	if v := os.Getenv("OURA_REFRESH_TOKEN"); v != "" {
		if cfg.Token == nil {
			cfg.Token = &Token{}
		}
		cfg.Token.RefreshToken = v
		over.RefreshToken = true
	}
	return over
}

func splitScopes(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\n' || r == '\t'
	})
	out := make([]string, 0, len(fields))
	seen := map[string]bool{}
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if seen[f] {
			continue
		}
		seen[f] = true
		out = append(out, f)
	}
	return out
}
