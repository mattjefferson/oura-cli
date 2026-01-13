package app

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/mattjefferson/oura-cli/internal/config"
	"github.com/mattjefferson/oura-cli/internal/oura"
	"github.com/mattjefferson/oura-cli/internal/output"
	termutil "github.com/mattjefferson/oura-cli/internal/term"
)

type loadedConfig struct {
	Path string
	Cfg  config.Config
	Env  config.EnvOverrides
}

func loadConfig(opts GlobalOptions) (loadedConfig, error) {
	path := opts.ConfigPath
	if path == "" {
		p, err := config.DefaultPath()
		if err != nil {
			return loadedConfig{}, err
		}
		path = p
	}
	cfg, _, err := config.Load(path)
	if err != nil {
		return loadedConfig{}, err
	}
	env := config.ApplyEnv(&cfg)
	return loadedConfig{Path: path, Cfg: cfg, Env: env}, nil
}

func loadClient(opts GlobalOptions, printer *output.Printer) (*oura.Client, int, error) {
	loaded, err := loadConfig(opts)
	if err != nil {
		return nil, 1, err
	}
	if loaded.Cfg.Token == nil || loaded.Cfg.Token.AccessToken == "" {
		return nil, 3, errors.New("not authenticated")
	}
	client := oura.NewClient(&loaded.Cfg, loaded.Path, loaded.Env, opts.Timeout, printer)
	return client, 0, nil
}

func promptString(prompt string, out *output.Printer, noInput bool) (string, error) {
	if noInput {
		return "", errors.New("input disabled")
	}
	if !termutil.IsTTY(os.Stdin) {
		return "", errors.New("stdin is not a tty")
	}
	out.WriteErr(prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func promptSecret(prompt string, out *output.Printer, noInput bool) (string, error) {
	if noInput {
		return "", errors.New("input disabled")
	}
	if !termutil.IsTTY(os.Stdin) {
		return "", errors.New("stdin is not a tty")
	}
	return termutil.ReadPassword(prompt, os.Stderr)
}

func readSecretFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func parseScopes(s string) []string {
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

func exitCodeForStatus(status int) int {
	switch status {
	case 400:
		return 2
	case 401:
		return 3
	case 429:
		return 5
	default:
		if status >= 400 {
			return 1
		}
		return 0
	}
}

func parseDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}

func parseDateTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func formatDateTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func selectValue(flagValue string, fallback string) string {
	if strings.TrimSpace(flagValue) != "" {
		return strings.TrimSpace(flagValue)
	}
	return fallback
}

func statusSummary(cfg config.Config) map[string]any {
	result := map[string]any{
		"client_id":     cfg.ClientID,
		"redirect_uri":  cfg.RedirectURI,
		"scopes":        cfg.Scopes,
		"token_present": cfg.Token != nil && cfg.Token.AccessToken != "",
	}
	if cfg.Token != nil {
		if cfg.Token.ExpiresAt != "" {
			result["expires_at"] = cfg.Token.ExpiresAt
		}
		if cfg.Token.TokenType != "" {
			result["token_type"] = cfg.Token.TokenType
		}
	}
	return result
}

func apiErrorMessage(body []byte) string {
	text := strings.TrimSpace(string(body))
	if text == "" {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err == nil {
		if msg, ok := payload["message"].(string); ok && msg != "" {
			return msg
		}
		if msg, ok := payload["error"].(string); ok && msg != "" {
			return msg
		}
	}
	return text
}
