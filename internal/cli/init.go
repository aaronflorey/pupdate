package cli

import (
	"fmt"

	"github.com/aaronflorey/pupdate/internal/shellhook"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var shell string
	var mode string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Print shell integration script",
		Long:  "Print the shell hook snippet for bash, zsh, or fish.",
		Example: `eval "$(pupdate init --shell bash)"
eval "$(pupdate init --shell zsh)"
eval "$(pupdate init --shell fish)"
eval "$(pupdate init --shell bash --mode foreground)"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := shellhook.ResolveShell(shell)
			if err != nil {
				return err
			}

			resolvedMode, err := shellhook.ResolveMode(mode)
			if err != nil {
				return err
			}

			snippet, err := shellhook.SnippetForShell(resolved, resolvedMode)
			if err != nil {
				return err
			}

			_, err = fmt.Fprint(cmd.OutOrStdout(), snippet)
			return err
		},
	}

	cmd.Flags().StringVar(&shell, "shell", "", "shell to configure (bash, zsh, or fish)")
	cmd.Flags().StringVar(&mode, "mode", shellhook.ModeAsync, "hook execution mode (async default, or foreground)")
	cmd.AddCommand(newInitInstallCmd())
	cmd.AddCommand(newInitUninstallCmd())
	return cmd
}
