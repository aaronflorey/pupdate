package freshness

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/state"
)

func TestEvaluateNoStateRuns(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "bun.lock", "one")

	results, err := Evaluate(
		dir,
		[]detection.DetectionResult{
			{
				Ecosystem:    detection.EcosystemNode,
				MatchedFiles: []string{"bun.lock"},
			},
		},
		state.Empty(),
	)
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(results))
	}
	if results[0].Decision != DecisionUpdate {
		t.Fatalf("expected DecisionUpdate, got %q", results[0].Decision)
	}
	if results[0].Reason != "missing prior lockfile hash" {
		t.Fatalf("expected missing-prior-hash reason, got %q", results[0].Reason)
	}
}

func TestEvaluateChangedLockfileRuns(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "composer.lock", "one")
	oldHash := hashText("one")
	writeFile(t, dir, "composer.lock", "two")

	current := state.Empty()
	current.Ecosystems["php"] = state.EcosystemState{
		LastSuccessAt: "2026-03-01T12:00:00Z",
		Lockfiles: map[string]string{
			"composer.lock": oldHash,
		},
	}

	results, err := Evaluate(
		dir,
		[]detection.DetectionResult{
			{
				Ecosystem:    detection.EcosystemPHP,
				MatchedFiles: []string{"composer.lock"},
			},
		},
		current,
	)
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if results[0].Decision != DecisionUpdate {
		t.Fatalf("expected DecisionUpdate, got %q", results[0].Decision)
	}
	if results[0].Reason != "dependency lockfiles changed since last successful run" {
		t.Fatalf("expected changed-lockfiles reason, got %q", results[0].Reason)
	}
}

func TestEvaluateEqualHashesSkips(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "composer.lock", "same")
	currentHash := hashText("same")

	equalState := state.Empty()
	equalState.Ecosystems["php"] = state.EcosystemState{
		LastSuccessAt: "2026-03-01T14:00:00Z",
		Lockfiles: map[string]string{
			"composer.lock": currentHash,
		},
	}

	equalResults, err := Evaluate(
		dir,
		[]detection.DetectionResult{
			{
				Ecosystem:    detection.EcosystemPHP,
				MatchedFiles: []string{"composer.lock"},
			},
		},
		equalState,
	)
	if err != nil {
		t.Fatalf("Evaluate (equal) returned error: %v", err)
	}
	if equalResults[0].Decision != DecisionSkip {
		t.Fatalf("expected DecisionSkip for equal hash, got %q", equalResults[0].Decision)
	}
	if equalResults[0].Reason != "dependency lockfiles unchanged since last successful run" {
		t.Fatalf("expected unchanged-lockfiles reason, got %q", equalResults[0].Reason)
	}
}

func TestHasGitSubmoduleDrift(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		drift bool
	}{
		{name: "clean", lines: []string{" 3f4d1f0 libs/foo (heads/main)"}, drift: false},
		{name: "uninitialized", lines: []string{"-3f4d1f0 libs/foo"}, drift: true},
		{name: "checked-out-different", lines: []string{"+3f4d1f0 libs/foo"}, drift: true},
		{name: "merge-conflict", lines: []string{"U3f4d1f0 libs/foo"}, drift: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasGitSubmoduleDrift(tt.lines)
			if got != tt.drift {
				t.Fatalf("unexpected drift detection for %q: got %v want %v", tt.name, got, tt.drift)
			}
		})
	}
}

func TestEvaluateGitSubmoduleDriftOverridesUnchangedHashDecision(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".gitmodules", "[submodule \"libs/foo\"]\npath = libs/foo\nurl = git@example.com:foo.git\n")

	currentHash := hashText("[submodule \"libs/foo\"]\npath = libs/foo\nurl = git@example.com:foo.git\n")
	current := state.Empty()
	current.Ecosystems["git"] = state.EcosystemState{
		LastSuccessAt: "2026-03-01T14:00:00Z",
		Lockfiles: map[string]string{
			".gitmodules": currentHash,
		},
	}

	gitSubmoduleStatusFn = func(string) ([]string, error) {
		return []string{"-3f4d1f0 libs/foo"}, nil
	}
	t.Cleanup(func() {
		gitSubmoduleStatusFn = defaultGitSubmoduleStatus
	})

	results, err := Evaluate(
		dir,
		[]detection.DetectionResult{{
			Ecosystem:    detection.Ecosystem("git"),
			MatchedFiles: []string{".gitmodules"},
		}},
		current,
	)
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected one decision, got %d", len(results))
	}
	if results[0].Decision != DecisionUpdate {
		t.Fatalf("expected drifted submodule decision to update, got %q", results[0].Decision)
	}
	if results[0].Reason != "git submodule state drifted from recorded revision" {
		t.Fatalf("unexpected drift reason: %q", results[0].Reason)
	}
}

func TestEvaluateGitSubmoduleStatusFailureDoesNotFailEvaluation(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".gitmodules", "[submodule \"libs/foo\"]\n")

	gitSubmoduleStatusFn = func(string) ([]string, error) {
		return nil, errors.New("status failed")
	}
	t.Cleanup(func() {
		gitSubmoduleStatusFn = defaultGitSubmoduleStatus
	})

	results, err := Evaluate(
		dir,
		[]detection.DetectionResult{{
			Ecosystem:    detection.Ecosystem("git"),
			MatchedFiles: []string{".gitmodules"},
		}},
		state.Empty(),
	)
	if err != nil {
		t.Fatalf("Evaluate should not fail when git status check fails: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one decision, got %d", len(results))
	}
	if !slices.Contains([]Decision{DecisionSkip, DecisionUpdate}, results[0].Decision) {
		t.Fatalf("expected non-crashing decision, got %q", results[0].Decision)
	}
}

func writeFile(t *testing.T, dir, rel, contents string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func hashText(input string) string {
	hasher := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hasher[:])
}
