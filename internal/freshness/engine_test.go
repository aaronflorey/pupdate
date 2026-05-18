package freshness

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

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
	info, err := os.Stat(filepath.Join(dir, "composer.lock"))
	if err != nil {
		t.Fatalf("stat composer.lock: %v", err)
	}
	currentHash := hashText("same")

	equalState := state.Empty()
	equalState.Ecosystems["php"] = state.EcosystemState{
		LastSuccessAt: "2026-03-01T14:00:00Z",
		Lockfiles: map[string]string{
			"composer.lock": currentHash,
		},
		LockfileMetadata: map[string]state.LockfileMetadata{
			"composer.lock": metadataForFile(info),
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

func TestEvaluateMatchingIdentityAndMetadataReusesStoredHash(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "bun.lock", "same")
	info, err := os.Stat(filepath.Join(dir, "bun.lock"))
	if err != nil {
		t.Fatalf("stat bun.lock: %v", err)
	}

	originalHashFile := hashFileFn
	called := 0
	hashFileFn = func(string) (string, error) {
		called++
		return hashText("same"), nil
	}
	t.Cleanup(func() {
		hashFileFn = originalHashFile
	})

	current := state.Empty()
	current.Ecosystems["node"] = state.EcosystemState{
		LastSuccessAt: "2026-03-01T14:00:00Z",
		Lockfiles: map[string]string{
			"bun.lock": hashText("same"),
		},
		LockfileMetadata: map[string]state.LockfileMetadata{
			"bun.lock": metadataForFile(info),
		},
	}

	results, err := Evaluate(dir, []detection.DetectionResult{{
		Ecosystem:    detection.EcosystemNode,
		MatchedFiles: []string{"bun.lock"},
	}}, current)
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if called != 0 {
		t.Fatalf("expected stored hash reuse when metadata and identity match, got %d hash calls", called)
	}
	if results[0].Decision != DecisionSkip {
		t.Fatalf("expected DecisionSkip, got %q", results[0].Decision)
	}
}

func TestEvaluateMatchingMetadataWithoutStrongIdentityStillRehashesLockfile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "bun.lock", "same")
	info, err := os.Stat(filepath.Join(dir, "bun.lock"))
	if err != nil {
		t.Fatalf("stat bun.lock: %v", err)
	}

	originalHashFile := hashFileFn
	called := 0
	hashFileFn = func(string) (string, error) {
		called++
		return hashText("same"), nil
	}
	t.Cleanup(func() {
		hashFileFn = originalHashFile
	})

	current := state.Empty()
	current.Ecosystems["node"] = state.EcosystemState{
		LastSuccessAt: "2026-03-01T14:00:00Z",
		Lockfiles: map[string]string{
			"bun.lock": hashText("same"),
		},
		LockfileMetadata: map[string]state.LockfileMetadata{
			"bun.lock": {
				Size:            info.Size(),
				ModTimeUnixNano: info.ModTime().UTC().UnixNano(),
				Mode:            info.Mode().String(),
			},
		},
	}

	results, err := Evaluate(dir, []detection.DetectionResult{{
		Ecosystem:    detection.EcosystemNode,
		MatchedFiles: []string{"bun.lock"},
	}}, current)
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if called != 1 {
		t.Fatalf("expected lockfile hash to be recomputed when prior state lacks strong identity, got %d calls", called)
	}
	if results[0].Decision != DecisionSkip {
		t.Fatalf("expected DecisionSkip, got %q", results[0].Decision)
	}
}

func TestEvaluateChangedMetadataRehashesLockfile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "bun.lock", "new")

	originalHashFile := hashFileFn
	called := 0
	hashFileFn = func(path string) (string, error) {
		called++
		return originalHashFile(path)
	}
	t.Cleanup(func() {
		hashFileFn = originalHashFile
	})

	current := state.Empty()
	current.Ecosystems["node"] = state.EcosystemState{
		LastSuccessAt: "2026-03-01T14:00:00Z",
		Lockfiles: map[string]string{
			"bun.lock": hashText("old"),
		},
		LockfileMetadata: map[string]state.LockfileMetadata{
			"bun.lock": {Size: 3, ModTimeUnixNano: 1, Mode: "-rw-r--r--", FileID: "1:2", ChangeTimeUnixNano: 3},
		},
	}

	results, err := Evaluate(dir, []detection.DetectionResult{{
		Ecosystem:    detection.EcosystemNode,
		MatchedFiles: []string{"bun.lock"},
	}}, current)
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if called != 1 {
		t.Fatalf("expected lockfile rehash when metadata changes, got %d calls", called)
	}
	if results[0].Decision != DecisionUpdate {
		t.Fatalf("expected DecisionUpdate, got %q", results[0].Decision)
	}
}

func TestEvaluateSameMetadataDifferentContentTriggersUpdate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bun.lock")
	writeFile(t, dir, "bun.lock", "same")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat bun.lock: %v", err)
	}

	if err := os.WriteFile(path, []byte("diff"), 0o644); err != nil {
		t.Fatalf("rewrite bun.lock: %v", err)
	}
	if err := os.Chmod(path, info.Mode()); err != nil {
		t.Fatalf("chmod bun.lock: %v", err)
	}
	modTime := info.ModTime()
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("chtimes bun.lock: %v", err)
	}

	current := state.Empty()
	current.Ecosystems["node"] = state.EcosystemState{
		LastSuccessAt: "2026-03-01T14:00:00Z",
		Lockfiles: map[string]string{
			"bun.lock": hashText("same"),
		},
		LockfileMetadata: map[string]state.LockfileMetadata{
			"bun.lock": metadataForFile(info),
		},
	}

	results, err := Evaluate(dir, []detection.DetectionResult{{
		Ecosystem:    detection.EcosystemNode,
		MatchedFiles: []string{"bun.lock"},
	}}, current)
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if results[0].Decision != DecisionUpdate {
		t.Fatalf("expected DecisionUpdate when content changes without metadata drift, got %q", results[0].Decision)
	}
}

func TestEvaluateMissingLockfileReturnsError(t *testing.T) {
	_, err := Evaluate(t.TempDir(), []detection.DetectionResult{{
		Ecosystem:    detection.EcosystemNode,
		MatchedFiles: []string{"bun.lock"},
	}}, state.Empty())
	if err == nil {
		t.Fatal("expected missing lockfile error")
	}
	if !strings.Contains(err.Error(), `stat matched file "bun.lock":`) {
		t.Fatalf("unexpected missing lockfile error: %v", err)
	}
}

func TestEvaluateRenamedLockfileTriggersUpdate(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "frontend/package-lock.json", "same")

	current := state.Empty()
	current.Ecosystems["node@frontend"] = state.EcosystemState{
		LastSuccessAt: "2026-03-01T14:00:00Z",
		Lockfiles: map[string]string{
			"package-lock.json": hashText("same"),
		},
	}

	results, err := Evaluate(dir, []detection.DetectionResult{{
		Ecosystem:    detection.EcosystemNode,
		Directory:    "frontend",
		MatchedFiles: []string{"frontend/package-lock.json"},
	}}, current)
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if results[0].Decision != DecisionUpdate {
		t.Fatalf("expected DecisionUpdate for renamed lockfile path, got %q", results[0].Decision)
	}
}

func TestEvaluateRustCanonicalLockfileUsesLowercaseStateKey(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "Cargo.lock", "same")

	current := state.Empty()
	current.Ecosystems["rust"] = state.EcosystemState{
		LastSuccessAt: "2026-03-01T14:00:00Z",
		Lockfiles: map[string]string{
			"cargo.lock": hashText("same"),
		},
	}

	results, err := Evaluate(
		dir,
		[]detection.DetectionResult{{
			Ecosystem:    detection.EcosystemRust,
			MatchedFiles: []string{"Cargo.lock"},
		}},
		current,
	)
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if results[0].Decision != DecisionSkip {
		t.Fatalf("expected canonical Cargo.lock to compare against lowercase state key, got %q", results[0].Decision)
	}
}

func TestEvaluateMixedCaseLockfileUsesActualMatchedPath(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "BUN.LOCK", "same")

	current := state.Empty()
	current.Ecosystems["node"] = state.EcosystemState{
		LastSuccessAt: "2026-03-01T14:00:00Z",
		Lockfiles: map[string]string{
			"bun.lock": hashText("same"),
		},
	}

	results, err := Evaluate(
		dir,
		[]detection.DetectionResult{{
			Ecosystem:    detection.EcosystemNode,
			MatchedFiles: []string{"BUN.LOCK"},
		}},
		current,
	)
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if results[0].Decision != DecisionSkip {
		t.Fatalf("expected mixed-case lockfile to compare against lowercase state key, got %q", results[0].Decision)
	}
}

func TestEvaluatePHPLegacyVendorChecksumSkipsWhenComposerLockUnchanged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "composer.lock", "same")
	writeFile(t, dir, "vendor/composer/installed.php", "same")

	current := state.Empty()
	current.Ecosystems["php"] = state.EcosystemState{
		LastSuccessAt: "2026-03-01T14:00:00Z",
		Lockfiles: map[string]string{
			"composer.lock":      hashText("same"),
			phpVendorChecksumKey: hashText("stale"),
		},
	}

	results, err := Evaluate(
		dir,
		[]detection.DetectionResult{{
			Ecosystem:    detection.EcosystemPHP,
			MatchedFiles: []string{"composer.lock"},
		}},
		current,
	)
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if results[0].Decision != DecisionSkip {
		t.Fatalf("expected DecisionSkip, got %q", results[0].Decision)
	}
	if results[0].Reason != "dependency lockfiles unchanged since last successful run" {
		t.Fatalf("expected unchanged-lockfiles reason, got %q", results[0].Reason)
	}
}

func TestEvaluatePHPMissingVendorDirectoryRunsWhenComposerLockUnchanged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "composer.lock", "same")

	current := state.Empty()
	current.Ecosystems["php"] = state.EcosystemState{
		LastSuccessAt: "2026-03-01T14:00:00Z",
		Lockfiles: map[string]string{
			"composer.lock":      hashText("same"),
			phpVendorChecksumKey: hashText("legacy"),
		},
	}

	results, err := Evaluate(
		dir,
		[]detection.DetectionResult{{
			Ecosystem:    detection.EcosystemPHP,
			MatchedFiles: []string{"composer.lock"},
		}},
		current,
	)
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if results[0].Decision != DecisionUpdate {
		t.Fatalf("expected DecisionUpdate for missing vendor directory, got %q", results[0].Decision)
	}
	if results[0].Reason != "composer vendor directory missing since last successful run" {
		t.Fatalf("unexpected missing-vendor reason: %q", results[0].Reason)
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
	if results[0].Reason != "git submodule status failed: status failed" {
		t.Fatalf("unexpected failure reason: %q", results[0].Reason)
	}
}

func TestDefaultGitSubmoduleStatusTimesOut(t *testing.T) {
	originalRunner := runGitSubmoduleStatusCommand
	originalTimeout := gitSubmoduleStatusTimeout
	runGitSubmoduleStatusCommand = func(ctx context.Context, dir string) ([]byte, []byte, error) {
		<-ctx.Done()
		return nil, nil, ctx.Err()
	}
	gitSubmoduleStatusTimeout = 5 * time.Millisecond
	t.Cleanup(func() {
		runGitSubmoduleStatusCommand = originalRunner
		gitSubmoduleStatusTimeout = originalTimeout
	})

	_, err := defaultGitSubmoduleStatus(t.TempDir())
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if err.Error() != "timed out after 5ms" {
		t.Fatalf("unexpected timeout error: %v", err)
	}
}

func TestDefaultGitSubmoduleStatusIncludesStderr(t *testing.T) {
	originalRunner := runGitSubmoduleStatusCommand
	runGitSubmoduleStatusCommand = func(context.Context, string) ([]byte, []byte, error) {
		return nil, []byte("fatal: not a git repository\n"), errors.New("exit status 128")
	}
	t.Cleanup(func() {
		runGitSubmoduleStatusCommand = originalRunner
	})

	_, err := defaultGitSubmoduleStatus(t.TempDir())
	if err == nil {
		t.Fatal("expected command error")
	}
	if err.Error() != "exit status 128: fatal: not a git repository" {
		t.Fatalf("unexpected command error: %v", err)
	}
}

func TestEvaluateUsesNamespacedStateKeysForSubdirectories(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "frontend"), 0o755); err != nil {
		t.Fatalf("mkdir frontend: %v", err)
	}
	writeFile(t, filepath.Join(dir, "frontend"), "package-lock.json", "same")

	current := state.Empty()
	current.Ecosystems["node"] = state.EcosystemState{
		LastSuccessAt: "2026-03-01T14:00:00Z",
		Lockfiles: map[string]string{
			"package-lock.json": hashText("different"),
		},
	}
	current.Ecosystems["node@frontend"] = state.EcosystemState{
		LastSuccessAt: "2026-03-01T14:00:00Z",
		Lockfiles: map[string]string{
			"frontend/package-lock.json": hashText("same"),
		},
	}

	decisions, err := Evaluate(
		dir,
		[]detection.DetectionResult{{
			Ecosystem:    detection.EcosystemNode,
			Directory:    "frontend",
			MatchedFiles: []string{"frontend/package-lock.json"},
		}},
		current,
	)
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if len(decisions) != 1 {
		t.Fatalf("expected one decision, got %d", len(decisions))
	}
	if decisions[0].StateKey != "node@frontend" {
		t.Fatalf("expected namespaced state key, got %q", decisions[0].StateKey)
	}
	if decisions[0].Decision != DecisionSkip {
		t.Fatalf("expected subdirectory state comparison to skip, got %q", decisions[0].Decision)
	}
}

func writeFile(t *testing.T, dir, rel, contents string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func metadataForFile(info os.FileInfo) state.LockfileMetadata {
	metadata := state.LockfileMetadata{
		Size:            info.Size(),
		ModTimeUnixNano: info.ModTime().UTC().UnixNano(),
		Mode:            info.Mode().String(),
	}
	enrichLockfileMetadata(info, &metadata)
	return metadata
}

func hashText(input string) string {
	hasher := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hasher[:])
}
