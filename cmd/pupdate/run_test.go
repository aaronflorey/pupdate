package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/freshness"
	"github.com/aaronflorey/pupdate/internal/state"
)

type runOutput struct {
	Directory  string               `json:"directory"`
	Ecosystems []runEcosystemOutput `json:"ecosystems"`
	Warnings   []runWarningOutput   `json:"warnings"`
}

type runEcosystemOutput struct {
	Ecosystem    string             `json:"ecosystem"`
	Managers     []string           `json:"managers"`
	MatchedFiles []string           `json:"matched_files"`
	Warnings     []runWarningOutput `json:"warnings"`
}

type runWarningOutput struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeFixtureFiles(t *testing.T, dir string, files ...string) {
	t.Helper()
	for _, file := range files {
		path := filepath.Join(dir, file)
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

func parseRunOutput(t *testing.T, stdout bytes.Buffer) runOutput {
	t.Helper()
	var out runOutput
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal run output: %v; raw=%q", err, stdout.String())
	}
	return out
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

func TestRunOutputsDeterministicMultiEcosystemJSON(t *testing.T) {
	disableInstall(t)
	dir := t.TempDir()
	writeFixtureFiles(t, dir,
		"bun.lock",
		"composer.lock",
		"requirements.txt",
		"go.mod",
		"cargo.lock",
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

	out := parseRunOutput(t, stdout)
	if out.Directory != "." {
		t.Fatalf("expected directory '.', got %q", out.Directory)
	}

	expectedOrder := []string{"node", "php", "go", "rust", "python"}
	if len(out.Ecosystems) != len(expectedOrder) {
		t.Fatalf("expected %d ecosystems, got %d", len(expectedOrder), len(out.Ecosystems))
	}
	for i, expected := range expectedOrder {
		if out.Ecosystems[i].Ecosystem != expected {
			t.Fatalf("unexpected ecosystem order at %d: got %q want %q", i, out.Ecosystems[i].Ecosystem, expected)
		}
		if out.Ecosystems[i].MatchedFiles == nil || len(out.Ecosystems[i].MatchedFiles) == 0 {
			t.Fatalf("ecosystem %q missing matched_files", out.Ecosystems[i].Ecosystem)
		}
	}

	if !slices.Contains(out.Ecosystems[0].MatchedFiles, "bun.lock") {
		t.Fatalf("node matched_files missing lockfiles: %#v", out.Ecosystems[0].MatchedFiles)
	}
	if !slices.Contains(out.Ecosystems[2].MatchedFiles, "go.mod") {
		t.Fatalf("go matched_files missing go.mod: %#v", out.Ecosystems[2].MatchedFiles)
	}
	if !slices.Contains(out.Ecosystems[3].MatchedFiles, "cargo.lock") {
		t.Fatalf("rust matched_files missing cargo.lock: %#v", out.Ecosystems[3].MatchedFiles)
	}
	if !slices.Contains(out.Ecosystems[4].MatchedFiles, "requirements.txt") {
		t.Fatalf("python matched_files missing requirements.txt: %#v", out.Ecosystems[4].MatchedFiles)
	}
	if len(out.Warnings) != 0 {
		t.Fatalf("expected no top-level warnings, got %#v", out.Warnings)
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

func TestRunOutputIncludesMatchedFilesFieldPerEcosystem(t *testing.T) {
	disableInstall(t)
	dir := t.TempDir()
	writeFixtureFiles(t, dir, "composer.lock", "go.mod")
	withChdir(t, dir)

	var stdout bytes.Buffer
	cmd := newRunCmd()
	cmd.SetOut(&stdout)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("run command failed: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal raw output: %v", err)
	}
	ecosystems, ok := raw["ecosystems"].([]any)
	if !ok {
		t.Fatalf("ecosystems missing or wrong type: %#v", raw["ecosystems"])
	}
	for i, item := range ecosystems {
		obj, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("ecosystem entry %d is not object: %#v", i, item)
		}
		if _, ok := obj["matched_files"]; !ok {
			t.Fatalf("ecosystem entry %d missing matched_files: %#v", i, obj)
		}
	}
}

func TestRunOutputHasNoWarningsForSingleLockfile(t *testing.T) {
	disableInstall(t)
	dir := t.TempDir()
	writeFixtureFiles(t, dir, "composer.lock")
	withChdir(t, dir)

	var stdout bytes.Buffer
	cmd := newRunCmd()
	cmd.SetOut(&stdout)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("run command failed: %v", err)
	}

	out := parseRunOutput(t, stdout)
	if len(out.Warnings) != 0 {
		t.Fatalf("expected no top-level warnings, got %#v", out.Warnings)
	}
	if len(out.Ecosystems) != 1 || out.Ecosystems[0].Ecosystem != "php" {
		t.Fatalf("expected only php ecosystem, got %#v", out.Ecosystems)
	}
	if len(out.Ecosystems[0].Warnings) != 0 {
		t.Fatalf("expected no ecosystem warnings for go, got %#v", out.Ecosystems[0].Warnings)
	}
}

func TestRunOutputIncludesNodeManagers(t *testing.T) {
	disableInstall(t)
	dir := t.TempDir()
	writeFixtureFiles(t, dir, "bun.lock", "package-lock.json", "pnpm-lock.yaml", "yarn.lock")
	withChdir(t, dir)

	var stdout bytes.Buffer
	cmd := newRunCmd()
	cmd.SetOut(&stdout)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("run command failed: %v", err)
	}

	out := parseRunOutput(t, stdout)
	if len(out.Ecosystems) == 0 || out.Ecosystems[0].Ecosystem != "node" {
		t.Fatalf("expected first ecosystem to be node, got %#v", out.Ecosystems)
	}
	managers := out.Ecosystems[0].Managers
	if !slices.Equal(managers, []string{"bun", "npm", "pnpm", "yarn"}) {
		t.Fatalf("expected node managers to include bun/npm/pnpm/yarn, got %#v", managers)
	}
}

func TestRunOutputIncludesExpandedEcosystemManagers(t *testing.T) {
	disableInstall(t)
	dir := t.TempDir()
	writeFixtureFiles(t, dir, "requirements.txt", "go.mod", "cargo.lock")
	withChdir(t, dir)

	var stdout bytes.Buffer
	cmd := newRunCmd()
	cmd.SetOut(&stdout)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("run command failed: %v", err)
	}

	out := parseRunOutput(t, stdout)

	var pythonManagers, goManagers, rustManagers []string
	for _, ecosystem := range out.Ecosystems {
		switch ecosystem.Ecosystem {
		case "python":
			pythonManagers = ecosystem.Managers
		case "go":
			goManagers = ecosystem.Managers
		case "rust":
			rustManagers = ecosystem.Managers
		}
	}

	if !slices.Equal(pythonManagers, []string{"pip"}) {
		t.Fatalf("expected python manager pip, got %#v", pythonManagers)
	}
	if !slices.Equal(goManagers, []string{"go"}) {
		t.Fatalf("expected go manager list, got %#v", goManagers)
	}
	if !slices.Equal(rustManagers, []string{"cargo"}) {
		t.Fatalf("expected rust manager list, got %#v", rustManagers)
	}
}

func TestRunPupignorePrintsSkipRepoAndSkipsInstalls(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFiles(t, dir, "bun.lock", ".pupignore")
	withChdir(t, dir)

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
	})

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
	})

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
	})

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
			plan, ok, reason := selectManagerPlan(tt.result)
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
}
