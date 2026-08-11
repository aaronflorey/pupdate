package cli

import (
	"fmt"

	"github.com/aaronflorey/pupdate/internal/repolock"
	"github.com/spf13/cobra"
)

func newHookCmd() *cobra.Command {
	var quiet bool
	var async bool
	var child bool
	var lockFile string

	cmd := &cobra.Command{
		Use:          "hook",
		Short:        "Run internal shell-hook flow",
		Hidden:       true,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeHook(cmd, quiet, async, child, lockFile)
		},
	}

	cmd.Flags().BoolVar(&quiet, "quiet", false, "suppress no-op output and child command output")
	cmd.Flags().BoolVar(&async, "async", false, "launch the hook work in the background")
	cmd.Flags().BoolVar(&child, "child", false, "run as the detached background hook worker")
	cmd.Flags().StringVar(&lockFile, "lock-file", "", "legacy background hook lock file path")
	return cmd
}

func executeHook(cmd *cobra.Command, quiet bool, async bool, child bool, _ string) error {
	deps := dependenciesFor(cmd)
	if child || !async {
		return deps.Run(cmd, quiet, false, false)
	}
	return launchBackgroundHook(cmd, quiet, deps)
}

func launchBackgroundHook(cmd *cobra.Command, quiet bool, deps dependencies) error {
	executable, err := deps.ResolveExecutable()
	if err != nil {
		return fmt.Errorf("failed to resolve current executable: %w", err)
	}

	args := []string{"hook", "--child"}
	if quiet {
		args = []string{"hook", "--quiet", "--child"}
	}
	if _, err := deps.StartBackground(executable, args); err != nil {
		return fmt.Errorf("failed to start background hook: %w", err)
	}
	return nil
}

func currentBackgroundHookStatus(root string) (string, string, error) {
	lockPath := repolock.Path(root)
	active, err := repolock.Active(root)
	if err != nil {
		return lockPath, "", err
	}
	if active {
		return lockPath, "active", nil
	}
	return lockPath, "idle", nil
}
