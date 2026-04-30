---
phase: 09-post-v1-hardening-and-hermeticity
plan: 04
subsystem: freshness
tags: [performance, lockfiles, state, metadata]
requires: [09-03]
provides:
  - Metadata-first lockfile freshness checks that skip rehashing unchanged files
  - Persisted per-lockfile metadata to support future fast-path evaluations
affects: [internal-freshness, internal-state, cmd-pupdate]
tech-stack:
  added: []
  patterns: [metadata-fast-path, backward-compatible-state-extension]
key-files:
  created:
    - .planning/phases/09-post-v1-hardening-and-hermeticity/09-04-SUMMARY.md
  modified:
    - internal/freshness/engine.go
    - internal/freshness/engine_test.go
    - internal/state/model.go
    - internal/state/model_test.go
    - cmd/pupdate/run_execution.go
    - cmd/pupdate/run_state.go
    - cmd/pupdate/run_state_test.go
key-decisions:
  - "Extend ecosystem state with optional `lockfile_metadata` instead of changing the schema version so existing `.pupdate` files remain readable and new files can accelerate future runs."
  - "Reuse stored hashes only when size, modtime, and mode all match; otherwise fall back to a fresh content hash to preserve change detection correctness."
requirements-completed: []
duration: 16m
completed: 2026-04-29
---

# Phase 09 Plan 04: Lockfile hashing fast path summary

Phase 09 plan 04 reduces hook-path hashing work by persisting lockfile metadata alongside stored hashes and reusing the old hash when the on-disk metadata is unchanged.

## Verification

- `go test ./internal/freshness -count=1`
- `go test ./cmd/pupdate -count=1`
- `go test ./internal/state -count=1`

## Files Created/Modified

- `internal/freshness/engine.go` - Adds metadata-aware lockfile comparison and only rehashes files whose metadata changed.
- `internal/freshness/engine_test.go` - Covers unchanged fast-path reuse, changed-file rehashing, renamed-path updates, and missing-file behavior.
- `internal/state/model.go` - Extends saved ecosystem state with optional lockfile metadata maps.
- `cmd/pupdate/run_execution.go` - Carries lockfile metadata through successful run outcomes.
- `cmd/pupdate/run_state.go` - Persists cloned metadata with successful state updates.
