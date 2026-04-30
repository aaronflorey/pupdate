package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestExecuteHookForegroundDelegatesToRun(t *testing.T) {
	previousExecuteRunFn := executeRunFn
	t.Cleanup(func() {
		executeRunFn = previousExecuteRunFn
	})

	called := 0
	executeRunFn = func(cmd *cobra.Command, quietFlag bool, allowScriptsFlag bool) error {
		called++
		if !quietFlag {
			t.Fatal("expected quiet flag to propagate to run execution")
		}
		if allowScriptsFlag {
			t.Fatal("expected hook flow to preserve default allow-scripts behavior")
		}
		return nil
	}

	if err := executeHook(&cobra.Command{}, true, false, false, ""); err != nil {
		t.Fatalf("execute hook: %v", err)
	}
	if called != 1 {
		t.Fatalf("expected foreground hook to invoke run once, got %d", called)
	}
}

func TestExecuteHookAsyncSkipsWhenBackgroundRunAlreadyActive(t *testing.T) {
	dir := t.TempDir()
	withChdir(t, dir)
	lockPath := backgroundHookLockPath(".")
	if err := os.WriteFile(lockPath, []byte("busy\n"), 0o600); err != nil {
		t.Fatalf("write lock file: %v", err)
	}
	now := time.Now()
	if err := os.Chtimes(lockPath, now, now); err != nil {
		t.Fatalf("touch lock file: %v", err)
	}

	var stderr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&stderr)

	if err := executeHook(cmd, false, true, false, ""); err != nil {
		t.Fatalf("execute hook: %v", err)
	}
	if !strings.Contains(stderr.String(), "pupdate: skip repo (background run already active)") {
		t.Fatalf("expected overlap status message, got %q", stderr.String())
	}
}

func TestLaunchBackgroundHookStartsDetachedChild(t *testing.T) {
	dir := t.TempDir()
	withChdir(t, dir)

	previousResolveExecutablePath := resolveExecutablePath
	previousStartBackgroundProcess := startBackgroundProcess
	t.Cleanup(func() {
		resolveExecutablePath = previousResolveExecutablePath
		startBackgroundProcess = previousStartBackgroundProcess
	})

	resolveExecutablePath = func() (string, error) {
		return "/fake/bin/pupdate", nil
	}

	var startedExecutable string
	var startedArgs []string
	startBackgroundProcess = func(executable string, args []string, stderr io.Writer) (int, error) {
		startedExecutable = executable
		startedArgs = append([]string(nil), args...)
		return 1234, nil
	}

	cmd := &cobra.Command{}
	cmd.SetErr(&bytes.Buffer{})

	if err := launchBackgroundHook(cmd, true); err != nil {
		t.Fatalf("launch background hook: %v", err)
	}
	if startedExecutable != "/fake/bin/pupdate" {
		t.Fatalf("expected detached child to use current executable, got %q", startedExecutable)
	}
	if len(startedArgs) != 5 || startedArgs[0] != "hook" || startedArgs[1] != "--quiet" || startedArgs[2] != "--child" || startedArgs[3] != "--lock-file" {
		t.Fatalf("unexpected detached child args: %#v", startedArgs)
	}
	if startedArgs[4] != filepath.Join(".", backgroundHookLockFileName) {
		t.Fatalf("expected detached child lock file arg, got %#v", startedArgs)
	}
	lock, _, err := readBackgroundHookLock(filepath.Join(dir, backgroundHookLockFileName))
	if err != nil {
		t.Fatalf("expected background hook lock to be readable, err=%v", err)
	}
	if lock.PID != 1234 {
		t.Fatalf("expected background hook lock pid to match child pid, got %#v", lock)
	}
}

func TestLaunchBackgroundHookPreservesNonQuietChildMode(t *testing.T) {
	dir := t.TempDir()
	withChdir(t, dir)

	previousResolveExecutablePath := resolveExecutablePath
	previousStartBackgroundProcess := startBackgroundProcess
	t.Cleanup(func() {
		resolveExecutablePath = previousResolveExecutablePath
		startBackgroundProcess = previousStartBackgroundProcess
	})

	resolveExecutablePath = func() (string, error) {
		return "/fake/bin/pupdate", nil
	}

	var startedArgs []string
	startBackgroundProcess = func(executable string, args []string, stderr io.Writer) (int, error) {
		startedArgs = append([]string(nil), args...)
		return 4321, nil
	}

	cmd := &cobra.Command{}
	cmd.SetErr(&bytes.Buffer{})

	if err := launchBackgroundHook(cmd, false); err != nil {
		t.Fatalf("launch background hook: %v", err)
	}
	for _, arg := range startedArgs {
		if arg == "--quiet" {
			t.Fatalf("expected non-quiet async hook to preserve child verbosity, got %#v", startedArgs)
		}
	}
}

func TestLaunchBackgroundHookRemovesLockWhenStartFails(t *testing.T) {
	dir := t.TempDir()
	withChdir(t, dir)

	previousResolveExecutablePath := resolveExecutablePath
	previousStartBackgroundProcess := startBackgroundProcess
	t.Cleanup(func() {
		resolveExecutablePath = previousResolveExecutablePath
		startBackgroundProcess = previousStartBackgroundProcess
	})

	resolveExecutablePath = func() (string, error) {
		return "/fake/bin/pupdate", nil
	}
	startBackgroundProcess = func(string, []string, io.Writer) (int, error) {
		return 0, errors.New("boom")
	}

	cmd := &cobra.Command{}
	cmd.SetErr(&bytes.Buffer{})

	err := launchBackgroundHook(cmd, true)
	if err == nil {
		t.Fatalf("expected background hook launch to fail")
	}
	if !strings.Contains(err.Error(), "failed to start background hook: boom") {
		t.Fatalf("unexpected launch error: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, backgroundHookLockFileName)); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("expected failed launch to clean up lock file, err=%v", statErr)
	}
}

func TestClaimBackgroundHookLockTreatsOldLockAsStale(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, backgroundHookLockFileName)
	if err := os.WriteFile(lockPath, []byte("old\n"), 0o600); err != nil {
		t.Fatalf("write lock file: %v", err)
	}
	old := time.Now().Add(-backgroundHookStaleAfter - time.Minute)
	if err := os.Chtimes(lockPath, old, old); err != nil {
		t.Fatalf("touch lock file: %v", err)
	}

	claimed, err := claimBackgroundHookLock(lockPath, time.Now())
	if err != nil {
		t.Fatalf("claim lock: %v", err)
	}
	if !claimed {
		t.Fatal("expected stale lock to be replaced")
	}
}

func TestClaimBackgroundHookLockKeepsRunningProcessLockActive(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, backgroundHookLockFileName)
	old := time.Now().Add(-backgroundHookStaleAfter - time.Minute)
	if err := writeBackgroundHookLock(lockPath, backgroundHookLock{ClaimedAtUnix: old.Unix(), PID: os.Getpid()}); err != nil {
		t.Fatalf("write lock file: %v", err)
	}
	if err := os.Chtimes(lockPath, old, old); err != nil {
		t.Fatalf("touch lock file: %v", err)
	}

	claimed, err := claimBackgroundHookLock(lockPath, time.Now())
	if err != nil {
		t.Fatalf("claim lock: %v", err)
	}
	if claimed {
		t.Fatal("expected running process lock to remain active")
	}
}

func TestCurrentBackgroundHookStatusReportsRunningPidAsActive(t *testing.T) {
	dir := t.TempDir()
	old := time.Now().Add(-backgroundHookStaleAfter - time.Minute)
	lockPath := filepath.Join(dir, backgroundHookLockFileName)
	if err := writeBackgroundHookLock(lockPath, backgroundHookLock{ClaimedAtUnix: old.Unix(), PID: os.Getpid()}); err != nil {
		t.Fatalf("write lock file: %v", err)
	}
	if err := os.Chtimes(lockPath, old, old); err != nil {
		t.Fatalf("touch lock file: %v", err)
	}

	gotPath, status, err := currentBackgroundHookStatus(dir, time.Now())
	if err != nil {
		t.Fatalf("current background hook status: %v", err)
	}
	if gotPath != lockPath {
		t.Fatalf("expected lock path %q, got %q", lockPath, gotPath)
	}
	if status != "active" {
		t.Fatalf("expected running pid lock to stay active, got %q", status)
	}
}

func TestExecuteHookChildRemovesLockFileAfterRun(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, backgroundHookLockFileName)
	if err := os.WriteFile(lockPath, []byte("busy\n"), 0o600); err != nil {
		t.Fatalf("write lock file: %v", err)
	}

	previousExecuteRunFn := executeRunFn
	t.Cleanup(func() {
		executeRunFn = previousExecuteRunFn
	})
	executeRunFn = func(cmd *cobra.Command, quietFlag bool, allowScriptsFlag bool) error {
		return errors.New("run failed")
	}

	err := executeHook(&cobra.Command{}, true, false, true, lockPath)
	if err == nil || !strings.Contains(err.Error(), "run failed") {
		t.Fatalf("expected child hook to return run error, got %v", err)
	}
	if _, statErr := os.Stat(lockPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("expected child hook to remove lock file, err=%v", statErr)
	}
}
