package term

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

func IsTTY(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

func ReadPassword(prompt string, out io.Writer) (string, error) {
	if prompt != "" {
		if _, err := fmt.Fprint(out, prompt); err != nil {
			return "", err
		}
	}
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	if _, err := fmt.Fprintln(out); err != nil {
		return "", err
	}
	return string(b), nil
}
