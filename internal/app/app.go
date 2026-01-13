package app

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mattjefferson/oura-cli/internal/output"
	termutil "github.com/mattjefferson/oura-cli/internal/term"
)

type GlobalOptions struct {
	ConfigPath string
	Timeout    time.Duration
	JSON       bool
	Quiet      bool
	Verbose    bool
	NoInput    bool
	NoColor    bool
	Help       bool
	Version    bool
}

func Run(args []string) int {
	opts, rest, code := parseGlobalFlags(args)
	if code != 0 {
		return code
	}

	pretty := !opts.JSON && termutil.IsTTY(os.Stdout)
	printer := output.New(os.Stdout, os.Stderr, opts.Quiet, opts.Verbose, opts.JSON, pretty)

	if opts.Version {
		printer.Write(versionString())
		return 0
	}

	if opts.Help || len(rest) == 0 {
		printer.Write(rootUsage())
		return 0
	}

	switch rest[0] {
	case "help":
		return runHelp(printer, rest[1:])
	case "auth":
		return runAuth(printer, opts, rest[1:])
	case "list":
		return runList(printer, opts, rest[1:])
	case "get":
		return runGet(printer, opts, rest[1:])
	case "resources":
		return runResources(printer)
	case "whoami":
		return runWhoami(printer, opts)
	default:
		printer.Errorf("unknown command: %s", rest[0])
		printer.WriteErr("\n")
		printer.WriteErr(rootUsage())
		return 2
	}
}

func parseGlobalFlags(args []string) (GlobalOptions, []string, int) {
	var opts GlobalOptions
	fs := flag.NewFlagSet("oura", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&opts.ConfigPath, "config", "", "config file path")
	fs.DurationVar(&opts.Timeout, "timeout", 30*time.Second, "http timeout")
	fs.BoolVar(&opts.JSON, "json", false, "compact json output")
	fs.BoolVar(&opts.Quiet, "quiet", false, "suppress non-data output")
	fs.BoolVar(&opts.Quiet, "q", false, "suppress non-data output")
	fs.BoolVar(&opts.Verbose, "verbose", false, "verbose logging")
	fs.BoolVar(&opts.Verbose, "v", false, "verbose logging")
	fs.BoolVar(&opts.NoInput, "no-input", false, "disable prompts")
	fs.BoolVar(&opts.NoColor, "no-color", false, "disable color")
	fs.BoolVar(&opts.Help, "help", false, "show help")
	fs.BoolVar(&opts.Help, "h", false, "show help")
	fs.BoolVar(&opts.Version, "version", false, "show version")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "flag error: %v\n", err)
		fmt.Fprint(os.Stderr, rootUsage())
		return opts, nil, 2
	}

	return opts, fs.Args(), 0
}
