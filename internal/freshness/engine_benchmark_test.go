package freshness

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aaronflorey/pupdate/internal/state"
)

func BenchmarkHashMatchedFiles(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "bun.lock")
	payload := make([]byte, 1<<20)
	for i := range payload {
		payload[i] = byte('a' + (i % 26))
	}
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		b.Fatalf("write bun.lock: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		b.Fatalf("stat bun.lock: %v", err)
	}
	metadata := metadataForFile(info)
	hash, err := hashFile(path)
	if err != nil {
		b.Fatalf("hash bun.lock: %v", err)
	}

	b.Run("reuse stored hash", func(b *testing.B) {
		previous := state.EcosystemState{
			Lockfiles: map[string]string{"bun.lock": hash},
			LockfileMetadata: map[string]state.LockfileMetadata{
				"bun.lock": metadata,
			},
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, _, err := hashMatchedFiles(dir, []string{"bun.lock"}, previous); err != nil {
				b.Fatalf("hashMatchedFiles: %v", err)
			}
		}
	})

	b.Run("rehash file", func(b *testing.B) {
		previous := state.EcosystemState{}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, _, err := hashMatchedFiles(dir, []string{"bun.lock"}, previous); err != nil {
				b.Fatalf("hashMatchedFiles: %v", err)
			}
		}
	})
}
