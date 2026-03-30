package detection

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func writeFiles(t *testing.T, dir string, files ...string) {
	t.Helper()
	for _, file := range files {
		path := filepath.Join(dir, file)
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("write %s: %v", file, err)
		}
	}
}

func assertContainsEcosystem(t *testing.T, results []DetectionResult, eco Ecosystem) DetectionResult {
	t.Helper()
	for _, result := range results {
		if result.Ecosystem == eco {
			return result
		}
	}
	t.Fatalf("expected ecosystem %q in results: %#v", eco, results)
	return DetectionResult{}
}

func assertHasFile(t *testing.T, files []string, file string) {
	t.Helper()
	if !slices.Contains(files, file) {
		t.Fatalf("expected matched file %q in %#v", file, files)
	}
}

func assertHasWarning(t *testing.T, warnings []Warning, code WarningCode) {
	t.Helper()
	for _, warning := range warnings {
		if warning.Code == code {
			return
		}
	}
	t.Fatalf("expected warning code %q in %#v", code, warnings)
}

func TestDetectNodeBunSignals(t *testing.T) {
	dir := t.TempDir()
	writeFiles(
		t,
		dir,
		"bun.lock",
		"pnpm-lock.yaml",
		"pnpm-lock.yml",
		"yarn.lock",
		"package-lock.json",
		"bun.lock.bak",
		"pnpm-lock.yaml.tmp",
		"yarn.lock.orig",
		"package-lock.json.swp",
	)

	// Planned signature contract: Detect(dir string) ([]DetectionResult, error).
	results, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected exactly one ecosystem, got %d: %#v", len(results), results)
	}

	node := assertContainsEcosystem(t, results, EcosystemNode)
	assertHasFile(t, node.MatchedFiles, "bun.lock")
	assertHasFile(t, node.MatchedFiles, "pnpm-lock.yaml")
	assertHasFile(t, node.MatchedFiles, "pnpm-lock.yml")
	assertHasFile(t, node.MatchedFiles, "yarn.lock")
	assertHasFile(t, node.MatchedFiles, "package-lock.json")
	if len(node.MatchedFiles) != 5 {
		t.Fatalf("expected only canonical node lockfiles, got %#v", node.MatchedFiles)
	}
	if !slices.Equal(node.Managers, []string{"bun", "npm", "pnpm", "yarn"}) {
		t.Fatalf("unexpected node managers order/content: %#v", node.Managers)
	}
	assertHasWarning(t, node.Warnings, WarningNodeMultipleLockfiles)
}

func TestDetectPHP(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "composer.lock")

	results, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}

	php := assertContainsEcosystem(t, results, EcosystemPHP)
	if php.Ecosystem != EcosystemPHP {
		t.Fatalf("expected php ecosystem, got %q", php.Ecosystem)
	}
	if len(php.MatchedFiles) != 1 {
		t.Fatalf("expected one matched PHP file, got %#v", php.MatchedFiles)
	}
	assertHasFile(t, php.MatchedFiles, "composer.lock")
}

func TestDetectGo(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "go.mod")

	results, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}

	goResult := assertContainsEcosystem(t, results, EcosystemGo)
	if goResult.Ecosystem != EcosystemGo {
		t.Fatalf("expected go ecosystem, got %q", goResult.Ecosystem)
	}
	if len(goResult.MatchedFiles) != 1 {
		t.Fatalf("expected one matched Go file, got %#v", goResult.MatchedFiles)
	}
	assertHasFile(t, goResult.MatchedFiles, "go.mod")
}

func TestDetectRust(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "Cargo.toml")

	results, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}

	rust := assertContainsEcosystem(t, results, EcosystemRust)
	if rust.Ecosystem != EcosystemRust {
		t.Fatalf("expected rust ecosystem, got %q", rust.Ecosystem)
	}
	if len(rust.MatchedFiles) != 1 {
		t.Fatalf("expected one matched Rust file, got %#v", rust.MatchedFiles)
	}
	assertHasFile(t, rust.MatchedFiles, "cargo.toml")
}

func TestDetectPythonSignals(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "pyproject.toml", "poetry.lock", "uv.lock", "Pipfile.lock", "requirements.txt")

	results, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}

	python := assertContainsEcosystem(t, results, EcosystemPython)
	// D-08 requires aggregated matched files when Python signals coexist.
	if len(python.MatchedFiles) != 5 {
		t.Fatalf("expected one python result containing all matched signals, got %#v", python.MatchedFiles)
	}
	assertHasFile(t, python.MatchedFiles, "pipfile.lock")
	assertHasFile(t, python.MatchedFiles, "pyproject.toml")
	assertHasFile(t, python.MatchedFiles, "poetry.lock")
	assertHasFile(t, python.MatchedFiles, "uv.lock")
	assertHasFile(t, python.MatchedFiles, "requirements.txt")
}

func TestDetectMultiEcosystemDeterministic(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "go.mod", "composer.lock", "bun.lock", "Cargo.toml", "requirements.txt")

	results, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}

	expectedOrder := []Ecosystem{
		EcosystemNode,
		EcosystemPHP,
		EcosystemGo,
		EcosystemRust,
		EcosystemPython,
	}
	if len(results) != len(expectedOrder) {
		t.Fatalf("expected %d ecosystems, got %d: %#v", len(expectedOrder), len(results), results)
	}

	for i, ecosystem := range expectedOrder {
		if results[i].Ecosystem != ecosystem {
			t.Fatalf("unexpected ecosystem order at %d: got %q want %q", i, results[i].Ecosystem, ecosystem)
		}
	}

	assertHasFile(t, results[0].MatchedFiles, "bun.lock")
	assertHasFile(t, results[1].MatchedFiles, "composer.lock")
	assertHasFile(t, results[2].MatchedFiles, "go.mod")
	assertHasFile(t, results[3].MatchedFiles, "cargo.toml")
	assertHasFile(t, results[4].MatchedFiles, "requirements.txt")
}

func TestDetectCaseInsensitiveCanonicalSignals(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "PACKAGE-LOCK.JSON", "CARGO.TOML", "PIPFILE.LOCK")

	results, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}

	node := assertContainsEcosystem(t, results, EcosystemNode)
	assertHasFile(t, node.MatchedFiles, "package-lock.json")
	if len(node.Managers) != 1 || node.Managers[0] != "npm" {
		t.Fatalf("expected npm manager for package-lock.json, got %#v", node.Managers)
	}

	rust := assertContainsEcosystem(t, results, EcosystemRust)
	assertHasFile(t, rust.MatchedFiles, "cargo.toml")

	python := assertContainsEcosystem(t, results, EcosystemPython)
	assertHasFile(t, python.MatchedFiles, "pipfile.lock")
}

func TestDetectSkipsSymlinkSignals(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "real-lockfile")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatalf("write symlink target: %v", err)
	}
	if err := os.Symlink(target, filepath.Join(dir, "composer.lock")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	results, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected symlinked signal to be ignored, got %#v", results)
	}
}

func TestDetectSingleNodeLockfileHasNoWarning(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "yarn.lock")

	results, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}

	node := assertContainsEcosystem(t, results, EcosystemNode)
	if len(node.Warnings) != 0 {
		t.Fatalf("expected no node ambiguity warnings for single lockfile, got %#v", node.Warnings)
	}
}
