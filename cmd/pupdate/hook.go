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

type backgroundHookLock struct {
	ClaimedAtUnix int64
	PID           int
}

var startBackgroundProcess = func(executable string, args []string, stderr io.Writer) (int, error) {
	cmd := exec.Command(executable, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = stderr
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
		return executeRunFn(cmd, quiet, false, false)
	}

	if !async {
		return executeRunFn(cmd, quiet, false, false)
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

	args := []string{"hook", "--child", "--lock-file", lockPath}
	if quiet {
		args = []string{"hook", "--quiet", "--child", "--lock-file", lockPath}
	}
	pid, err := startBackgroundProcess(executable, args, cmd.ErrOrStderr())
	if err != nil {
		removeBackgroundHookLock(lockPath)
		return fmt.Errorf("failed to start background hook: %w", err)
	}
	if err := writeBackgroundHookLock(lockPath, backgroundHookLock{ClaimedAtUnix: backgroundHookNow().Unix(), PID: pid}); err != nil {
		removeBackgroundHookLock(lockPath)
		return err
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
			if _, writeErr := file.WriteString(formatBackgroundHookLock(backgroundHookLock{ClaimedAtUnix: now.Unix()})); writeErr != nil {
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
	lock, info, err := readBackgroundHookLock(path)
	if err != nil {
		return false, err
	}
	if info == nil {
		return true, nil
	}

	if lock.PID > 0 {
		running, runErr := backgroundHookProcessRunning(lock.PID)
		if runErr != nil {
			return false, runErr
		}
		if running {
			return false, nil
		}
		return true, nil
	}

	if now.Sub(info.ModTime()) > backgroundHookStaleAfter {
		return true, nil
	}

	if info.Size() == 0 {
		return true, nil
	}

	return false, nil
}

func writeBackgroundHookLock(path string, lock backgroundHookLock) error {
	if err := os.WriteFile(path, []byte(formatBackgroundHookLock(lock)), 0o600); err != nil {
		return fmt.Errorf("failed to write background hook lock %s: %w", path, err)
	}
	return nil
}

func formatBackgroundHookLock(lock backgroundHookLock) string {
	parts := []string{strconv.FormatInt(lock.ClaimedAtUnix, 10)}
	if lock.PID > 0 {
		parts = append(parts, fmt.Sprintf("pid=%d", lock.PID))
	}
	return strings.Join(parts, "\n") + "\n"
}

func readBackgroundHookLock(path string) (backgroundHookLock, os.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return backgroundHookLock{}, nil, nil
		}
		return backgroundHookLock{}, nil, fmt.Errorf("failed to stat background hook lock %s: %w", path, err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return backgroundHookLock{}, nil, fmt.Errorf("failed to read background hook lock %s: %w", path, err)
	}

	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	lock := backgroundHookLock{}
	if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
		claimedAt, parseErr := strconv.ParseInt(strings.TrimSpace(lines[0]), 10, 64)
		if parseErr == nil {
			lock.ClaimedAtUnix = claimedAt
		}
	}
	for _, line := range lines[1:] {
		if after, ok := strings.CutPrefix(strings.TrimSpace(line), "pid="); ok {
			pid, parseErr := strconv.Atoi(after)
			if parseErr == nil {
				lock.PID = pid
			}
		}
	}

	return lock, info, nil
}

func removeBackgroundHookLock(path string) {
	if path == "" {
		return
	}
	_ = os.Remove(path)
}

func currentBackgroundHookStatus(root string, now time.Time) (string, string, error) {
	lockPath := backgroundHookLockPath(root)
	lock, info, err := readBackgroundHookLock(lockPath)
	if err != nil {
		return lockPath, "", err
	}
	if info == nil {
		return lockPath, "idle", nil
	}

	if lock.PID > 0 {
		running, runErr := backgroundHookProcessRunning(lock.PID)
		if runErr != nil {
			return lockPath, "", runErr
		}
		if running {
			return lockPath, "active", nil
		}
		return lockPath, "stale", nil
	}

	if now.Sub(info.ModTime()) > backgroundHookStaleAfter || info.Size() == 0 {
		return lockPath, "stale", nil
	}

	return lockPath, "active", nil
}
