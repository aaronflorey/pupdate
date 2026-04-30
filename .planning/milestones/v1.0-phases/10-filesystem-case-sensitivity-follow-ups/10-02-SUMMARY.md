---
phase: 10-filesystem-case-sensitivity-follow-ups
plan: 02
subsystem: detection-freshness
tags: [filesystem, case-sensitivity, lockfiles]
requires: [10-01]
provides:
  - Actual matched lockfile paths preserved through detection results
  - Regression coverage for mixed-case lockfiles that must still compare against existing lowercase state keys
affects: [internal-detection, internal-freshness]
tech-stack:
  added: []
  patterns: [canonical-dedupe-with-real-paths, mixed-case-lockfile-regression-tests]
key-files:
  created:
    - .planning/phases/10-filesystem-case-sensitivity-follow-ups/10-02-SUMMARY.md
  modified:
    - internal/detection/detector.go
    - internal/detection/detector_test.go
    - internal/freshness/engine_test.go
key-decisions:
  - "Keep canonical signal names for dedupe and manager lookup, but carry the actual on-disk filename in `MatchedFiles`."
  - "Preserve compatibility with existing lowercase freshness state keys by keeping hash/state normalization unchanged."
requirements-completed: []
duration: 8m
completed: 2026-04-30
---

# Phase 10 Plan 02: Matched lockfile path summary

Phase 10 plan 02 preserves the real matched lockfile path casing from directory scans so freshness hashes the file that actually exists on disk, while continuing to compare against the existing lowercase state keys.

## Verification

- `go test ./internal/detection ./internal/freshness -count=1`

## Files Created/Modified

- `internal/detection/detector.go` - Preserves actual entry names in `MatchedFiles` while still deduplicating by canonical signal identity.
- `internal/detection/detector_test.go` - Covers uppercase-only lockfiles and preserved casing in subdirectories.
- `internal/freshness/engine_test.go` - Verifies mixed-case lockfiles still compare successfully against lowercase stored state keys.
