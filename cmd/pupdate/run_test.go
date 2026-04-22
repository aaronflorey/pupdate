package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/freshness"
	"github.com/aaronflorey/pupdate/internal/state"
)

func writeFixtureFiles(t *testing.T, dir string, files ...string) {
	t.Helper()
	for _, file := range files {
		path := filepath.Join(dir, file)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(file), err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("write %s: %v", file, err)
		}
	}
}

func disableInstall(t *testing.T) {
	t.Helper()
	t.Setenv("PUPDATE_SKIP_INSTALL", "1")
}

func withChdir(t *testing.T, dir string) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})
}

func hashFileForTest(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file for hash: %v", err)
	}
	sum := sha256.Sum256(raw)
	return fmt.Sprintf("%x", sum)
}

func assertSameDirectory(t *testing.T, expected string, actual string) {
	t.Helper()
	if sameDirectory(expected, actual) {
		return
	}
	t.Fatalf("expected install command to run in %q, got %q", expected, actual)
}

func TestRunManualModeUsesHumanReadableStatusWithoutStdout(t *testing.T) {
	disableInstall(t)
	dir := t.TempDir()
	writeFixtureFiles(t, dir,
		"bun.lock",
	)
	withChdir(t, dir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := newRunCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("run command failed: %v (stderr=%q)", err, stderr.String())
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected manual run to avoid stdout output, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "pupdate: installs disabled via PUPDATE_SKIP_INSTALL") {
		t.Fatalf("expected manual run status output, got %q", stderr.String())
	}
}

func TestRunReturnsDetectionFailedPrefixOnDetectorError(t *testing.T) {
	disableInstall(t)
	t.Cleanup(func() {
		detectFn = detection.Detect
	})
	detectFn = func(string) ([]detection.DetectionResult, error) {
		return nil, errors.New("boom")
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs([]string{"run"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected non-zero/errored execution")
	}
	if !strings.Contains(err.Error(), "detection failed:") {
		t.Fatalf("expected error prefix in returned error, got %q", err)
	}
	combined := stdout.String() + stderr.String()
	if !strings.Contains(combined, "detection failed:") && !strings.Contains(err.Error(), "detection failed:") {
		t.Fatalf("expected command output to include detection failure prefix; stdout=%q stderr=%q err=%q", stdout.String(), stderr.String(), err.Error())
	}
}

func TestRunPupignorePrintsSkipRepoAndSkipsInstalls(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFiles(t, dir, "bun.lock", ".pupignore")
	withChdir(t, dir)

	freshnessCalls := 0
	t.Cleanup(func() {
		evaluateFreshnessFn = freshness.Evaluate
	})
	evaluateFreshnessFn = func(string, []detection.DetectionResult, state.FileState) ([]freshness.EcosystemDecision, error) {
		freshnessCalls++
		return nil, errors.New("freshness should not run when .pupignore is present")
	}

	detectCalls := 0
	t.Cleanup(func() {
		detectFn = detection.Detect
	})
	detectFn = func(string) ([]detection.DetectionResult, error) {
		detectCalls++
		return nil, errors.New("detect should not run when .pupignore is present")
	}

	calls := 0
	t.Cleanup(func() {
		execCommand = exec.CommandContext
	})
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		calls++
		return exec.CommandContext(ctx, name, args...)
	}

	var stderr bytes.Buffer
	cmd := newRunCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("run command failed: %v", err)
	}

	if !strings.Contains(stderr.String(), "pupdate: skip repo (.pupignore)") {
		t.Fatalf("expected .pupignore skip status line, got %q", stderr.String())
	}
	if calls != 0 {
		t.Fatalf("expected no install execution when .pupignore is present, got %d calls", calls)
	}
	if detectCalls != 0 {
		t.Fatalf("expected no detection when .pupignore is present, got %d calls", detectCalls)
	}
	if freshnessCalls != 0 {
		t.Fatalf("expected no freshness evaluation when .pupignore is present, got %d calls", freshnessCalls)
	}
}

func TestHasPupIgnore(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		dir := t.TempDir()

		ignored, err := hasPupIgnore(dir)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if ignored {
			t.Fatal("expected missing .pupignore to be treated as not ignored")
		}
	})

	t.Run("regular file", func(t *testing.T) {
		dir := t.TempDir()
		writeFixtureFiles(t, dir, ".pupignore")

		ignored, err := hasPupIgnore(dir)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !ignored {
			t.Fatal("expected .pupignore file to disable run")
		}
	})

	t.Run("directory", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, ".pupignore"), 0o755); err != nil {
			t.Fatalf("make .pupignore directory: %v", err)
		}

		ignored, err := hasPupIgnore(dir)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if ignored {
			t.Fatal("expected .pupignore directory to not be treated as ignore marker")
		}
	})
}

func TestRunQuietSuppressesSkipStatusOnStderr(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFiles(t, dir, ".pupignore")
	withChdir(t, dir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := newRootCmd()
	cmd.SetArgs([]string{"run", "--quiet"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("run command failed: %v", err)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected quiet run to suppress stdout output, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected quiet run to stay silent when no update runs, got %q", stderr.String())
	}
}

func TestRunSkipsHomeDirectoryInManualAndQuietModes(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	withChdir(t, homeDir)

	var manualStdout bytes.Buffer
	var manualStderr bytes.Buffer
	manualCmd := newRootCmd()
	manualCmd.SetArgs([]string{"run"})
	manualCmd.SetOut(&manualStdout)
	manualCmd.SetErr(&manualStderr)
	if err := manualCmd.Execute(); err != nil {
		t.Fatalf("manual run failed: %v", err)
	}
	if manualStdout.Len() != 0 {
		t.Fatalf("expected manual home-directory skip to avoid stdout, got %q", manualStdout.String())
	}
	if !strings.Contains(manualStderr.String(), "pupdate: skip repo ($HOME)") {
		t.Fatalf("expected manual home-directory skip status, got %q", manualStderr.String())
	}

	var quietStdout bytes.Buffer
	var quietStderr bytes.Buffer
	quietCmd := newRootCmd()
	quietCmd.SetArgs([]string{"run", "--quiet"})
	quietCmd.SetOut(&quietStdout)
	quietCmd.SetErr(&quietStderr)
	if err := quietCmd.Execute(); err != nil {
		t.Fatalf("quiet run failed: %v", err)
	}
	if quietStdout.Len() != 0 {
		t.Fatalf("expected quiet home-directory skip to avoid stdout, got %q", quietStdout.String())
	}
	if quietStderr.Len() != 0 {
		t.Fatalf("expected quiet home-directory skip to stay silent, got %q", quietStderr.String())
	}
}

func TestRunSkipsHomeDirectoryWhenCurrentUserHomeFallbackMatches(t *testing.T) {
	homeDir := t.TempDir()
	withChdir(t, homeDir)

	t.Cleanup(func() {
		userHomeDir = os.UserHomeDir
		currentUserHomeDir = func() (string, error) {
			current, err := user.Current()
			if err != nil {
				return "", err
			}
			return current.HomeDir, nil
		}
	})

	userHomeDir = func() (string, error) {
		return filepath.Join(homeDir, "not-home"), nil
	}
	currentUserHomeDir = func() (string, error) {
		return homeDir, nil
	}

	var manualStdout bytes.Buffer
	var manualStderr bytes.Buffer
	manualCmd := newRootCmd()
	manualCmd.SetArgs([]string{"run"})
	manualCmd.SetOut(&manualStdout)
	manualCmd.SetErr(&manualStderr)
	if err := manualCmd.Execute(); err != nil {
		t.Fatalf("manual fallback run failed: %v", err)
	}
	if manualStdout.Len() != 0 {
		t.Fatalf("expected manual fallback home-directory skip to avoid stdout, got %q", manualStdout.String())
	}
	if !strings.Contains(manualStderr.String(), "pupdate: skip repo ($HOME)") {
		t.Fatalf("expected manual fallback home-directory skip status, got %q", manualStderr.String())
	}

	var quietStdout bytes.Buffer
	var quietStderr bytes.Buffer
	quietCmd := newRootCmd()
	quietCmd.SetArgs([]string{"run", "--quiet"})
	quietCmd.SetOut(&quietStdout)
	quietCmd.SetErr(&quietStderr)
	if err := quietCmd.Execute(); err != nil {
		t.Fatalf("quiet fallback run failed: %v", err)
	}
	if quietStdout.Len() != 0 {
		t.Fatalf("expected quiet fallback home-directory skip to avoid stdout, got %q", quietStdout.String())
	}
	if quietStderr.Len() != 0 {
		t.Fatalf("expected quiet fallback home-directory skip to stay silent, got %q", quietStderr.String())
	}
}

func TestRunSkipsOutsideConfiguredRootDirectory(t *testing.T) {
	configHome := t.TempDir()
	allowedRoot := filepath.Join(configHome, "workspace")
	projectDir := t.TempDir()
	writeFixtureFiles(t, configHome,
		filepath.Join("pupdate", "config.yaml"),
	)
	configPath := filepath.Join(configHome, "pupdate", "config.yaml")
	if err := os.WriteFile(configPath, []byte("root_directory: "+allowedRoot+"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", configHome)
	withChdir(t, projectDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs([]string{"run"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected restricted run to avoid stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "pupdate: skip repo (outside configured root_directory)") {
		t.Fatalf("expected configured-root skip status, got %q", stderr.String())
	}
	if strings.Contains(stderr.String(), "pupdate: installs disabled") {
		t.Fatalf("expected early skip before install flow, got %q", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".pupdate")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no state write when repo is outside configured root, err=%v", err)
	}

	var quietStdout bytes.Buffer
	var quietStderr bytes.Buffer
	quietCmd := newRootCmd()
	quietCmd.SetArgs([]string{"run", "--quiet"})
	quietCmd.SetOut(&quietStdout)
	quietCmd.SetErr(&quietStderr)
	if err := quietCmd.Execute(); err != nil {
		t.Fatalf("quiet run failed: %v", err)
	}
	if quietStdout.Len() != 0 {
		t.Fatalf("expected quiet restricted run to avoid stdout, got %q", quietStdout.String())
	}
	if quietStderr.Len() != 0 {
		t.Fatalf("expected quiet restricted run to stay silent, got %q", quietStderr.String())
	}
}

func TestRunAllowsProjectInsideConfiguredRootDirectoryWithHomeExpansion(t *testing.T) {
	homeDir := t.TempDir()
	projectDir := filepath.Join(homeDir, "src", "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}
	writeFixtureFiles(t, projectDir, "package-lock.json")
	writeFixtureFiles(t, filepath.Join(homeDir, ".config"), filepath.Join("pupdate", "config.yaml"))
	configPath := filepath.Join(homeDir, ".config", "pupdate", "config.yaml")
	if err := os.WriteFile(configPath, []byte("root_directory: ~/src\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(homeDir, ".config"))
	withChdir(t, projectDir)

	t.Cleanup(func() {
		lookPath = exec.LookPath
		execCommand = exec.CommandContext
	})
	lookPath = func(file string) (string, error) {
		return file, nil
	}
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "true")
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs([]string{"run", "--quiet"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected quiet run to avoid stdout, got %q", stdout.String())
	}
	if strings.Contains(stderr.String(), "outside configured root_directory") {
		t.Fatalf("expected configured root to allow project, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "pupdate: run npm ci --ignore-scripts") {
		t.Fatalf("expected install to run inside configured root, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "pupdate: done npm") {
		t.Fatalf("expected install completion inside configured root, got %q", stderr.String())
	}
}

func TestRunAllowsProjectInsideConfiguredRootDirectoryWhenSetToHomeShortcut(t *testing.T) {
	homeDir := t.TempDir()
	projectDir := filepath.Join(homeDir, "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}
	writeFixtureFiles(t, projectDir, "package-lock.json")
	writeFixtureFiles(t, filepath.Join(homeDir, ".config"), filepath.Join("pupdate", "config.yaml"))
	configPath := filepath.Join(homeDir, ".config", "pupdate", "config.yaml")
	if err := os.WriteFile(configPath, []byte("root_directory: ~\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(homeDir, ".config"))
	withChdir(t, projectDir)

	t.Cleanup(func() {
		lookPath = exec.LookPath
		execCommand = exec.CommandContext
	})
	lookPath = func(file string) (string, error) {
		return file, nil
	}
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "true")
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs([]string{"run", "--quiet"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected quiet run to avoid stdout, got %q", stdout.String())
	}
	if strings.Contains(stderr.String(), "outside configured root_directory") {
		t.Fatalf("expected root_directory=~ to allow project inside home, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "pupdate: run npm ci --ignore-scripts") {
		t.Fatalf("expected install to run inside root_directory=~, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "pupdate: done npm") {
		t.Fatalf("expected install completion inside root_directory=~, got %q", stderr.String())
	}
}

func TestRunReturnsParseErrorWhenYAMLIsInvalid(t *testing.T) {
	configHome := t.TempDir()
	projectDir := t.TempDir()
	writeFixtureFiles(t, configHome, filepath.Join("pupdate", "config.yaml"))
	configPath := filepath.Join(configHome, "pupdate", "config.yaml")
	if err := os.WriteFile(configPath, []byte("root_directory: [oops\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", configHome)
	withChdir(t, projectDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs([]string{"run"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected run command to fail")
	}
	if !strings.Contains(err.Error(), "failed to parse "+configPath) {
		t.Fatalf("expected parse error with config path, got %q", err.Error())
	}
}

func TestRunReturnsUserConfigDirResolutionError(t *testing.T) {
	dir := t.TempDir()
	withChdir(t, dir)

	t.Cleanup(func() {
		userConfigDir = os.UserConfigDir
		runtimeGOOS = runtime.GOOS
	})
	runtimeGOOS = "linux"
	userConfigDir = func() (string, error) {
		return "", errors.New("boom")
	}

	cmd := newRootCmd()
	cmd.SetArgs([]string{"run"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected run command to fail")
	}
	if !strings.Contains(err.Error(), "failed to resolve user config directory: boom") {
		t.Fatalf("expected config-dir resolution error, got %q", err.Error())
	}
}

func TestRunReturnsReadErrorWhenConfigPathIsDirectory(t *testing.T) {
	configHome := t.TempDir()
	projectDir := t.TempDir()
	configPath := filepath.Join(configHome, "pupdate", "config.yaml")
	if err := os.MkdirAll(configPath, 0o755); err != nil {
		t.Fatalf("mkdir config path: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", configHome)
	withChdir(t, projectDir)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"run"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected run command to fail")
	}
	if !strings.Contains(err.Error(), "failed to read "+configPath) {
		t.Fatalf("expected read error with config path, got %q", err.Error())
	}
}

func TestRunQuietPrintsRunAndDoneWhenUpdateRuns(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFiles(t, dir, "package-lock.json")
	withChdir(t, dir)

	t.Cleanup(func() {
		lookPath = exec.LookPath
		execCommand = exec.CommandContext
	})
	lookPath = func(file string) (string, error) {
		return file, nil
	}
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "true")
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs([]string{"run", "--quiet"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("quiet run command failed: %v", err)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected quiet run to avoid stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "pupdate: run npm ci --ignore-scripts") {
		t.Fatalf("expected quiet run line when update executes, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "pupdate: done npm") {
		t.Fatalf("expected quiet completion line when update succeeds, got %q", stderr.String())
	}
}

func TestRunPrintsSkipStatusForUnchangedEcosystem(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFiles(t, dir, "composer.lock")
	withChdir(t, dir)

	initial := state.Empty()
	initial.Ecosystems["php"] = state.EcosystemState{
		LastSuccessAt: "2026-03-01T12:00:00Z",
		Lockfiles: map[string]string{
			"composer.lock": hashFileForTest(t, filepath.Join(dir, "composer.lock")),
		},
	}
	if err := state.NewStore(dir).Save(initial); err != nil {
		t.Fatalf("seed state: %v", err)
	}

	t.Cleanup(func() {
		execCommand = exec.CommandContext
	})
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		t.Fatalf("install should not execute on unchanged lockfiles")
		return exec.CommandContext(ctx, name, args...)
	}

	var stderr bytes.Buffer
	cmd := newRunCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("run command failed: %v", err)
	}

	line := "pupdate: skip php (dependency lockfiles unchanged since last successful run)"
	if !strings.Contains(stderr.String(), line) {
		t.Fatalf("expected unchanged skip status line, got %q", stderr.String())
	}
}

func TestRunUpdatesDepthOneSubdirectoryAndSavesNamespacedState(t *testing.T) {
	dir := t.TempDir()
	frontendDir := filepath.Join(dir, "frontend")
	if err := os.Mkdir(frontendDir, 0o755); err != nil {
		t.Fatalf("mkdir frontend: %v", err)
	}
	writeFixtureFiles(t, frontendDir, "package-lock.json")
	withChdir(t, dir)

	ranDirFile := filepath.Join(dir, "ran_dir")
	t.Cleanup(func() {
		lookPath = exec.LookPath
		execCommand = exec.CommandContext
	})
	lookPath = func(file string) (string, error) {
		return file, nil
	}
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "pwd > \""+ranDirFile+"\"")
	}

	var stderr bytes.Buffer
	cmd := newRunCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("run command failed: %v", err)
	}

	ranDirRaw, err := os.ReadFile(ranDirFile)
	if err != nil {
		t.Fatalf("read ran_dir output: %v", err)
	}
	ranDir := strings.TrimSpace(string(ranDirRaw))
	assertSameDirectory(t, frontendDir, ranDir)
	if !strings.Contains(stderr.String(), "pupdate: run npm ci --ignore-scripts (in frontend)") {
		t.Fatalf("expected depth-1 run status line, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "pupdate: done npm (in frontend)") {
		t.Fatalf("expected depth-1 completion status line, got %q", stderr.String())
	}

	stored, warnings, err := state.NewStore(dir).Load()
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected state warnings: %v", warnings)
	}
	entry, ok := stored.Ecosystems["node@frontend"]
	if !ok {
		t.Fatalf("expected namespaced node state key, got %#v", stored.Ecosystems)
	}
	if entry.LastSuccessAt == "" {
		t.Fatalf("expected last_success_at for namespaced state key")
	}
	if _, ok := entry.Lockfiles["frontend/package-lock.json"]; !ok {
		t.Fatalf("expected namespaced state to include subdirectory lockfile hash, got %#v", entry.Lockfiles)
	}
}

func TestRunUpdatesPackagesChildAndSavesPackagesNamespacedState(t *testing.T) {
	dir := t.TempDir()
	packageDir := filepath.Join(dir, "packages", "web")
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		t.Fatalf("mkdir packages/web: %v", err)
	}
	writeFixtureFiles(t, packageDir, "package-lock.json")
	withChdir(t, dir)

	ranDirFile := filepath.Join(dir, "ran_dir")
	t.Cleanup(func() {
		lookPath = exec.LookPath
		execCommand = exec.CommandContext
	})
	lookPath = func(file string) (string, error) {
		return file, nil
	}
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "pwd > \""+ranDirFile+"\"")
	}

	var stderr bytes.Buffer
	cmd := newRunCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("run command failed: %v", err)
	}

	ranDirRaw, err := os.ReadFile(ranDirFile)
	if err != nil {
		t.Fatalf("read ran_dir output: %v", err)
	}
	ranDir := strings.TrimSpace(string(ranDirRaw))
	assertSameDirectory(t, packageDir, ranDir)
	if !strings.Contains(stderr.String(), "pupdate: run npm ci --ignore-scripts (in packages/web)") {
		t.Fatalf("expected packages child run status line, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "pupdate: done npm (in packages/web)") {
		t.Fatalf("expected packages child completion status line, got %q", stderr.String())
	}

	stored, warnings, err := state.NewStore(dir).Load()
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected state warnings: %v", warnings)
	}
	entry, ok := stored.Ecosystems["node@packages/web"]
	if !ok {
		t.Fatalf("expected namespaced packages state key, got %#v", stored.Ecosystems)
	}
	if entry.LastSuccessAt == "" {
		t.Fatalf("expected last_success_at for packages child state key")
	}
	if _, ok := entry.Lockfiles["packages/web/package-lock.json"]; !ok {
		t.Fatalf("expected namespaced packages state to include child lockfile hash, got %#v", entry.Lockfiles)
	}
}

func TestRunPrintsErrorStatusWhenInstallFails(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFiles(t, dir, "bun.lock")
	withChdir(t, dir)

	t.Cleanup(func() {
		execCommand = exec.CommandContext
		lookPath = exec.LookPath
	})
	lookPath = func(file string) (string, error) {
		return file, nil
	}
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "false")
	}

	var stderr bytes.Buffer
	cmd := newRunCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("run command failed: %v", err)
	}

	if !strings.Contains(stderr.String(), "pupdate: run bun install --frozen-lockfile --ignore-scripts") {
		t.Fatalf("expected run status line before install, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "pupdate: error bun install failed:") {
		t.Fatalf("expected install failure error status line, got %q", stderr.String())
	}
}

func TestSelectManagerPlanBunUsesSafeFlags(t *testing.T) {
	plan, ok, reason := selectManagerPlan(detection.DetectionResult{
		Ecosystem: detection.EcosystemNode,
		Managers:  []string{"bun"},
	}, false)

	if !ok {
		t.Fatalf("expected bun manager plan to be supported, reason=%q", reason)
	}
	if plan.Manager != "bun" {
		t.Fatalf("expected bun manager, got %q", plan.Manager)
	}
	if !slices.Equal(plan.Args, []string{"install", "--frozen-lockfile", "--ignore-scripts"}) {
		t.Fatalf("unexpected bun args: %#v", plan.Args)
	}
}

func TestSelectManagerPlanComposerUsesSafeFlags(t *testing.T) {
	plan, ok, reason := selectManagerPlan(detection.DetectionResult{
		Ecosystem: detection.EcosystemPHP,
	}, false)

	if !ok {
		t.Fatalf("expected composer manager plan to be supported, reason=%q", reason)
	}
	if plan.Manager != "composer" {
		t.Fatalf("expected composer manager, got %q", plan.Manager)
	}
	if !slices.Equal(plan.Args, []string{"install", "--no-interaction", "--prefer-dist", "--no-scripts"}) {
		t.Fatalf("unexpected composer args: %#v", plan.Args)
	}
}

func TestSelectManagerPlanUnsupportedStillSkips(t *testing.T) {
	_, ok, reason := selectManagerPlan(detection.DetectionResult{
		Ecosystem: detection.EcosystemNode,
		Managers:  []string{"npm", "pnpm"},
	}, false)

	if ok {
		t.Fatalf("expected unsupported manager to skip")
	}
	if !strings.Contains(reason, "multiple Node lockfiles detected") {
		t.Fatalf("expected explicit unsupported-manager reason, got %q", reason)
	}
}

func TestSelectManagerPlanExpandedManagersUseSafeFlags(t *testing.T) {
	tests := []struct {
		name    string
		result  detection.DetectionResult
		manager string
		args    []string
	}{
		{
			name:    "node npm",
			result:  detection.DetectionResult{Ecosystem: detection.EcosystemNode, Managers: []string{"npm"}},
			manager: "npm",
			args:    []string{"ci", "--ignore-scripts"},
		},
		{
			name:    "node pnpm",
			result:  detection.DetectionResult{Ecosystem: detection.EcosystemNode, Managers: []string{"pnpm"}},
			manager: "pnpm",
			args:    []string{"install", "--frozen-lockfile", "--ignore-scripts"},
		},
		{
			name:    "node yarn",
			result:  detection.DetectionResult{Ecosystem: detection.EcosystemNode, Managers: []string{"yarn"}},
			manager: "yarn",
			args:    []string{"install", "--frozen-lockfile", "--ignore-scripts"},
		},
		{
			name:    "python uv",
			result:  detection.DetectionResult{Ecosystem: detection.EcosystemPython, Managers: []string{"uv"}},
			manager: "uv",
			args:    []string{"sync", "--frozen"},
		},
		{
			name:    "python poetry",
			result:  detection.DetectionResult{Ecosystem: detection.EcosystemPython, Managers: []string{"poetry"}},
			manager: "poetry",
			args:    []string{"install", "--no-interaction", "--sync"},
		},
		{
			name:    "python pip",
			result:  detection.DetectionResult{Ecosystem: detection.EcosystemPython, Managers: []string{"pip"}},
			manager: "pip",
			args:    []string{"install", "-r", "requirements.txt", "--disable-pip-version-check", "--no-input"},
		},
		{
			name:    "go",
			result:  detection.DetectionResult{Ecosystem: detection.EcosystemGo, Managers: []string{"go"}},
			manager: "go",
			args:    []string{"mod", "download"},
		},
		{
			name:    "rust",
			result:  detection.DetectionResult{Ecosystem: detection.EcosystemRust, Managers: []string{"cargo"}},
			manager: "cargo",
			args:    []string{"fetch", "--locked"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, ok, reason := selectManagerPlan(tt.result, false)
			if !ok {
				t.Fatalf("expected manager plan to be supported, reason=%q", reason)
			}
			if plan.Manager != tt.manager {
				t.Fatalf("expected manager %q, got %q", tt.manager, plan.Manager)
			}
			if !slices.Equal(plan.Args, tt.args) {
				t.Fatalf("unexpected args: got %#v want %#v", plan.Args, tt.args)
			}
		})
	}
}

func TestSelectManagerPlanAllowScriptsDropsScriptBlockingFlags(t *testing.T) {
	tests := []struct {
		name   string
		result detection.DetectionResult
		args   []string
	}{
		{
			name:   "composer",
			result: detection.DetectionResult{Ecosystem: detection.EcosystemPHP},
			args:   []string{"install", "--no-interaction", "--prefer-dist"},
		},
		{
			name:   "bun",
			result: detection.DetectionResult{Ecosystem: detection.EcosystemNode, Managers: []string{"bun"}},
			args:   []string{"install", "--frozen-lockfile"},
		},
		{
			name:   "npm",
			result: detection.DetectionResult{Ecosystem: detection.EcosystemNode, Managers: []string{"npm"}},
			args:   []string{"ci"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, ok, reason := selectManagerPlan(tt.result, true)
			if !ok {
				t.Fatalf("expected manager plan to be supported, reason=%q", reason)
			}
			if !slices.Equal(plan.Args, tt.args) {
				t.Fatalf("unexpected args with allow-scripts: got %#v want %#v", plan.Args, tt.args)
			}
		})
	}
}

func TestRunAllowScriptsUsesOptInFlags(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFiles(t, dir, "package-lock.json")
	withChdir(t, dir)

	t.Cleanup(func() {
		lookPath = exec.LookPath
		execCommand = exec.CommandContext
	})
	lookPath = func(file string) (string, error) {
		return file, nil
	}

	var ranName string
	var ranArgs []string
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		ranName = name
		ranArgs = append([]string(nil), args...)
		return exec.CommandContext(ctx, "true")
	}

	var stderr bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs([]string{"run", "--allow-scripts"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("run command failed: %v", err)
	}

	if ranName != "npm" {
		t.Fatalf("expected npm execution, got %q", ranName)
	}
	if !slices.Equal(ranArgs, []string{"ci"}) {
		t.Fatalf("expected allow-scripts run to omit --ignore-scripts, got %#v", ranArgs)
	}
	if !strings.Contains(stderr.String(), "pupdate: run npm ci") {
		t.Fatalf("expected allow-scripts status line, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "pupdate: done npm") {
		t.Fatalf("expected allow-scripts completion line, got %q", stderr.String())
	}
	if strings.Contains(stderr.String(), "--ignore-scripts") {
		t.Fatalf("expected allow-scripts status line to omit script-blocking flag, got %q", stderr.String())
	}
}

func TestRunSkipsWhenExpandedManagerMissingFromPath(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFiles(t, dir, "package-lock.json")
	withChdir(t, dir)

	t.Cleanup(func() {
		lookPath = exec.LookPath
		execCommand = exec.CommandContext
	})
	lookPath = func(file string) (string, error) {
		if file == "npm" {
			return "", errors.New("missing")
		}
		return file, nil
	}
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		t.Fatalf("install should not execute when manager missing on PATH")
		return exec.CommandContext(ctx, name, args...)
	}

	var stderr bytes.Buffer
	cmd := newRunCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("run command failed: %v", err)
	}

	if !strings.Contains(stderr.String(), "pupdate: skip node (npm not found on PATH)") {
		t.Fatalf("expected node PATH skip line for npm, got %q", stderr.String())
	}
}

func TestRunPrintsRunLineForExpandedManagers(t *testing.T) {
	tests := []struct {
		name      string
		file      string
		ecosystem string
		manager   string
		args      string
	}{
		{name: "node npm", file: "package-lock.json", ecosystem: "node", manager: "npm", args: "ci --ignore-scripts"},
		{name: "python pip", file: "requirements.txt", ecosystem: "python", manager: "pip", args: "install -r requirements.txt --disable-pip-version-check --no-input"},
		{name: "go", file: "go.mod", ecosystem: "go", manager: "go", args: "mod download"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFixtureFiles(t, dir, tt.file)
			withChdir(t, dir)

			t.Cleanup(func() {
				lookPath = exec.LookPath
				execCommand = exec.CommandContext
			})
			lookPath = func(file string) (string, error) {
				return file, nil
			}
			execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
				return exec.CommandContext(ctx, "true")
			}

			var stderr bytes.Buffer
			cmd := newRunCmd()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&stderr)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("run command failed: %v", err)
			}

			runLine := "pupdate: run " + tt.manager + " " + tt.args
			if !strings.Contains(stderr.String(), runLine) {
				t.Fatalf("expected run line %q, got %q", runLine, stderr.String())
			}
			if !strings.Contains(stderr.String(), "pupdate: done "+tt.manager) {
				t.Fatalf("expected completion line for %s, got %q", tt.ecosystem, stderr.String())
			}
			if !strings.Contains(stderr.String(), "pupdate: run") {
				t.Fatalf("expected run status output for %s", tt.ecosystem)
			}
		})
	}
}

func TestRunExecutesGitSubmoduleUpdateWhenGitDecisionRequiresUpdate(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFiles(t, dir, ".gitmodules")
	withChdir(t, dir)

	t.Cleanup(func() {
		detectFn = detection.Detect
		lookPath = exec.LookPath
		execCommand = exec.CommandContext
		evaluateFreshnessFn = freshness.Evaluate
	})
	detectFn = func(string) ([]detection.DetectionResult, error) {
		return []detection.DetectionResult{{
			Ecosystem:    detection.Ecosystem("git"),
			Managers:     []string{"git"},
			MatchedFiles: []string{".gitmodules"},
		}}, nil
	}
	lookPath = func(file string) (string, error) {
		return file, nil
	}
	evaluateFreshnessFn = func(dir string, detections []detection.DetectionResult, current state.FileState) ([]freshness.EcosystemDecision, error) {
		return []freshness.EcosystemDecision{{
			Ecosystem: "git",
			StateKey:  "git",
			Decision:  freshness.DecisionUpdate,
			Reason:    "git submodule state drifted from recorded revision",
			Lockfiles: map[string]string{".gitmodules": "hash"},
		}}, nil
	}

	var ranName string
	var ranArgs []string
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		ranName = name
		ranArgs = append([]string(nil), args...)
		return exec.CommandContext(ctx, "true")
	}

	var stderr bytes.Buffer
	cmd := newRunCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("run command failed: %v", err)
	}

	if ranName != "git" {
		t.Fatalf("expected git manager execution, got %q", ranName)
	}
	if !slices.Equal(ranArgs, []string{"submodule", "update", "--init", "--recursive"}) {
		t.Fatalf("unexpected git args: %#v", ranArgs)
	}
	if !strings.Contains(stderr.String(), "pupdate: run git submodule update --init --recursive") {
		t.Fatalf("expected git run status line, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "pupdate: done git") {
		t.Fatalf("expected git completion status line, got %q", stderr.String())
	}
}
