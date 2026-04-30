---
phase: 15-performance-diagnostics-and-config-hook-follow-ups
plan: 04
subsystem: performance
tags: [performance, benchmarks, ci, freshness, detection]
requires: []
provides:
  - Focused detection and freshness hot-path benchmarks
  - Relative-speed guardrail for stored-hash reuse versus full rehash
  - CI benchmark execution for automatic visibility into latency drift
affects: [internal-detection, internal-freshness, ci]
tech-stack:
  added: []
  patterns: [hot-path-benchmark, relative-performance-guardrail, ci-benchmark-run]
key-files:
  created:
    - .planning/phases/15-performance-diagnostics-and-config-hook-follow-ups/15-04-SUMMARY.md
    - internal/detection/detector_benchmark_test.go
    - internal/freshness/performance_guardrail_test.go
  modified:
    - .github/workflows/ci.yml
    - .planning/ROADMAP.md
    - .planning/STATE.md
key-decisions:
  - "Guard the stored-hash fast path with a relative benchmark comparison instead of an absolute timing threshold so CI remains stable across runners."
  - "Benchmark both detection and freshness because those paths most directly affect shell-hook latency on directory entry."
  - "Run the targeted benchmarks in CI logs for visibility while keeping the hard fail condition in a normal test so regressions are caught automatically."
requirements-completed: []
duration: 22m
completed: 2026-04-30
---

# Phase 15 Plan 04: Performance guardrails summary

Phase 15 plan 04 adds explicit measurements for the two hottest directory-entry paths, then protects the highest-value freshness shortcut with a CI-safe relative latency guardrail instead of a flaky absolute runtime budget.

## Verification

- `go test ./internal/detection ./internal/freshness -count=1`
- `go test ./... -count=1`
- `go test ./internal/detection ./internal/freshness -run '^$' -bench 'Benchmark(DetectProjectTree|HashMatchedFiles)$' -count=1`

Benchmark snapshot on Linux:

- `BenchmarkDetectProjectTree` about `93954 ns/op`
- `BenchmarkHashMatchedFiles/reuse_stored_hash` about `1675 ns/op`
- `BenchmarkHashMatchedFiles/rehash_file` about `2085555 ns/op`

## Files Created/Modified

- `internal/detection/detector_benchmark_test.go` - Benchmarks detection over a representative multi-ecosystem project tree.
- `internal/freshness/performance_guardrail_test.go` - Fails if stored-hash reuse stops being meaningfully faster than full lockfile rehashing.
- `.github/workflows/ci.yml` - Runs the targeted hot-path benchmarks automatically in CI.
- `.planning/ROADMAP.md` - Marks Phase 15 plans 03 and 04 complete.
- `.planning/STATE.md` - Advances the active plan to 15-05 after completing the performance work.
