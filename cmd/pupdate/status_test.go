package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/freshness"
	"github.com/aaronflorey/pupdate/internal/state"
)

func TestStatusShowsReadyTarget(t *testing.T) {
	dir := t.TempDir()
	homeDir := t.TempDir()
	configHome := filepath.Join(homeDir, ".config")
	configPath := filepath.Join(configHome, "pupdate", "config.yaml")
	writeFixtureFiles(t, dir, "bun.lock")
	writeFixtureFiles(t, configHome, filepath.Join("pupdate", "config.yaml"))
	if err := os.WriteFile(configPath, []byte("quiet: true\nallow_scripts: true\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	withChdir(t, dir)
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", configHome)

	t.Cleanup(func() {
		detectFn = detection.DetectWithWorkspaceGlobs
		evaluateFreshnessFn = freshness.Evaluate
		lookPath = exec.LookPath
	})
	detectFn = func(string, []string) ([]detection.DetectionResult, error) {
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
		if file == "bun" {
			return "/fake/bin/bun", nil
		}
		return "", exec.ErrNotFound
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs([]string{"status"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("status command failed: %v (stderr=%q)", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "run_status: ready") {
		t.Fatalf("expected ready run status, got %q", out)
	}
	if !strings.Contains(out, "background_hook_lock_status: idle") {
		t.Fatalf("expected idle background hook status, got %q", out)
	}
	if !strings.Contains(out, "run_reason: 1 ecosystem needs updates") {
		t.Fatalf("expected ready run reason, got %q", out)
	}
	if !strings.Contains(out, "run_guidance: (none)") {
		t.Fatalf("expected no top-level guidance for ready status, got %q", out)
	}
	if !strings.Contains(out, "detected_targets: 1") {
		t.Fatalf("expected one detected target, got %q", out)
	}
	if !strings.Contains(out, "[node]") {
		t.Fatalf("expected node target section, got %q", out)
	}
	if !strings.Contains(out, "freshness: update") {
		t.Fatalf("expected freshness update, got %q", out)
	}
	if !strings.Contains(out, "install_status: ready") {
		t.Fatalf("expected ready install status, got %q", out)
	}
	if !strings.Contains(out, "manager_path: /fake/bin/bun") {
		t.Fatalf("expected manager path, got %q", out)
	}
	if !strings.Contains(out, "install_guidance: (none)") {
		t.Fatalf("expected no install guidance for ready target, got %q", out)
	}
	if !strings.Contains(out, "quiet: true") {
		t.Fatalf("expected quiet config in output, got %q", out)
	}
	if !strings.Contains(out, "allow_scripts: true") {
		t.Fatalf("expected allow_scripts config in output, got %q", out)
	}
	if !strings.Contains(out, "effective_allow_scripts: true") {
		t.Fatalf("expected effective allow_scripts in output, got %q", out)
	}
	if !strings.Contains(out, "install_command: bun install --frozen-lockfile") || strings.Contains(out, "install_command: bun install --frozen-lockfile --ignore-scripts") {
		t.Fatalf("expected status command to reflect allow_scripts config, got %q", out)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestStatusShowsRepoSkipForHomeDirectory(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	withChdir(t, homeDir)

	detectCalls := 0
	freshnessCalls := 0
	t.Cleanup(func() {
		detectFn = detection.DetectWithWorkspaceGlobs
		evaluateFreshnessFn = freshness.Evaluate
	})
	detectFn = func(string, []string) ([]detection.DetectionResult, error) {
		detectCalls++
		return nil, errors.New("detect should not run")
	}
	evaluateFreshnessFn = func(string, []detection.DetectionResult, state.FileState) ([]freshness.EcosystemDecision, error) {
		freshnessCalls++
		return nil, errors.New("freshness should not run")
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs([]string{"status"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("status command failed: %v (stderr=%q)", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "run_status: skip") {
		t.Fatalf("expected skip run status, got %q", out)
	}
	if !strings.Contains(out, "run_reason: current directory is $HOME") {
		t.Fatalf("expected home-directory skip reason, got %q", out)
	}
	if !strings.Contains(out, "run_guidance: run pupdate from a project directory instead of $HOME") {
		t.Fatalf("expected home-directory guidance, got %q", out)
	}
	if !strings.Contains(out, "detected_targets: 0") {
		t.Fatalf("expected zero detected targets, got %q", out)
	}
	if detectCalls != 0 {
		t.Fatalf("expected detectFn to be skipped, got %d calls", detectCalls)
	}
	if freshnessCalls != 0 {
		t.Fatalf("expected freshness to be skipped, got %d calls", freshnessCalls)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestStatusShowsRepoSkipForPupIgnore(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFiles(t, dir, ".pupignore")
	if err := os.WriteFile(filepath.Join(dir, state.FileName), []byte("not-json"), 0o644); err != nil {
		t.Fatalf("write invalid state: %v", err)
	}
	withChdir(t, dir)

	detectCalls := 0
	t.Cleanup(func() {
		detectFn = detection.DetectWithWorkspaceGlobs
	})
	detectFn = func(string, []string) ([]detection.DetectionResult, error) {
		detectCalls++
		return nil, errors.New("detect should not run")
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs([]string{"status"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("status command failed: %v (stderr=%q)", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "run_status: skip") {
		t.Fatalf("expected skip run status, got %q", out)
	}
	if !strings.Contains(out, "run_reason: repo marked with .pupignore") {
		t.Fatalf("expected .pupignore skip reason, got %q", out)
	}
	if !strings.Contains(out, "run_guidance: remove .pupignore if you want pupdate to manage this repo") {
		t.Fatalf("expected .pupignore guidance, got %q", out)
	}
	if !strings.Contains(out, "state_warnings: state file is invalid; treating as empty") {
		t.Fatalf("expected invalid state warning, got %q", out)
	}
	if !strings.Contains(out, "detected_targets: 0") {
		t.Fatalf("expected zero detected targets, got %q", out)
	}
	if !strings.Contains(out, "background_hook_lock_status: idle") {
		t.Fatalf("expected idle background hook status, got %q", out)
	}
	if detectCalls != 0 {
		t.Fatalf("expected detectFn to be skipped, got %d calls", detectCalls)
	}
}

func TestStatusShowsRepoSkipOutsideConfiguredRootDirectories(t *testing.T) {
	configHome := t.TempDir()
	allowedRoot := filepath.Join(configHome, "workspace")
	projectDir := t.TempDir()
	writeFixtureFiles(t, configHome, filepath.Join("pupdate", "config.yaml"))
	configPath := filepath.Join(configHome, "pupdate", "config.yaml")
	if err := os.WriteFile(configPath, []byte("root_directories:\n  - "+allowedRoot+"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", configHome)
	withChdir(t, projectDir)

	detectCalls := 0
	freshnessCalls := 0
	t.Cleanup(func() {
		detectFn = detection.DetectWithWorkspaceGlobs
		evaluateFreshnessFn = freshness.Evaluate
	})
	detectFn = func(string, []string) ([]detection.DetectionResult, error) {
		detectCalls++
		return nil, errors.New("detect should not run")
	}
	evaluateFreshnessFn = func(string, []detection.DetectionResult, state.FileState) ([]freshness.EcosystemDecision, error) {
		freshnessCalls++
		return nil, errors.New("freshness should not run")
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs([]string{"status"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("status command failed: %v (stderr=%q)", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "run_status: skip") {
		t.Fatalf("expected skip run status, got %q", out)
	}
	if !strings.Contains(out, "run_reason: current directory is outside configured root_directories") {
		t.Fatalf("expected configured-root skip reason, got %q", out)
	}
	if !strings.Contains(out, "run_guidance: move this project under root_directories, or update root_directories to include it") {
		t.Fatalf("expected configured-root guidance, got %q", out)
	}
	if !strings.Contains(out, "detected_targets: 0") {
		t.Fatalf("expected zero detected targets, got %q", out)
	}
	if detectCalls != 0 {
		t.Fatalf("expected detectFn to be skipped, got %d calls", detectCalls)
	}
	if freshnessCalls != 0 {
		t.Fatalf("expected freshness to be skipped, got %d calls", freshnessCalls)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestStatusPassesResolvedWorkspaceGlobsToDetection(t *testing.T) {
	dir := t.TempDir()
	configHome := t.TempDir()
	configPath := filepath.Join(configHome, "pupdate", "config.yaml")
	writeFixtureFiles(t, configHome, filepath.Join("pupdate", "config.yaml"))
	if err := os.WriteFile(configPath, []byte("workspace_globs:\n  - ' apps/* '\n  - ./services/*\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", configHome)
	withChdir(t, dir)

	t.Cleanup(func() {
		detectFn = detection.DetectWithWorkspaceGlobs
		evaluateFreshnessFn = freshness.Evaluate
	})

	var gotDir string
	var gotGlobs []string
	detectFn = func(dir string, workspaceGlobs []string) ([]detection.DetectionResult, error) {
		gotDir = dir
		gotGlobs = append([]string(nil), workspaceGlobs...)
		return nil, nil
	}
	evaluateFreshnessFn = func(string, []detection.DetectionResult, state.FileState) ([]freshness.EcosystemDecision, error) {
		return nil, nil
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs([]string{"status"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("status command failed: %v (stderr=%q)", err, stderr.String())
	}

	if gotDir != "." {
		t.Fatalf("expected detectFn dir to be '.', got %q", gotDir)
	}
	if !slices.Equal(gotGlobs, []string{"apps/*", "services/*"}) {
		t.Fatalf("expected resolved workspace globs, got %#v", gotGlobs)
	}
	if !strings.Contains(stdout.String(), "run_status: idle") {
		t.Fatalf("expected idle status when detector returns nothing, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestStatusShowsActiveBackgroundHookLock(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFiles(t, dir, backgroundHookLockFileName)
	withChdir(t, dir)

	now := time.Now()
	lockPath := filepath.Join(dir, backgroundHookLockFileName)
	if err := os.Chtimes(lockPath, now, now); err != nil {
		t.Fatalf("touch lock file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs([]string{"status"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("status command failed: %v (stderr=%q)", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "background_hook_lock_status: active") {
		t.Fatalf("expected active background hook status, got %q", out)
	}
	if !strings.Contains(out, "background_hook_lock_path: "+filepath.Join(".", backgroundHookLockFileName)) {
		t.Fatalf("expected background hook lock path, got %q", out)
	}
	if !strings.Contains(out, "run_guidance: wait for the current background hook run to finish before expecting another async hook run") {
		t.Fatalf("expected active-hook guidance, got %q", out)
	}
}

func TestStatusReturnsConfigParseErrorWhenYAMLIsInvalid(t *testing.T) {
	homeDir := t.TempDir()
	configHome := filepath.Join(homeDir, ".config")
	projectDir := t.TempDir()
	configPath := filepath.Join(configHome, "pupdate", "config.yaml")
	writeFixtureFiles(t, configHome, filepath.Join("pupdate", "config.yaml"))
	if err := os.WriteFile(configPath, []byte("root_directories: [oops\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", configHome)
	withChdir(t, projectDir)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"status"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected status command to fail")
	}
	if !strings.Contains(err.Error(), "failed to parse "+configPath) {
		t.Fatalf("expected parse error with config path, got %q", err.Error())
	}
}

func TestStatusShowsBlockedTargetWhenManagerMissing(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFiles(t, dir, "bun.lock")
	withChdir(t, dir)

	t.Cleanup(func() {
		detectFn = detection.DetectWithWorkspaceGlobs
		evaluateFreshnessFn = freshness.Evaluate
		lookPath = exec.LookPath
	})
	detectFn = func(string, []string) ([]detection.DetectionResult, error) {
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
	lookPath = func(string) (string, error) {
		return "", exec.ErrNotFound
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs([]string{"status"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("status command failed: %v (stderr=%q)", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "run_status: blocked") {
		t.Fatalf("expected blocked run status, got %q", out)
	}
	if !strings.Contains(out, "install_status: blocked") {
		t.Fatalf("expected blocked install status, got %q", out)
	}
	if !strings.Contains(out, "install_reason: bun not found on PATH") {
		t.Fatalf("expected missing manager reason, got %q", out)
	}
	if !strings.Contains(out, "manager_path: (none)") {
		t.Fatalf("expected empty manager path, got %q", out)
	}
	if !strings.Contains(out, "install_guidance: install bun or add it to PATH, then rerun pupdate status") {
		t.Fatalf("expected missing-manager guidance, got %q", out)
	}
}
