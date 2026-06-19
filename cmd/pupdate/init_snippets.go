package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// resolvedExecutablePath returns the absolute, symlink-resolved path of the
// current binary. Falls back to "pupdate" if resolution fails.
func resolvedExecutablePath() string {
	exe, err := os.Executable()
	if err != nil {
		return "pupdate"
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return exe
	}
	return resolved
}

// shellQuote wraps s in single quotes for POSIX shell safety.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func initSnippetForShell(shell string, mode string) (string, error) {
	hookCommand := shellQuote(resolvedExecutablePath()) + " hook --quiet"
	if mode == hookModeAsync {
		hookCommand += " --async"
	}

	switch shell {
	case "bash":
		return fmt.Sprintf(bashInitSnippet, hookCommand), nil
	case "zsh":
		return fmt.Sprintf(zshInitSnippet, hookCommand), nil
	case "fish":
		return fmt.Sprintf(fishInitSnippet, hookCommand), nil
	default:
		return "", fmt.Errorf("unsupported shell %q; supported shells: %s", shell, supportedInitShellsText())
	}
}
