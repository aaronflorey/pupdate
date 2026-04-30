---
phase: 11-kasetto-and-freshness-correctness-follow-ups
plan: 01
subsystem: execution
tags: [kasetto, scope, execution]
requires: []
provides:
  - Project-scoped Kasetto execution args
  - Regression coverage for the emitted project scope flag
affects: [cmd-pupdate]
tech-stack:
  added: []
  patterns: [manager-arg-hardening, command-regression-tests]
key-files:
  created:
    - .planning/phases/11-kasetto-and-freshness-correctness-follow-ups/11-01-SUMMARY.md
  modified:
    - cmd/pupdate/run_install.go
    - cmd/pupdate/run_test.go
key-decisions:
  - "Force Kasetto runs through `kst sync --project` so detected projects do not mutate ambient global state."
  - "Keep the change isolated to execution args so the follow-up config-selection work can layer on independently."
requirements-completed: []
duration: 4m
completed: 2026-04-30
---

# Phase 11 Plan 01: Kasetto project scope summary

Phase 11 plan 01 hardens Kasetto execution by making every detected Kasetto run explicitly project-scoped with `--project`.

## Verification

- `go test ./cmd/pupdate -count=1`

## Files Created/Modified

- `cmd/pupdate/run_install.go` - Adds `--project` to the Kasetto sync command.
- `cmd/pupdate/run_test.go` - Verifies Kasetto plan args and emitted run lines include project scope.
