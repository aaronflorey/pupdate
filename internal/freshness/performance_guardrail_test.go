package freshness

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aaronflorey/pupdate/internal/state"
)

func TestHashMatchedFilesStoredHashReuseLatencyGuardrail(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bun.lock")
	payload := make([]byte, 1<<20)
	for i := range payload {
		payload[i] = byte('a' + (i % 26))
	}
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		t.Fatalf("write bun.lock: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat bun.lock: %v", err)
	}
	metadata := metadataForFile(info)
	hash, err := hashFile(path)
	if err != nil {
		t.Fatalf("hash bun.lock: %v", err)
	}

	reuse := testing.Benchmark(func(b *testing.B) {
		previous := state.EcosystemState{
			Lockfiles: map[string]string{"bun.lock": hash},
			LockfileMetadata: map[string]state.LockfileMetadata{
				"bun.lock": metadata,
			},
		}

		for i := 0; i < b.N; i++ {
			if _, _, err := hashMatchedFiles(dir, []string{"bun.lock"}, previous); err != nil {
				b.Fatalf("hashMatchedFiles reuse: %v", err)
			}
		}
	})
	rehash := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, _, err := hashMatchedFiles(dir, []string{"bun.lock"}, state.EcosystemState{}); err != nil {
				b.Fatalf("hashMatchedFiles rehash: %v", err)
			}
		}
	})

	reuseNs := reuse.NsPerOp()
	rehashNs := rehash.NsPerOp()
	if reuseNs <= 0 || rehashNs <= 0 {
		t.Fatalf("expected positive benchmark timings, reuse=%d rehash=%d", reuseNs, rehashNs)
	}
	if rehashNs < reuseNs*20 {
		t.Fatalf("expected stored-hash reuse to stay at least 20x faster than rehashing, reuse=%d ns/op rehash=%d ns/op", reuseNs, rehashNs)
	}
}
