package main

import "fmt"

func initSnippetForShell(shell string) (string, error) {
	switch shell {
	case "bash":
		return bashInitSnippet, nil
	case "zsh":
		return zshInitSnippet, nil
	case "fish":
		return fishInitSnippet, nil
	default:
		return "", fmt.Errorf("unsupported shell %q; supported shells: %s", shell, supportedInitShellsText())
	}
}
