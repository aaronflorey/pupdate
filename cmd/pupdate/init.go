package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var shell string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Print shell integration script",
		Long:  "Print the shell hook snippet for bash, zsh, or fish.",
		Example: `eval "$(pupdate init --shell bash)"
eval "$(pupdate init --shell zsh)"
eval "$(pupdate init --shell fish)"`,
		Args: cobra.NoArgs,
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
