package detection

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkDetectProjectTree(b *testing.B) {
	dir := b.TempDir()
	writeBenchmarkFile(b, dir, ".gitmodules")
	writeBenchmarkFile(b, dir, "bun.lock")
	writeBenchmarkFile(b, dir, "composer.lock")
	writeBenchmarkFile(b, dir, "go.mod")
	writeBenchmarkFile(b, dir, "Cargo.lock")
	writeBenchmarkFile(b, dir, "requirements.txt")
	writeBenchmarkFile(b, dir, "kasetto.yaml")
	writeBenchmarkFile(b, dir, "frontend/package-lock.json")
	writeBenchmarkFile(b, dir, "backend/poetry.lock")
	writeBenchmarkFile(b, dir, "packages/app-one/bun.lock")
	writeBenchmarkFile(b, dir, "packages/app-two/composer.lock")
	writeBenchmarkFile(b, dir, "packages/app-three/go.mod")
	writeBenchmarkFile(b, dir, "packages/app-four/Cargo.lock")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Detect(dir); err != nil {
			b.Fatalf("Detect: %v", err)
		}
	}
}

func writeBenchmarkFile(b *testing.B, dir string, rel string) {
	b.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		b.Fatalf("mkdir %s: %v", filepath.Dir(rel), err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		b.Fatalf("write %s: %v", rel, err)
	}
}
