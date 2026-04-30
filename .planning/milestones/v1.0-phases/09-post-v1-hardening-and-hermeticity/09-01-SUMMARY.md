---
phase: 09-post-v1-hardening-and-hermeticity
plan: 01
subsystem: detection
tags: [rust, lockfiles, freshness, compatibility]
requires: []
provides:
  - Canonical `Cargo.lock` detection results on case-sensitive filesystems
  - Regression coverage for Rust detection and freshness compatibility
affects: [internal-detection, internal-freshness]
tech-stack:
  added: []
  patterns: [canonical-signal-preservation, backward-compatible-state-keys]
key-files:
  created:
    - .planning/phases/09-post-v1-hardening-and-hermeticity/09-CONTEXT.md
    - .planning/phases/09-post-v1-hardening-and-hermeticity/09-01-SUMMARY.md
  modified:
    - internal/detection/matrix.go
    - internal/detection/detector_test.go
    - internal/freshness/engine_test.go
key-decisions:
  - "Preserve the canonical `Cargo.lock` filename in detection results so freshness hashes the real on-disk path."
  - "Keep lowercase lockfile keys in freshness state comparisons so existing stored state remains compatible."
requirements-completed: []
duration: 10m
completed: 2026-04-29
---

# Phase 09 Plan 01: Cargo.lock case handling summary

Phase 09 plan 01 fixes the Rust lockfile path contract by preserving the canonical `Cargo.lock` filename in detection results while keeping lowercase freshness state keys for compatibility with existing `.pupdate` data.

## Verification

- `go test ./internal/detection -count=1`
- `go test ./internal/freshness -count=1`

## Files Created/Modified

- `internal/detection/matrix.go` - Stops lowercasing canonical signal names before they reach detection results.
- `internal/detection/detector_test.go` - Updates Rust expectations to the canonical `Cargo.lock` filename.
- `internal/freshness/engine_test.go` - Adds a regression proving canonical Rust filenames still compare against lowercase stored state keys.
- `.planning/phases/09-post-v1-hardening-and-hermeticity/09-CONTEXT.md` - Records the discuss/context artifact for autonomous execution.
