package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/aaronflorey/pupdate/internal/repolock"
	"github.com/spf13/cobra"
)

func TestExecuteHookForegroundDelegatesToRun(t *testing.T) {
	previousExecuteRunFn := executeRunFn
	t.Cleanup(func() { executeRunFn = previousExecuteRunFn })

	called := 0
	executeRunFn = func(_ *cobra.Command, quiet, allowScripts, dryRun bool) error {
		called++
		if !quiet || allowScripts || dryRun {
			t.Fatalf("unexpected hook options: quiet=%t allow=%t dry=%t", quiet, allowScripts, dryRun)
		}
		return nil
	}
	cmd := &cobra.Command{}
	cmd.SetContext(contextWithDependencies(context.Background(), testDependencies()))
	if err := executeHook(cmd, true, false, false, ""); err != nil {
		t.Fatalf("execute hook: %v", err)
	}
	if called != 1 {
		t.Fatalf("expected foreground hook to invoke run once, got %d", called)
	}
}

func TestLaunchBackgroundHookOnlyLaunchesChild(t *testing.T) {
	dir := t.TempDir()
	withChdir(t, dir)
	previousResolve := resolveExecutablePath
	previousStart := startBackgroundProcess
	t.Cleanup(func() {
		resolveExecutablePath = previousResolve
		startBackgroundProcess = previousStart
	})
	resolveExecutablePath = func() (string, error) { return "/fake/bin/pupdate", nil }

	var executable string
	var args []string
	startBackgroundProcess = func(gotExecutable string, gotArgs []string) (int, error) {
		executable = gotExecutable
		args = append([]string(nil), gotArgs...)
		return 1234, nil
	}
	cmd := &cobra.Command{}
	cmd.SetErr(&bytes.Buffer{})
	if err := launchBackgroundHook(cmd, true, testDependencies()); err != nil {
		t.Fatalf("launch background hook: %v", err)
	}
	if executable != "/fake/bin/pupdate" || !slices.Equal(args, []string{"hook", "--quiet", "--child"}) {
		t.Fatalf("unexpected child launch: executable=%q args=%#v", executable, args)
	}
	if _, err := os.Stat(filepath.Join(dir, repolock.FileName)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("async parent must not create the repository lock, err=%v", err)
	}
}

func TestLaunchBackgroundHookReturnsStartFailure(t *testing.T) {
	previousResolve := resolveExecutablePath
	previousStart := startBackgroundProcess
	t.Cleanup(func() {
		resolveExecutablePath = previousResolve
		startBackgroundProcess = previousStart
	})
	resolveExecutablePath = func() (string, error) { return "/fake/bin/pupdate", nil }
	startBackgroundProcess = func(string, []string) (int, error) { return 0, errors.New("boom") }
	err := launchBackgroundHook(&cobra.Command{}, true, testDependencies())
	if err == nil || !strings.Contains(err.Error(), "failed to start background hook: boom") {
		t.Fatalf("unexpected launch error: %v", err)
	}
}

func TestRepositoryLockTreatsUnlockedLegacyPIDFileAsIdle(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, repolock.FileName)
	legacy := []byte("123\npid=999999\n")
	if err := os.WriteFile(lockPath, legacy, 0o600); err != nil {
		t.Fatalf("write legacy lock: %v", err)
	}

	lock, acquired, err := repolock.Acquire(dir)
	if err != nil || !acquired {
		t.Fatalf("acquire unlocked legacy lock: acquired=%t err=%v", acquired, err)
	}
	if err := lock.Release(); err != nil {
		t.Fatalf("release lock: %v", err)
	}
	raw, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("read persistent lock: %v", err)
	}
	if !bytes.Equal(raw, legacy) {
		t.Fatalf("expected legacy contents to remain irrelevant, got %q", raw)
	}
	_, status, err := currentBackgroundHookStatus(dir)
	if err != nil || status != "idle" {
		t.Fatalf("expected released legacy file to be idle, status=%q err=%v", status, err)
	}
}

func TestRepositoryLockReportsContentionAndPersists(t *testing.T) {
	dir := t.TempDir()
	first, acquired, err := repolock.Acquire(dir)
	if err != nil || !acquired {
		t.Fatalf("acquire first lock: acquired=%t err=%v", acquired, err)
	}
	_, acquired, err = repolock.Acquire(dir)
	if err != nil || acquired {
		t.Fatalf("expected second acquisition to contend: acquired=%t err=%v", acquired, err)
	}
	_, status, err := currentBackgroundHookStatus(dir)
	if err != nil || status != "active" {
		t.Fatalf("expected active status, status=%q err=%v", status, err)
	}
	if err := first.Release(); err != nil {
		t.Fatalf("release first lock: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, repolock.FileName)); err != nil {
		t.Fatalf("coordination file should remain after release: %v", err)
	}
}

func TestManualRunAndResetFailOnRepositoryLockContention(t *testing.T) {
	dir := t.TempDir()
	withChdir(t, dir)
	lock, acquired, err := repolock.Acquire(dir)
	if err != nil || !acquired {
		t.Fatalf("acquire lock: acquired=%t err=%v", acquired, err)
	}
	t.Cleanup(func() { _ = lock.Release() })

	for _, args := range [][]string{{"run"}, {"reset"}} {
		cmd := newTestRootCmd()
		cmd.SetArgs(args)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		err := cmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "another pupdate operation is already active") {
			t.Fatalf("expected %v to fail on contention, got %v", args, err)
		}
	}
}

func TestQuietHookSkipsRepositoryLockContention(t *testing.T) {
	dir := t.TempDir()
	withChdir(t, dir)
	lock, acquired, err := repolock.Acquire(dir)
	if err != nil || !acquired {
		t.Fatalf("acquire lock: acquired=%t err=%v", acquired, err)
	}
	t.Cleanup(func() { _ = lock.Release() })

	var stderr bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"hook", "--quiet", "--child"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("quiet hook contention should be a successful skip: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("quiet hook contention should be silent, got %q", stderr.String())
	}
}

func TestExecuteHookQuietChildSuppressesSuccessfulUpdateStatus(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFiles(t, dir, "package-lock.json")
	withChdir(t, dir)
	t.Cleanup(func() {
		lookPath = exec.LookPath
		execCommand = exec.CommandContext
	})
	lookPath = func(file string) (string, error) { return file, nil }
	execCommand = func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "true")
	}

	var stderr bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"hook", "--quiet", "--child"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute quiet child hook: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected successful quiet child hook to be silent, got %q", stderr.String())
	}
}

func TestHookChildReturnsConfigParseErrorWithoutUsage(t *testing.T) {
	configHome := t.TempDir()
	projectDir := t.TempDir()
	writeFixtureFiles(t, configHome, filepath.Join("pupdate", "config.yaml"))
	configPath := filepath.Join(configHome, "pupdate", "config.yaml")
	if err := os.WriteFile(configPath, []byte("root_directories: [oops\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", configHome)
	withChdir(t, projectDir)
	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"hook", "--child"})
	assertExecuteCmdShowsErrorWithoutUsage(t, cmd, "pupdate: error: failed to parse "+configPath)
}
