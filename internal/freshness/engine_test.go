package freshness

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/state"
)

func TestEvaluateNoStateRuns(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package-lock.json", "one")

	results, err := Evaluate(
		dir,
		[]detection.DetectionResult{
			{
				Ecosystem:    detection.EcosystemNode,
				MatchedFiles: []string{"package-lock.json"},
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
}

func TestEvaluateChangedLockfileRuns(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "one")
	oldHash := hashText("one")
	writeFile(t, dir, "go.mod", "two")

	current := state.Empty()
	current.Ecosystems["go"] = state.EcosystemState{
		LastSuccessAt: "2026-03-01T12:00:00Z",
		Lockfiles: map[string]string{
			"go.mod": oldHash,
		},
	}

	results, err := Evaluate(
		dir,
		[]detection.DetectionResult{
			{
				Ecosystem:    detection.EcosystemGo,
				MatchedFiles: []string{"go.mod"},
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
}

func TestEvaluateEqualHashesSkips(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "Cargo.toml", "same")
	currentHash := hashText("same")

	equalState := state.Empty()
	equalState.Ecosystems["rust"] = state.EcosystemState{
		LastSuccessAt: "2026-03-01T14:00:00Z",
		Lockfiles: map[string]string{
			"cargo.toml": currentHash,
		},
	}

	equalResults, err := Evaluate(
		dir,
		[]detection.DetectionResult{
			{
				Ecosystem:    detection.EcosystemRust,
				MatchedFiles: []string{"Cargo.toml"},
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
}

func TestEvaluateDetectsAnyLockfileChange(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pyproject.toml", "old")
	writeFile(t, dir, "requirements.txt", "same")
	writeFile(t, dir, "pyproject.toml", "new")

	current := state.Empty()
	current.Ecosystems["python"] = state.EcosystemState{
		LastSuccessAt: "2026-03-01T08:30:00Z",
		Lockfiles: map[string]string{
			"pyproject.toml":   hashText("old"),
			"requirements.txt": hashText("same"),
		},
	}

	results, err := Evaluate(
		dir,
		[]detection.DetectionResult{
			{
				Ecosystem:    detection.EcosystemPython,
				MatchedFiles: []string{"pyproject.toml", "requirements.txt"},
			},
		},
		current,
	)
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if results[0].Decision != DecisionUpdate {
		t.Fatalf("expected DecisionUpdate when a lockfile changes, got %q", results[0].Decision)
	}
	if results[0].Lockfiles["pyproject.toml"] != hashText("new") {
		t.Fatalf("expected updated pyproject.toml hash, got %q", results[0].Lockfiles["pyproject.toml"])
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
