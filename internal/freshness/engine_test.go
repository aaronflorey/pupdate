package freshness

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/state"
)

func TestEvaluateNoStateRuns(t *testing.T) {
	dir := t.TempDir()
	mtime := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	writeFileWithMTime(t, dir, "package-lock.json", mtime)

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

func TestEvaluateOlderStateRuns(t *testing.T) {
	dir := t.TempDir()
	mtime := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	writeFileWithMTime(t, dir, "go.mod", mtime)

	current := state.Empty()
	current.Ecosystems["go"] = state.EcosystemState{
		LastSuccessAt: state.FormatRFC3339UTC(mtime.Add(-time.Hour)),
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

func TestEvaluateEqualOrNewerStateSkips(t *testing.T) {
	dir := t.TempDir()
	mtime := time.Date(2026, 3, 1, 14, 0, 0, 0, time.UTC)
	writeFileWithMTime(t, dir, "Cargo.toml", mtime)

	equalState := state.Empty()
	equalState.Ecosystems["rust"] = state.EcosystemState{
		LastSuccessAt: state.FormatRFC3339UTC(mtime),
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
		t.Fatalf("expected DecisionSkip for equal timestamp, got %q", equalResults[0].Decision)
	}

	newerState := state.Empty()
	newerState.Ecosystems["rust"] = state.EcosystemState{
		LastSuccessAt: state.FormatRFC3339UTC(mtime.Add(time.Minute)),
	}
	newerResults, err := Evaluate(
		dir,
		[]detection.DetectionResult{
			{
				Ecosystem:    detection.EcosystemRust,
				MatchedFiles: []string{"Cargo.toml"},
			},
		},
		newerState,
	)
	if err != nil {
		t.Fatalf("Evaluate (newer) returned error: %v", err)
	}
	if newerResults[0].Decision != DecisionSkip {
		t.Fatalf("expected DecisionSkip for newer timestamp, got %q", newerResults[0].Decision)
	}
}

func TestEvaluateUsesMaxMatchedFileMTime(t *testing.T) {
	dir := t.TempDir()
	oldTime := time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC)
	newTime := time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)
	writeFileWithMTime(t, dir, "pyproject.toml", oldTime)
	writeFileWithMTime(t, dir, "requirements.txt", newTime)

	current := state.Empty()
	current.Ecosystems["python"] = state.EcosystemState{
		LastSuccessAt: state.FormatRFC3339UTC(oldTime.Add(30 * time.Minute)),
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
		t.Fatalf("expected DecisionUpdate based on max matched file mtime, got %q", results[0].Decision)
	}
	if !results[0].MaxMTime.Equal(newTime) {
		t.Fatalf("expected MaxMTime %s, got %s", newTime, results[0].MaxMTime)
	}
}

func writeFileWithMTime(t *testing.T, dir, rel string, mtime time.Time) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatalf("chtimes %s: %v", rel, err)
	}
}
