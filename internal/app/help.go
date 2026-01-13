package app

import (
	"strings"

	"github.com/mattjefferson/oura-cli/internal/output"
)

func runHelp(printer *output.Printer, args []string) int {
	if len(args) == 0 {
		printer.Write(rootUsage())
		return 0
	}
	if len(args) == 1 {
		switch args[0] {
		case "auth":
			printer.Write(authUsage())
			return 0
		case "list":
			printer.Write(listUsage())
			return 0
		case "get":
			printer.Write(getUsage())
			return 0
		case "resources":
			printer.Write(resourcesUsage())
			return 0
		case "whoami":
			printer.Write(whoamiUsage())
			return 0
		default:
			printer.Errorf("unknown command: %s", args[0])
			printer.WriteErr("\n")
			printer.WriteErr(rootUsage())
			return 2
		}
	}
	if len(args) >= 2 && args[0] == "auth" {
		switch args[1] {
		case "login":
			printer.Write(authLoginUsage())
			return 0
		case "status":
			printer.Write(authStatusUsage())
			return 0
		case "logout":
			printer.Write(authLogoutUsage())
			return 0
		default:
			printer.Errorf("unknown auth command: %s", args[1])
			printer.WriteErr("\n")
			printer.WriteErr(authUsage())
			return 2
		}
	}

	printer.Errorf("unknown help target: %s", strings.Join(args, " "))
	printer.WriteErr("\n")
	printer.WriteErr(rootUsage())
	return 2
}

func rootUsage() string {
	return `Usage:
  oura [global flags] <command> [args]

Commands:
  auth       OAuth2 login, status, logout
  list       List a resource collection
  get        Fetch a resource by id
  whoami     Fetch personal info
  resources  List available resources
  help       Show help for a command

Global flags:
  -h, --help           Show help
  --version            Print version
  -q, --quiet          Suppress non-data output
  -v, --verbose        Verbose logging
  --json               Compact JSON output
  --no-input           Disable prompts
  --config <path>      Config path (default ~/.config/oura/config.json)
  --timeout <dur>      HTTP timeout (default 30s)

Examples:
  oura auth login --scopes daily heartrate
  oura list sleep --start-date 2024-01-01 --end-date 2024-01-07
  oura get daily_activity <document_id>
  oura whoami
`
}

func authUsage() string {
	return `Usage:
  oura auth login [flags]
  oura auth status
  oura auth logout

Run:
  oura help auth login
`
}

func authLoginUsage() string {
	return `Usage:
  oura auth login [flags]

Flags:
  --client-id <id>           OAuth client id (env: OURA_CLIENT_ID)
  --client-secret-file <path> OAuth client secret file (env: OURA_CLIENT_SECRET)
  --redirect-uri <uri>       OAuth redirect URI (env: OURA_REDIRECT_URI)
  --scopes <list>            Scopes (space/comma-separated; default: daily)
  --no-open                  Do not open a browser
  --paste                    Print URL and prompt for the code

Notes:
  Redirect URI must match your Oura app settings.
`
}

func authStatusUsage() string {
	return `Usage:
  oura auth status
`
}

func authLogoutUsage() string {
	return `Usage:
  oura auth logout
`
}

func listUsage() string {
	return `Usage:
  oura list <resource> [flags]

Flags:
  --start-date <YYYY-MM-DD>
  --end-date <YYYY-MM-DD>
  --start-datetime <RFC3339>
  --end-datetime <RFC3339>
  --next-token <token>
  --sandbox
`
}

func getUsage() string {
	return `Usage:
  oura get <resource> [document_id] [flags]

Flags:
  --sandbox

Notes:
  personal_info does not require a document_id.
  heartrate is list-only.
`
}

func resourcesUsage() string {
	return `Usage:
  oura resources
`
}

func whoamiUsage() string {
	return `Usage:
  oura whoami
`
}
