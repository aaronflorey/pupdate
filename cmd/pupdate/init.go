package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const bashInitSnippet = `# pupdate hook
_pupdate_last_pwd=""
_pupdate_hook() {
  if [ "$PWD" != "$_pupdate_last_pwd" ]; then
    _pupdate_last_pwd="$PWD"
    pupdate run --quiet
  fi
}
if [ -n "$PROMPT_COMMAND" ]; then
  PROMPT_COMMAND="_pupdate_hook;$PROMPT_COMMAND"
else
  PROMPT_COMMAND="_pupdate_hook"
fi
`

const zshInitSnippet = `# pupdate hook
autoload -U add-zsh-hook
_pupdate_last_pwd=""
_pupdate_hook() {
  if [ "$PWD" != "$_pupdate_last_pwd" ]; then
    _pupdate_last_pwd="$PWD"
    pupdate run --quiet
  fi
}
add-zsh-hook chpwd _pupdate_hook
add-zsh-hook precmd _pupdate_hook
`

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

			var snippet string
			switch resolved {
			case "bash":
				snippet = bashInitSnippet
			case "zsh":
				snippet = zshInitSnippet
			default:
				return fmt.Errorf("unsupported shell %q", resolved)
			}

			_, err = fmt.Fprint(cmd.OutOrStdout(), snippet)
			return err
		},
	}

	cmd.Flags().StringVar(&shell, "shell", "", "shell to configure (bash or zsh)")
	return cmd
}

func resolveShell(requested string) (string, error) {
	if requested != "" {
		resolved := strings.ToLower(requested)
		switch resolved {
		case "bash", "zsh":
			return resolved, nil
		default:
			return "", fmt.Errorf("unsupported shell %q; supported shells: bash, zsh", requested)
		}
	}

	shell := filepath.Base(strings.TrimSpace(os.Getenv("SHELL")))
	if shell == "" {
		return "bash", nil
	}

	resolved := strings.ToLower(shell)
	switch resolved {
	case "bash", "zsh":
		return resolved, nil
	default:
		return "bash", nil
	}
}
