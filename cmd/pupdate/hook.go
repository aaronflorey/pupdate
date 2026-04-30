package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	backgroundHookLockFileName = ".pupdate.hook.lock"
	backgroundHookStaleAfter   = 10 * time.Minute
)

var backgroundHookNow = time.Now
var resolveExecutablePath = os.Executable
var executeRunFn = executeRun
var startBackgroundProcess = func(executable string, args []string, stderr io.Writer) error {
	cmd := exec.Command(executable, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = stderr
	return cmd.Start()
}

func newHookCmd() *cobra.Command {
	var quiet bool
	var async bool
	var child bool
	var lockFile string

	cmd := &cobra.Command{
		Use:           "hook",
		Short:         "Run internal shell-hook flow",
		Hidden:        true,
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeHook(cmd, quiet, async, child, lockFile)
		},
	}

	cmd.Flags().BoolVar(&quiet, "quiet", false, "suppress no-op output and child command output")
	cmd.Flags().BoolVar(&async, "async", false, "launch the hook work in the background")
	cmd.Flags().BoolVar(&child, "child", false, "run as the detached background hook worker")
	cmd.Flags().StringVar(&lockFile, "lock-file", "", "background hook lock file path")
	return cmd
}

func executeHook(cmd *cobra.Command, quiet bool, async bool, child bool, lockFile string) error {
	if child {
		if lockFile == "" {
			return fmt.Errorf("background hook child requires --lock-file")
		}
		defer removeBackgroundHookLock(lockFile)
		return executeRunFn(cmd, quiet, false)
	}

	if !async {
		return executeRunFn(cmd, quiet, false)
	}

	return launchBackgroundHook(cmd, quiet)
}

func launchBackgroundHook(cmd *cobra.Command, quiet bool) error {
	lockPath := backgroundHookLockPath(".")
	claimed, err := claimBackgroundHookLock(lockPath, backgroundHookNow())
	if err != nil {
		return err
	}
	if !claimed {
		printStatus(cmd, quiet, "pupdate: skip repo (background run already active)")
		return nil
	}

	executable, err := resolveExecutablePath()
	if err != nil {
		removeBackgroundHookLock(lockPath)
		return fmt.Errorf("failed to resolve current executable: %w", err)
	}

	args := []string{"hook", "--quiet", "--child", "--lock-file", lockPath}
	if err := startBackgroundProcess(executable, args, cmd.ErrOrStderr()); err != nil {
		removeBackgroundHookLock(lockPath)
		return fmt.Errorf("failed to start background hook: %w", err)
	}

	return nil
}

func backgroundHookLockPath(root string) string {
	return filepath.Join(root, backgroundHookLockFileName)
}

func claimBackgroundHookLock(path string, now time.Time) (bool, error) {
	for attempt := 0; attempt < 2; attempt++ {
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err == nil {
			content := strconv.FormatInt(now.Unix(), 10) + "\n"
			if _, writeErr := file.WriteString(content); writeErr != nil {
				file.Close()
				removeBackgroundHookLock(path)
				return false, fmt.Errorf("failed to write background hook lock %s: %w", path, writeErr)
			}
			if closeErr := file.Close(); closeErr != nil {
				removeBackgroundHookLock(path)
				return false, fmt.Errorf("failed to close background hook lock %s: %w", path, closeErr)
			}
			return true, nil
		}
		if !os.IsExist(err) {
			return false, fmt.Errorf("failed to create background hook lock %s: %w", path, err)
		}

		stale, staleErr := backgroundHookLockIsStale(path, now)
		if staleErr != nil {
			return false, staleErr
		}
		if !stale {
			return false, nil
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return false, fmt.Errorf("failed to remove stale background hook lock %s: %w", path, err)
		}
	}

	return false, nil
}

func backgroundHookLockIsStale(path string, now time.Time) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, fmt.Errorf("failed to stat background hook lock %s: %w", path, err)
	}

	if now.Sub(info.ModTime()) > backgroundHookStaleAfter {
		return true, nil
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("failed to read background hook lock %s: %w", path, err)
	}

	if strings.TrimSpace(string(raw)) == "" {
		return true, nil
	}

	return false, nil
}

func removeBackgroundHookLock(path string) {
	if path == "" {
		return
	}
	_ = os.Remove(path)
}

func currentBackgroundHookStatus(root string, now time.Time) (string, string, error) {
	lockPath := backgroundHookLockPath(root)
	info, err := os.Stat(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return lockPath, "idle", nil
		}
		return lockPath, "", fmt.Errorf("failed to stat background hook lock %s: %w", lockPath, err)
	}

	if now.Sub(info.ModTime()) > backgroundHookStaleAfter {
		return lockPath, "stale", nil
	}

	return lockPath, "active", nil
}
