---
phase: 11-kasetto-and-freshness-correctness-follow-ups
plan: 03
subsystem: freshness
tags: [freshness, hashing, correctness]
requires: [11-02]
provides:
  - Content-correct lockfile hashing even when file metadata is unchanged
  - Regression coverage for same-metadata rewrites and forced rehash behavior
affects: [internal-freshness]
tech-stack:
  added: []
  patterns: [content-hash-verification, metadata-regression-tests]
key-files:
  created:
    - .planning/phases/11-kasetto-and-freshness-correctness-follow-ups/11-03-SUMMARY.md
  modified:
    - internal/freshness/engine.go
    - internal/freshness/engine_test.go
key-decisions:
  - "Remove metadata-only hash reuse rather than trusting size, mode, and mtime as proof of unchanged lockfile content."
  - "Keep persisted lockfile metadata intact for compatibility and future use, but require a real content hash for skip decisions."
requirements-completed: []
duration: 5m
completed: 2026-04-30
---

# Phase 11 Plan 03: Freshness correctness summary

Phase 11 plan 03 removes the unsafe metadata-only lockfile hash reuse path so freshness decisions are once again based on actual file contents even when a rewrite preserves size, mode, and timestamp.

## Verification

- `go test ./internal/freshness -count=1`

## Files Created/Modified

- `internal/freshness/engine.go` - Always rehashes matched lockfiles instead of trusting stored metadata equality.
- `internal/freshness/engine_test.go` - Verifies rehashing still happens when metadata matches and catches same-metadata content changes.
