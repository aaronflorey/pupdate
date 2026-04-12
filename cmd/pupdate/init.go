package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var shell string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Print shell integration script",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := resolveShell(shell)
			if err != nil {
				return err
			}

			snippet, err := initSnippetForShell(resolved)
			if err != nil {
				return err
			}

			_, err = fmt.Fprint(cmd.OutOrStdout(), snippet)
			return err
		},
	}

	cmd.Flags().StringVar(&shell, "shell", "", "shell to configure (bash, zsh, or fish)")
	return cmd
}

func resolveShell(requested string) (string, error) {
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
