---
phase: 15-performance-diagnostics-and-config-hook-follow-ups
plan: 03
subsystem: state
tags: [state, cleanup, freshness, lifecycle]
requires: []
provides:
  - Pruning of stale ecosystem state entries when targets are no longer detected
  - Persistence of cleanup-only state saves even when no install runs
  - Regression coverage for removed subdirectory targets and active-target retention
affects: [cmd-pupdate, internal-state]
tech-stack:
  added: []
  patterns: [detected-target-pruning, cleanup-only-save, active-state-retention]
key-files:
  created:
    - .planning/phases/15-performance-diagnostics-and-config-hook-follow-ups/15-03-SUMMARY.md
  modified:
    - cmd/pupdate/run_execution.go
    - cmd/pupdate/run_state.go
    - cmd/pupdate/run_state_test.go
    - cmd/pupdate/run_test.go
key-decisions:
  - "Prune only ecosystem entries whose targets are no longer detected, rather than persisting changed lockfile hashes without a successful install."
  - "Allow cleanup-only state saves so removed root or subdirectory targets disappear from `.pupdate` even when the current run only skips or has installs disabled."
  - "Retain active ecosystem entries unchanged unless a success or metadata-refresh outcome explicitly updates them."
requirements-completed: []
duration: 18m
completed: 2026-04-30
---

# Phase 15 Plan 03: Stale state pruning summary

Phase 15 plan 03 keeps `.pupdate` truthful by removing ecosystem entries for targets that are no longer detected while preserving active entries until a successful install or metadata refresh can safely update them.

## Verification

- `go test ./cmd/pupdate ./internal/state ./internal/freshness -count=1`
- `go test ./... -count=1`

## Files Created/Modified

- `cmd/pupdate/run_execution.go` - Passes active detection results into state persistence so cleanup can consider the current target set.
- `cmd/pupdate/run_state.go` - Prunes undetected ecosystem keys, saves cleanup-only changes, and keeps active entries unless outcomes update them.
- `cmd/pupdate/run_state_test.go` - Covers pruning of removed targets and retention of active ecosystems without persisted outcomes.
- `cmd/pupdate/run_test.go` - Verifies end-to-end pruning of stale subdirectory state while unchanged active targets remain valid.
