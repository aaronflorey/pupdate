package cli

import (
	"fmt"

	"github.com/aaronflorey/pupdate/internal/shellhook"
	"github.com/spf13/cobra"
)

func newInitInstallCmd() *cobra.Command {
	var shell string
	var mode string
	var target string
	var apply bool
	cmd := &cobra.Command{
		Use:     "install",
		Short:   "Preview or install a persistent shell hook",
		Long:    "Preview or install a managed pupdate block in a shell startup file. Changes are only written when --apply is provided.",
		Example: "pupdate init install --shell bash\npupdate init install --shell zsh --mode foreground --apply",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			resolvedShell, err := shellhook.ResolveShell(shell)
			if err != nil {
				return err
			}
			resolvedMode, err := shellhook.ResolveMode(mode)
			if err != nil {
				return err
			}
			resolvedTarget, err := shellhook.ResolveTarget(resolvedShell, target)
			if err != nil {
				return err
			}
			snippet, err := shellhook.SnippetForShell(resolvedShell, resolvedMode)
			if err != nil {
				return err
			}
			change, err := shellhook.PlanChange(resolvedTarget, shellhook.Install, snippet)
			if err != nil {
				return err
			}
			if !apply {
				return writeManagedHookPreview(cmd, change)
			}
			if err := shellhook.ApplyChange(change); err != nil {
				return err
			}
			return writeManagedHookResult(cmd, change)
		},
	}
	cmd.Flags().StringVar(&shell, "shell", "", "shell to configure (bash, zsh, or fish)")
	cmd.Flags().StringVar(&mode, "mode", shellhook.ModeAsync, "hook execution mode (async or foreground)")
	cmd.Flags().StringVar(&target, "target", "", "startup file to manage")
	cmd.Flags().BoolVar(&apply, "apply", false, "apply the proposed startup-file change")
	return cmd
}

func newInitUninstallCmd() *cobra.Command {
	var shell string
	var target string
	var apply bool
	cmd := &cobra.Command{
		Use:     "uninstall",
		Short:   "Preview or remove a persistent shell hook",
		Long:    "Preview or remove only the managed pupdate block from a shell startup file. Changes are only written when --apply is provided.",
		Example: "pupdate init uninstall --shell bash\npupdate init uninstall --shell fish --apply",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			resolvedShell, err := shellhook.ResolveShell(shell)
			if err != nil {
				return err
			}
			resolvedTarget, err := shellhook.ResolveTarget(resolvedShell, target)
			if err != nil {
				return err
			}
			change, err := shellhook.PlanChange(resolvedTarget, shellhook.Uninstall, "")
			if err != nil {
				return err
			}
			if !apply {
				return writeManagedHookPreview(cmd, change)
			}
			if err := shellhook.ApplyChange(change); err != nil {
				return err
			}
			return writeManagedHookResult(cmd, change)
		},
	}
	cmd.Flags().StringVar(&shell, "shell", "", "shell to configure (bash, zsh, or fish)")
	cmd.Flags().StringVar(&target, "target", "", "startup file to manage")
	cmd.Flags().BoolVar(&apply, "apply", false, "apply the proposed startup-file change")
	return cmd
}

func writeManagedHookPreview(cmd *cobra.Command, change shellhook.Change) error {
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "target: %s\noperation: %s\nchange: %s\napply: false\n",
		safeDisplay(change.Target), change.Operation, shellhook.ChangeLabel(change))
	return err
}

func writeManagedHookResult(cmd *cobra.Command, change shellhook.Change) error {
	if !change.Changed {
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "no pupdate hook change needed in %s\n", safeDisplay(change.Target))
		return err
	}
	verb := "installed"
	if change.Operation == shellhook.Uninstall {
		verb = "removed"
	}
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s pupdate hook in %s\n", verb, safeDisplay(change.Target))
	return err
}
