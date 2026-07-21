package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/freshness"
	"github.com/aaronflorey/pupdate/internal/state"
	"github.com/spf13/cobra"
)

func TestCLIIntegrationRunExecutesInstallThroughRootCommand(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFiles(t, dir, "package-lock.json")
	withChdir(t, dir)

	previousLookPath := lookPath
	previousExecCommand := execCommand
	t.Cleanup(func() {
		lookPath = previousLookPath
		execCommand = previousExecCommand
	})

	installCalls := 0
	lookPath = func(file string) (string, error) {
		if file != "npm" {
			return "", exec.ErrNotFound
		}
		return "/fake/bin/npm", nil
	}
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		installCalls++
		if name != "npm" {
			t.Fatalf("expected npm install command, got %q", name)
		}
		if strings.Join(args, " ") != "ci --ignore-scripts" {
			t.Fatalf("expected npm ci install args, got %#v", args)
		}
		return exec.CommandContext(ctx, "true")
	}

	stdout, stderr, err := executeRootCommand(t, "run")
	if err != nil {
		t.Fatalf("run command failed: %v (stderr=%q)", err, stderr)
	}
	if stdout != "" {
		t.Fatalf("expected run command to avoid stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "pupdate: run npm ci --ignore-scripts") {
		t.Fatalf("expected run status line, got %q", stderr)
	}
	if !strings.Contains(stderr, "pupdate: done npm") {
		t.Fatalf("expected completion status line, got %q", stderr)
	}
	if installCalls != 1 {
		t.Fatalf("expected one install execution, got %d", installCalls)
	}

	stored, warnings, loadErr := state.NewStore(dir).Load()
	if loadErr != nil {
		t.Fatalf("load state: %v", loadErr)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected state warnings: %v", warnings)
	}
	if _, ok := stored.Ecosystems["node"]; !ok {
		t.Fatalf("expected node state to be persisted, got %#v", stored.Ecosystems)
	}
}

func TestCLIIntegrationStatusReportsReadyTargetThroughRootCommand(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFiles(t, dir, "bun.lock")
	withChdir(t, dir)

	previousDetectFn := detectFn
	previousEvaluateFreshnessFn := evaluateFreshnessFn
	previousLookPath := lookPath
	t.Cleanup(func() {
		detectFn = previousDetectFn
		evaluateFreshnessFn = previousEvaluateFreshnessFn
		lookPath = previousLookPath
	})

	detectFn = func(string, detection.Options) ([]detection.DetectionResult, error) {
		return []detection.DetectionResult{{
			Ecosystem:    detection.EcosystemNode,
			Managers:     []string{"bun"},
			MatchedFiles: []string{"bun.lock"},
		}}, nil
	}
	evaluateFreshnessFn = func(string, []detection.DetectionResult, state.FileState) ([]freshness.EcosystemDecision, error) {
		return []freshness.EcosystemDecision{{
			Ecosystem:        string(detection.EcosystemNode),
			StateKey:         "node",
			Decision:         freshness.DecisionUpdate,
			Reason:           "dependency lockfiles changed since last successful run",
			Lockfiles:        map[string]string{"bun.lock": "new"},
			LockfileMetadata: map[string]state.LockfileMetadata{"bun.lock": {Size: 1}},
		}}, nil
	}
	lookPath = func(file string) (string, error) {
		if file != "bun" {
			return "", exec.ErrNotFound
		}
		return "/fake/bin/bun", nil
	}

	stdout, stderr, err := executeRootCommand(t, "status")
	if err != nil {
		t.Fatalf("status command failed: %v (stderr=%q)", err, stderr)
	}
	if stderr != "" {
		t.Fatalf("expected status command to avoid stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "run_status: ready") {
		t.Fatalf("expected ready run status, got %q", stdout)
	}
	if !strings.Contains(stdout, "background_hook_lock_status: idle") {
		t.Fatalf("expected idle hook lock status, got %q", stdout)
	}
	if !strings.Contains(stdout, "install_status: ready") {
		t.Fatalf("expected ready install status, got %q", stdout)
	}
	if !strings.Contains(stdout, "manager_path: /fake/bin/bun") {
		t.Fatalf("expected stubbed manager path, got %q", stdout)
	}
	if !strings.Contains(stdout, "install_command: bun install --frozen-lockfile --ignore-scripts") {
		t.Fatalf("expected deterministic bun install command, got %q", stdout)
	}
}

func TestCLIIntegrationInitPrintsAsyncHookSnippet(t *testing.T) {
	stdout, stderr, err := executeRootCommand(t, "init", "--shell", "bash")
	if err != nil {
		t.Fatalf("init command failed: %v (stderr=%q)", err, stderr)
	}
	if stderr != "" {
		t.Fatalf("expected init command to avoid stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "PROMPT_COMMAND") {
		t.Fatalf("expected bash init snippet, got %q", stdout)
	}
	if !strings.Contains(stdout, expectedHookInvocation(t)+" --quiet --async") {
		t.Fatalf("expected default async hook snippet, got %q", stdout)
	}
}

func TestCLIIntegrationHookAsyncOnlyLaunchesChild(t *testing.T) {
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

	startCalls := 0
	var startedExecutable string
	var startedArgs []string
	startBackgroundProcess = func(executable string, args []string) (int, error) {
		startCalls++
		startedExecutable = executable
		startedArgs = append([]string(nil), args...)
		return 4242, nil
	}

	stdout, stderr, err := executeRootCommand(t, "hook", "--async", "--quiet")
	if err != nil {
		t.Fatalf("hook command failed: %v (stderr=%q)", err, stderr)
	}
	if stdout != "" || stderr != "" {
		t.Fatalf("expected async hook launch to avoid command output, stdout=%q stderr=%q", stdout, stderr)
	}
	if startCalls != 1 {
		t.Fatalf("expected one detached child launch, got %d", startCalls)
	}
	if startedExecutable != "/fake/bin/pupdate" {
		t.Fatalf("expected current executable path, got %q", startedExecutable)
	}
	if strings.Join(startedArgs, " ") != "hook --quiet --child" {
		t.Fatalf("unexpected detached child args: %#v", startedArgs)
	}
	if _, statErr := os.Stat(filepath.Join(dir, backgroundHookLockFileName)); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("async parent must not create a lock file, err=%v", statErr)
	}
}

func TestCLIIntegrationHookAsyncLaunchesEvenWhenWorkerLockIsActive(t *testing.T) {
	dir := t.TempDir()
	withChdir(t, dir)
	lock, acquired, err := acquireRepositoryLock(dir)
	if err != nil || !acquired {
		t.Fatalf("acquire worker lock: acquired=%t err=%v", acquired, err)
	}
	t.Cleanup(func() { _ = lock.release() })

	previousStartBackgroundProcess := startBackgroundProcess
	t.Cleanup(func() {
		startBackgroundProcess = previousStartBackgroundProcess
	})

	startCalls := 0
	startBackgroundProcess = func(string, []string) (int, error) {
		startCalls++
		return 0, nil
	}

	stdout, stderr, err := executeRootCommand(t, "hook", "--async", "--quiet")
	if err != nil {
		t.Fatalf("hook command failed: %v (stderr=%q)", err, stderr)
	}
	if stdout != "" || stderr != "" {
		t.Fatalf("expected launcher to remain silent, stdout=%q stderr=%q", stdout, stderr)
	}
	if startCalls != 1 {
		t.Fatalf("expected parent to launch a child that will arbitrate the lock, got %d calls", startCalls)
	}
}

func TestCLIIntegrationHookAsyncIgnoresLegacyPIDContents(t *testing.T) {
	dir := t.TempDir()
	withChdir(t, dir)

	lockPath := filepath.Join(dir, backgroundHookLockFileName)
	if err := os.WriteFile(lockPath, []byte("old\n"), 0o600); err != nil {
		t.Fatalf("write lock file: %v", err)
	}

	previousResolveExecutablePath := resolveExecutablePath
	previousStartBackgroundProcess := startBackgroundProcess
	t.Cleanup(func() {
		resolveExecutablePath = previousResolveExecutablePath
		startBackgroundProcess = previousStartBackgroundProcess
	})

	resolveExecutablePath = func() (string, error) {
		return "/fake/bin/pupdate", nil
	}
	startBackgroundProcess = func(string, []string) (int, error) {
		return 31337, nil
	}

	stdout, stderr, err := executeRootCommand(t, "hook", "--async", "--quiet")
	if err != nil {
		t.Fatalf("hook command failed: %v (stderr=%q)", err, stderr)
	}
	if stdout != "" || stderr != "" {
		t.Fatalf("expected legacy-lock launch to avoid command output, stdout=%q stderr=%q", stdout, stderr)
	}
	raw, readErr := os.ReadFile(lockPath)
	if readErr != nil || string(raw) != "old\n" {
		t.Fatalf("expected legacy contents to remain untouched, raw=%q err=%v", raw, readErr)
	}
}

func TestCLIIntegrationHookChildLeavesCoordinationFileInPlace(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, backgroundHookLockFileName)
	if err := os.WriteFile(lockPath, []byte("busy\n"), 0o600); err != nil {
		t.Fatalf("write lock file: %v", err)
	}

	previousExecuteRunFn := executeRunFn
	t.Cleanup(func() {
		executeRunFn = previousExecuteRunFn
	})

	called := 0
	executeRunFn = func(cmd *cobra.Command, quiet bool, allowScripts bool, dryRun bool) error {
		called++
		if !quiet {
			t.Fatal("expected child hook to preserve quiet flag")
		}
		if allowScripts {
			t.Fatal("expected child hook to preserve default allow-scripts behavior")
		}
		return errors.New("run failed")
	}

	_, _, err := executeRootCommand(t, "hook", "--child", "--quiet", "--lock-file", lockPath)
	if err == nil || !strings.Contains(err.Error(), "run failed") {
		t.Fatalf("expected child hook to return run error, got %v", err)
	}
	if called != 1 {
		t.Fatalf("expected one child run execution, got %d", called)
	}
	if _, statErr := os.Stat(lockPath); statErr != nil {
		t.Fatalf("expected child hook to leave coordination file in place, err=%v", statErr)
	}
}
