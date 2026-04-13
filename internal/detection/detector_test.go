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

func TestDetectNodeManagersFromLockfilesDeterministic(t *testing.T) {
	dir := t.TempDir()
	writeFiles(
		t,
		dir,
		"bun.lock",
		"package-lock.json",
		"pnpm-lock.yaml",
		"yarn.lock",
		"BUN.LOCK",
		"PACKAGE-LOCK.JSON",
		"PNPM-LOCK.YAML",
		"YARN.LOCK",
		"bun.lock.bak",
		"package-lock.json.swp",
		"pnpm-lock.yaml~",
		"yarn.lock.orig",
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
	assertHasFile(t, node.MatchedFiles, "package-lock.json")
	assertHasFile(t, node.MatchedFiles, "pnpm-lock.yaml")
	assertHasFile(t, node.MatchedFiles, "yarn.lock")
	if len(node.MatchedFiles) != 4 {
		t.Fatalf("expected only canonical node lockfiles, got %#v", node.MatchedFiles)
	}
	if !slices.Equal(node.Managers, []string{"bun", "npm", "pnpm", "yarn"}) {
		t.Fatalf("unexpected node managers order/content: %#v", node.Managers)
	}
	if len(node.Warnings) != 1 {
		t.Fatalf("expected exactly one node warning for multi-lockfile case, got %#v", node.Warnings)
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

func TestDetectPythonGoRustSignals(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "requirements.txt", "uv.lock", "poetry.lock", "go.mod", "cargo.lock")

	results, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected python/go/rust ecosystems, got %#v", results)
	}

	python := assertContainsEcosystem(t, results, EcosystemPython)
	if !slices.Equal(python.MatchedFiles, []string{"poetry.lock", "requirements.txt", "uv.lock"}) {
		t.Fatalf("unexpected python matched files: %#v", python.MatchedFiles)
	}

	goResult := assertContainsEcosystem(t, results, EcosystemGo)
	if !slices.Equal(goResult.MatchedFiles, []string{"go.mod"}) {
		t.Fatalf("unexpected go matched files: %#v", goResult.MatchedFiles)
	}

	rust := assertContainsEcosystem(t, results, EcosystemRust)
	if !slices.Equal(rust.MatchedFiles, []string{"cargo.lock"}) {
		t.Fatalf("unexpected rust matched files: %#v", rust.MatchedFiles)
	}
}

func TestDetectMultiEcosystemDeterministic(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "composer.lock", "bun.lock", "go.mod", "cargo.lock", "requirements.txt")

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
	assertHasFile(t, results[3].MatchedFiles, "cargo.lock")
	assertHasFile(t, results[4].MatchedFiles, "requirements.txt")
}

func TestDetectCaseInsensitiveCanonicalSignals(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "BUN.LOCK", "COMPOSER.LOCK", "PACKAGE-LOCK.JSON", "GO.MOD", "CARGO.LOCK", "REQUIREMENTS.TXT")

	results, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}

	node := assertContainsEcosystem(t, results, EcosystemNode)
	assertHasFile(t, node.MatchedFiles, "bun.lock")
	if !slices.Equal(node.Managers, []string{"bun", "npm"}) {
		t.Fatalf("expected bun/npm managers for node lockfiles, got %#v", node.Managers)
	}

	php := assertContainsEcosystem(t, results, EcosystemPHP)
	assertHasFile(t, php.MatchedFiles, "composer.lock")

	goResult := assertContainsEcosystem(t, results, EcosystemGo)
	assertHasFile(t, goResult.MatchedFiles, "go.mod")

	rust := assertContainsEcosystem(t, results, EcosystemRust)
	assertHasFile(t, rust.MatchedFiles, "cargo.lock")

	python := assertContainsEcosystem(t, results, EcosystemPython)
	assertHasFile(t, python.MatchedFiles, "requirements.txt")
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
	writeFiles(t, dir, "bun.lock")

	results, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}

	node := assertContainsEcosystem(t, results, EcosystemNode)
	if len(node.Warnings) != 0 {
		t.Fatalf("expected no node ambiguity warnings for single lockfile, got %#v", node.Warnings)
	}
}

func TestDetectIncludesDepthOneSubdirectories(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "frontend"), 0o755); err != nil {
		t.Fatalf("mkdir frontend: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "backend"), 0o755); err != nil {
		t.Fatalf("mkdir backend: %v", err)
	}
	writeFiles(t, filepath.Join(dir, "frontend"), "package-lock.json")
	writeFiles(t, filepath.Join(dir, "backend"), "composer.lock")

	results, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 ecosystems from depth-1 subdirectories, got %#v", results)
	}

	node := assertContainsEcosystem(t, results, EcosystemNode)
	if node.Directory != "frontend" {
		t.Fatalf("expected frontend directory for node detection, got %q", node.Directory)
	}
	assertHasFile(t, node.MatchedFiles, "frontend/package-lock.json")
	if node.StateKey() != "node@frontend" {
		t.Fatalf("unexpected node state key: %q", node.StateKey())
	}

	php := assertContainsEcosystem(t, results, EcosystemPHP)
	if php.Directory != "backend" {
		t.Fatalf("expected backend directory for php detection, got %q", php.Directory)
	}
	assertHasFile(t, php.MatchedFiles, "backend/composer.lock")
	if php.StateKey() != "php@backend" {
		t.Fatalf("unexpected php state key: %q", php.StateKey())
	}
}

func TestDetectDoesNotScanPastDepthOne(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "apps", "web"), 0o755); err != nil {
		t.Fatalf("mkdir apps/web: %v", err)
	}
	writeFiles(t, filepath.Join(dir, "apps", "web"), "package-lock.json")

	results, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("expected no detections beyond depth-1, got %#v", results)
	}
}

func TestDetectIncludesPackagesChildren(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "packages", "web"), 0o755); err != nil {
		t.Fatalf("mkdir packages/web: %v", err)
	}
	writeFiles(t, filepath.Join(dir, "packages", "web"), "package-lock.json")

	results, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected one ecosystem from packages child, got %#v", results)
	}

	node := assertContainsEcosystem(t, results, EcosystemNode)
	if node.Directory != "packages/web" {
		t.Fatalf("expected packages/web directory for node detection, got %q", node.Directory)
	}
	assertHasFile(t, node.MatchedFiles, "packages/web/package-lock.json")
}

func TestDetectDoesNotScanPastPackagesChildren(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "packages", "group", "web"), 0o755); err != nil {
		t.Fatalf("mkdir packages/group/web: %v", err)
	}
	writeFiles(t, filepath.Join(dir, "packages", "group", "web"), "package-lock.json")

	results, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("expected no detections beyond packages/*, got %#v", results)
	}
}

func TestDetectSkipsGitignoredDirectories(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "frontend"), 0o755); err != nil {
		t.Fatalf("mkdir frontend: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "packages", "legacy"), 0o755); err != nil {
		t.Fatalf("mkdir packages/legacy: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "packages", "active"), 0o755); err != nil {
		t.Fatalf("mkdir packages/active: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("frontend/\npackages/legacy/\n"), 0o644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}

	writeFiles(t, filepath.Join(dir, "frontend"), "package-lock.json")
	writeFiles(t, filepath.Join(dir, "packages", "legacy"), "composer.lock")
	writeFiles(t, filepath.Join(dir, "packages", "active"), "package-lock.json")

	results, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected only non-ignored detection, got %#v", results)
	}

	node := assertContainsEcosystem(t, results, EcosystemNode)
	if node.Directory != "packages/active" {
		t.Fatalf("expected packages/active directory for node detection, got %q", node.Directory)
	}
	assertHasFile(t, node.MatchedFiles, "packages/active/package-lock.json")
}
