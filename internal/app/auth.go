package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/mattjefferson/oura-cli/internal/config"
	"github.com/mattjefferson/oura-cli/internal/oura"
	"github.com/mattjefferson/oura-cli/internal/output"
)

const defaultRedirectURI = "http://127.0.0.1:8797/callback"

func runAuth(printer *output.Printer, opts GlobalOptions, args []string) int {
	if len(args) == 0 {
		printer.Write(authUsage())
		return 0
	}
	if args[0] == "--help" || args[0] == "-h" {
		printer.Write(authUsage())
		return 0
	}
	switch args[0] {
	case "login":
		return runAuthLogin(printer, opts, args[1:])
	case "status":
		return runAuthStatus(printer, opts)
	case "logout":
		return runAuthLogout(printer, opts)
	default:
		printer.Errorf("unknown auth command: %s", args[0])
		printer.WriteErr("\n")
		printer.WriteErr(authUsage())
		return 2
	}
}

func runAuthLogin(printer *output.Printer, opts GlobalOptions, args []string) int {
	fs := flag.NewFlagSet("auth login", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var clientID string
	var clientSecretFile string
	var redirectURI string
	var scopes string
	var noOpen bool
	var paste bool
	var help bool

	fs.StringVar(&clientID, "client-id", "", "oauth client id")
	fs.StringVar(&clientSecretFile, "client-secret-file", "", "oauth client secret file")
	fs.StringVar(&redirectURI, "redirect-uri", "", "oauth redirect uri")
	fs.StringVar(&scopes, "scopes", "", "scopes")
	fs.BoolVar(&noOpen, "no-open", false, "do not open browser")
	fs.BoolVar(&paste, "paste", false, "prompt for code")
	fs.BoolVar(&help, "help", false, "show help")
	fs.BoolVar(&help, "h", false, "show help")

	if err := fs.Parse(args); err != nil {
		printer.Errorf("flag error: %v", err)
		printer.WriteErr("\n")
		printer.WriteErr(authLoginUsage())
		return 2
	}
	if help {
		printer.Write(authLoginUsage())
		return 0
	}

	loaded, err := loadConfig(opts)
	if err != nil {
		printer.Errorf("config load failed: %v", err)
		return 1
	}
	cfg := loaded.Cfg

	if clientID != "" {
		cfg.ClientID = clientID
	}
	if redirectURI != "" {
		cfg.RedirectURI = redirectURI
	}
	if scopes != "" {
		cfg.Scopes = parseScopes(scopes)
	}
	if clientSecretFile != "" {
		secret, err := readSecretFile(clientSecretFile)
		if err != nil {
			printer.Errorf("client secret file read failed: %v", err)
			return 1
		}
		if secret == "" {
			printer.Errorf("client secret file is empty")
			return 1
		}
		cfg.ClientSecret = secret
	}

	if cfg.RedirectURI == "" {
		cfg.RedirectURI = defaultRedirectURI
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{"daily"}
	}

	if cfg.ClientID == "" {
		val, err := promptString("Client ID: ", printer, opts.NoInput)
		if err != nil {
			printer.Errorf("client id required: %v", err)
			return 2
		}
		if strings.TrimSpace(val) == "" {
			printer.Errorf("client id required")
			return 2
		}
		cfg.ClientID = val
	}
	if cfg.ClientSecret == "" {
		val, err := promptSecret("Client secret: ", printer, opts.NoInput)
		if err != nil {
			printer.Errorf("client secret required: %v", err)
			return 2
		}
		if strings.TrimSpace(val) == "" {
			printer.Errorf("client secret required")
			return 2
		}
		cfg.ClientSecret = val
	}

	state, err := randomState(24)
	if err != nil {
		printer.Errorf("state generation failed: %v", err)
		return 1
	}
	authURL, err := oura.BuildAuthURL(cfg.ClientID, cfg.RedirectURI, state, cfg.Scopes)
	if err != nil {
		printer.Errorf("auth url build failed: %v", err)
		return 1
	}

	var code string
	if paste {
		printer.Infof("Open this URL:\n%s", authURL)
		val, err := promptString("Authorization code: ", printer, opts.NoInput)
		if err != nil {
			printer.Errorf("authorization code required: %v", err)
			return 2
		}
		code = val
	} else {
		if err := validateLoopbackRedirect(cfg.RedirectURI); err != nil {
			printer.Errorf("redirect uri invalid for local server: %v", err)
			printer.Errorf("use --paste for manual flow")
			return 2
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		codeCh := make(chan string, 1)
		errCh := make(chan error, 1)
		go func() {
			code, err := waitForAuthCode(ctx, cfg.RedirectURI, state)
			if err != nil {
				errCh <- err
				return
			}
			codeCh <- code
		}()

		if !noOpen {
			openBrowser(authURL, printer)
		} else {
			printer.Infof("Open this URL:\n%s", authURL)
		}

		select {
		case code = <-codeCh:
		case err := <-errCh:
			printer.Errorf("auth failed: %v", err)
			return 1
		case <-ctx.Done():
			printer.Errorf("auth timed out")
			return 1
		}
	}

	httpClient := &http.Client{Timeout: opts.Timeout}
	tokenResp, err := oura.ExchangeCode(context.Background(), httpClient, cfg.ClientID, cfg.ClientSecret, cfg.RedirectURI, code)
	if err != nil {
		printer.Errorf("token exchange failed: %v", err)
		return 1
	}

	token := &config.Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
	}
	if tokenResp.ExpiresIn > 0 {
		expiresAt := time.Now().UTC().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		token.ExpiresAt = expiresAt.Format(time.RFC3339)
	}
	cfg.Token = token

	if err := config.Save(loaded.Path, cfg); err != nil {
		printer.Errorf("config save failed: %v", err)
		return 1
	}

	printer.Infof("auth complete; token saved")
	if token.ExpiresAt != "" {
		printer.Infof("token expires at %s", token.ExpiresAt)
	}
	return 0
}

func runAuthStatus(printer *output.Printer, opts GlobalOptions) int {
	loaded, err := loadConfig(opts)
	if err != nil {
		printer.Errorf("config load failed: %v", err)
		return 1
	}
	summary := statusSummary(loaded.Cfg)
	if opts.JSON {
		b, err := json.Marshal(summary)
		if err != nil {
			printer.Errorf("json encode failed: %v", err)
			return 1
		}
		if err := printer.PrintJSON(b); err != nil {
			printer.Errorf("output failed: %v", err)
			return 1
		}
		return 0
	}
	if summary["token_present"].(bool) {
		printer.Infof("logged in")
	} else {
		printer.Infof("not logged in")
		return 3
	}
	if v, ok := summary["expires_at"]; ok {
		printer.Infof("expires at %s", v)
	}
	if len(loaded.Cfg.Scopes) > 0 {
		printer.Infof("scopes: %s", strings.Join(loaded.Cfg.Scopes, " "))
	}
	if loaded.Cfg.ClientID != "" {
		printer.Infof("client id: %s", loaded.Cfg.ClientID)
	}
	if loaded.Cfg.RedirectURI != "" {
		printer.Infof("redirect uri: %s", loaded.Cfg.RedirectURI)
	}
	return 0
}

func runAuthLogout(printer *output.Printer, opts GlobalOptions) int {
	loaded, err := loadConfig(opts)
	if err != nil {
		printer.Errorf("config load failed: %v", err)
		return 1
	}
	if loaded.Cfg.Token == nil || loaded.Cfg.Token.AccessToken == "" {
		printer.Infof("no stored token")
		return 0
	}
	loaded.Cfg.Token = nil
	if err := config.Save(loaded.Path, loaded.Cfg); err != nil {
		printer.Errorf("config save failed: %v", err)
		return 1
	}
	printer.Infof("logged out")
	return 0
}

func randomState(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func waitForAuthCode(ctx context.Context, redirectURI, state string) (string, error) {
	u, err := url.Parse(redirectURI)
	if err != nil {
		return "", err
	}
	path := u.Path
	if path == "" {
		path = "/"
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if errParam := q.Get("error"); errParam != "" {
			http.Error(w, "authorization failed", http.StatusBadRequest)
			select {
			case errCh <- fmt.Errorf("authorization error: %s", errParam):
			default:
			}
			return
		}
		if state != "" && q.Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			select {
			case errCh <- errors.New("state mismatch"):
			default:
			}
			return
		}
		code := q.Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			select {
			case errCh <- errors.New("missing code"):
			default:
			}
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = fmt.Fprintln(w, "Authorization complete. You can close this window.")
		select {
		case codeCh <- code:
		default:
		}
	})

	ln, err := net.Listen("tcp", u.Host)
	if err != nil {
		return "", err
	}
	srv := &http.Server{Addr: u.Host, Handler: mux}
	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	defer func() {
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = srv.Shutdown(ctxShutdown)
		cancel()
	}()

	select {
	case code := <-codeCh:
		return code, nil
	case err := <-errCh:
		return "", err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func validateLoopbackRedirect(redirectURI string) error {
	u, err := url.Parse(redirectURI)
	if err != nil {
		return err
	}
	if u.Scheme != "http" {
		return errors.New("redirect uri must be http")
	}
	host := u.Hostname()
	if host != "127.0.0.1" && host != "localhost" {
		return errors.New("redirect uri must use localhost or 127.0.0.1")
	}
	if u.Port() == "" {
		return errors.New("redirect uri must include port")
	}
	return nil
}

func openBrowser(url string, printer *output.Printer) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		printer.Infof("Open this URL:\n%s", url)
		return
	}
	if err := cmd.Start(); err != nil {
		printer.Infof("Open this URL:\n%s", url)
	}
}
