---
phase: 10-filesystem-case-sensitivity-follow-ups
plan: 01
subsystem: config
tags: [filesystem, case-sensitivity, roots]
requires: []
provides:
  - Filesystem-aware `root_directories` matching without unconditional case folding
  - Regression coverage for mixed-case root mismatches in config and run behavior
affects: [cmd-pupdate]
tech-stack:
  added: []
  patterns: [path-normalization, command-regression-tests]
key-files:
  created:
    - .planning/phases/10-filesystem-case-sensitivity-follow-ups/10-01-SUMMARY.md
  modified:
    - cmd/pupdate/config.go
    - cmd/pupdate/config_test.go
    - cmd/pupdate/run_test.go
key-decisions:
  - "Keep root matching aligned with resolved path casing instead of forcing global case-insensitive comparisons."
  - "Preserve existing `~` expansion, top-level-only matching, and quiet-mode skip behavior while correcting the case-sensitive mismatch path."
requirements-completed: []
duration: 5m
completed: 2026-04-30
---

# Phase 10 Plan 01: Root directory matching summary

Phase 10 plan 01 removes unconditional lowercase path comparison from `root_directories` matching so configured roots only match when the resolved directory casing actually lines up on case-sensitive filesystems.

## Verification

- `go test ./cmd/pupdate -count=1`

## Files Created/Modified

- `cmd/pupdate/config.go` - Stops lowercasing normalized directories before root containment checks.
- `cmd/pupdate/config_test.go` - Verifies a differently cased configured root no longer matches.
- `cmd/pupdate/run_test.go` - Verifies a mixed-case configured root causes a quiet skip instead of running installs.
