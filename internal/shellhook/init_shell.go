package shellhook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var supportedInitShells = []string{"bash", "zsh", "fish"}

const (
	ModeForeground = "foreground"
	ModeAsync      = "async"
)

func ResolveShell(requested string) (string, error) {
	if requested != "" {
		resolved := strings.ToLower(requested)
		if isSupportedInitShell(resolved) {
			return resolved, nil
		}

		return "", fmt.Errorf("unsupported shell %q; supported shells: %s", requested, supportedInitShellsText())
	}

	shell := filepath.Base(strings.TrimSpace(os.Getenv("SHELL")))
	if shell == "" {
		return "bash", nil
	}

	resolved := strings.ToLower(shell)
	if isSupportedInitShell(resolved) {
		return resolved, nil
	}

	return "bash", nil
}

func supportedInitShellsText() string {
	return strings.Join(supportedInitShells, ", ")
}

func ResolveMode(requested string) (string, error) {
	resolved := strings.ToLower(strings.TrimSpace(requested))
	if resolved == "" {
		return ModeAsync, nil
	}

	switch resolved {
	case ModeForeground, ModeAsync:
		return resolved, nil
	default:
		return "", fmt.Errorf("unsupported hook mode %q; supported modes: %s, %s", requested, ModeForeground, ModeAsync)
	}
}

func isSupportedInitShell(shell string) bool {
	for _, supported := range supportedInitShells {
		if shell == supported {
			return true
		}
	}

	return false
}
