package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Printer struct {
	Stdout      io.Writer
	Stderr      io.Writer
	Quiet       bool
	Verbose     bool
	JSONCompact bool
	Pretty      bool
}

func New(stdout, stderr io.Writer, quiet, verbose, jsonCompact, pretty bool) *Printer {
	return &Printer{
		Stdout:      stdout,
		Stderr:      stderr,
		Quiet:       quiet,
		Verbose:     verbose,
		JSONCompact: jsonCompact,
		Pretty:      pretty,
	}
}

func (p *Printer) Infof(format string, args ...any) {
	if p.Quiet {
		return
	}
	fmt.Fprintf(p.Stderr, format+"\n", args...)
}

func (p *Printer) Debugf(format string, args ...any) {
	if p.Quiet || !p.Verbose {
		return
	}
	fmt.Fprintf(p.Stderr, format+"\n", args...)
}

func (p *Printer) Errorf(format string, args ...any) {
	fmt.Fprintf(p.Stderr, format+"\n", args...)
}

func (p *Printer) Write(s string) {
	fmt.Fprint(p.Stdout, s)
}

func (p *Printer) WriteErr(s string) {
	fmt.Fprint(p.Stderr, s)
}

func (p *Printer) PrintJSON(body []byte) error {
	trimmed := bytes.TrimSpace(body)
	if p.JSONCompact {
		_, err := fmt.Fprintln(p.Stdout, string(trimmed))
		return err
	}
	if p.Pretty {
		var out bytes.Buffer
		if err := json.Indent(&out, trimmed, "", "  "); err == nil {
			_, err = fmt.Fprintln(p.Stdout, out.String())
			return err
		}
	}
	_, err := fmt.Fprintln(p.Stdout, strings.TrimSpace(string(trimmed)))
	return err
}
