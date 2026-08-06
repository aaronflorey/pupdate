package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

const backgroundHookLockFileName = ".pupdate.hook.lock"

var resolveExecutablePath = os.Executable
var executeRunFn = executeRun

var startBackgroundProcess = func(executable string, args []string) (int, error) {
	cmd := exec.Command(executable, args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	configureBackgroundHookCommand(cmd)
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	return cmd.Process.Pid, nil
}

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
	if child || !async {
		return executeRunFn(cmd, quiet, false, false)
	}
	return launchBackgroundHook(cmd, quiet)
}

func launchBackgroundHook(cmd *cobra.Command, quiet bool) error {
	executable, err := resolveExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to resolve current executable: %w", err)
	}

	args := []string{"hook", "--child"}
	if quiet {
		args = []string{"hook", "--quiet", "--child"}
	}
	if _, err := startBackgroundProcess(executable, args); err != nil {
		return fmt.Errorf("failed to start background hook: %w", err)
	}
	return nil
}

func backgroundHookLockPath(root string) string {
	return filepath.Join(root, backgroundHookLockFileName)
}

func currentBackgroundHookStatus(root string) (string, string, error) {
	lockPath := backgroundHookLockPath(root)
	active, err := repositoryLockActive(root)
	if err != nil {
		return lockPath, "", err
	}
	if active {
		return lockPath, "active", nil
	}
	return lockPath, "idle", nil
}
