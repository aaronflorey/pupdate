---
phase: 15-performance-diagnostics-and-config-hook-follow-ups
plan: 01
subsystem: freshness
tags: [performance, freshness, state, metadata]
requires: []
provides:
  - Safe lockfile hash reuse gated by persisted file identity and change-time metadata
  - Metadata refresh on verified skip paths so older state files can learn the stronger fast-path fields
  - Focused benchmark coverage for reused-hash versus full-rehash hot paths
affects: [internal-freshness, internal-state, cmd-pupdate]
tech-stack:
  added: []
  patterns: [safe-metadata-fast-path, metadata-refresh-on-skip, benchmarked-hot-path]
key-files:
  created:
    - .planning/phases/15-performance-diagnostics-and-config-hook-follow-ups/15-01-SUMMARY.md
    - internal/freshness/engine_benchmark_test.go
    - internal/freshness/file_identity_linux.go
    - internal/freshness/file_identity_darwin.go
    - internal/freshness/file_identity_other.go
  modified:
    - internal/freshness/engine.go
    - internal/freshness/engine_test.go
    - internal/state/model.go
    - cmd/pupdate/run_execution.go
    - cmd/pupdate/run_state.go
    - cmd/pupdate/run_state_test.go
key-decisions:
  - "Only reuse a stored lockfile hash when size, mtime, mode, file identity, and change time all still match; otherwise fall back to a full content hash."
  - "Refresh lockfile metadata on verified skip paths without changing `last_success_at` so pre-upgrade `.pupdate` entries can gain the stronger fast-path fields."
  - "Use a platform fallback that disables the fast path where stable file identity data is unavailable rather than weakening correctness guarantees."
requirements-completed: []
duration: 24m
completed: 2026-04-30
---

# Phase 15 Plan 01: Safe lockfile metadata reuse summary

Phase 15 plan 01 restores a hot-path freshness shortcut without reviving the old metadata-only bug by requiring persisted file identity and change-time equality before reusing a stored hash, and by refreshing stronger metadata on verified skip paths so existing state can converge forward safely.

## Verification

- `go test ./internal/freshness ./internal/state ./cmd/pupdate -count=1`
- `go test ./internal/freshness -bench BenchmarkHashMatchedFiles -run '^$' -count=1`
- Benchmark snapshot on Linux for a 1 MiB lockfile: `reuse stored hash` about `2069 ns/op` vs `rehash file` about `2418320 ns/op`

## Files Created/Modified

- `internal/freshness/engine.go` - Reuses stored hashes only when stronger identity plus change-time metadata proves the file has not changed.
- `internal/freshness/engine_test.go` - Covers the new safe reuse path, old-state fallback rehashing, and same-metadata content rewrite detection.
- `internal/freshness/engine_benchmark_test.go` - Benchmarks the reused-hash and full-rehash hot paths.
- `internal/freshness/file_identity_linux.go` - Records Linux file identity and change-time metadata for the safe fast path.
- `internal/freshness/file_identity_darwin.go` - Records macOS file identity and change-time metadata for the safe fast path.
- `internal/freshness/file_identity_other.go` - Disables the fast path on unsupported platforms instead of weakening correctness.
- `internal/state/model.go` - Extends persisted lockfile metadata with optional file identity and change-time fields.
- `cmd/pupdate/run_execution.go` - Persists metadata refreshes for verified skip paths that do not execute installs.
- `cmd/pupdate/run_state.go` - Distinguishes successful install updates from metadata-only refreshes so `last_success_at` only changes on real install success.
- `cmd/pupdate/run_state_test.go` - Verifies metadata refreshes preserve the prior success timestamp.
