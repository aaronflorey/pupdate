package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/aaronflorey/pupdate/internal/detection"
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

func TestRunOutputsDeterministicMultiEcosystemJSON(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFiles(t, dir,
		"bun.lock",
		"package-lock.json",
		"composer.lock",
		"go.mod",
		"Cargo.toml",
		"requirements.txt",
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

	if !slices.Contains(out.Ecosystems[0].MatchedFiles, "bun.lock") || !slices.Contains(out.Ecosystems[0].MatchedFiles, "package-lock.json") {
		t.Fatalf("node matched_files missing lockfiles: %#v", out.Ecosystems[0].MatchedFiles)
	}
	if len(out.Warnings) == 0 {
		t.Fatalf("expected top-level warnings to include node ambiguity warning")
	}
	if out.Warnings[0].Code != "node_multiple_lockfiles" {
		t.Fatalf("expected node ambiguity warning, got %#v", out.Warnings)
	}
}

func TestRunReturnsDetectionFailedPrefixOnDetectorError(t *testing.T) {
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
	dir := t.TempDir()
	writeFixtureFiles(t, dir, "go.mod")
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
	if len(out.Ecosystems) != 1 || out.Ecosystems[0].Ecosystem != "go" {
		t.Fatalf("expected only go ecosystem, got %#v", out.Ecosystems)
	}
	if len(out.Ecosystems[0].Warnings) != 0 {
		t.Fatalf("expected no ecosystem warnings for go, got %#v", out.Ecosystems[0].Warnings)
	}
}

func TestRunOutputIncludesNodeManagers(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFiles(t, dir, "bun.lock", "package-lock.json")
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
	if !slices.Contains(managers, "bun") || !slices.Contains(managers, "npm") {
		t.Fatalf("expected node managers to contain bun and npm, got %#v", managers)
	}
}
