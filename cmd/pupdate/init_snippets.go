package main

import "fmt"

func initSnippetForShell(shell string, mode string) (string, error) {
	hookCommand := "pupdate hook --quiet"
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
