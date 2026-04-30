---
phase: 11-kasetto-and-freshness-correctness-follow-ups
plan: 02
subsystem: kasetto-execution
tags: [kasetto, config, detection]
requires: [11-01]
provides:
  - Explicit `--config` wiring for local `kasetto.yaml` and `kasetto.yml` detections
  - Regression coverage for lock-only Kasetto skips and yml config selection
affects: [cmd-pupdate]
tech-stack:
  added: []
  patterns: [matched-file-selection, config-aware-execution]
key-files:
  created:
    - .planning/phases/11-kasetto-and-freshness-correctness-follow-ups/11-02-SUMMARY.md
  modified:
    - cmd/pupdate/run_install.go
    - cmd/pupdate/run_test.go
key-decisions:
  - "Treat only `kasetto.yaml` and `kasetto.yml` as executable local configs, passed through verbatim from detection to `kst sync --config`."
  - "Skip lock-only Kasetto detections rather than allowing execution to fall back to global or default config discovery."
requirements-completed: []
duration: 7m
completed: 2026-04-30
---

# Phase 11 Plan 02: Kasetto config alignment summary

Phase 11 plan 02 aligns Kasetto detection and execution by passing detected local config files explicitly and skipping lock-only detections that cannot be tied to a local project config.

## Verification

- `go test ./cmd/pupdate ./internal/detection -count=1`

## Files Created/Modified

- `cmd/pupdate/run_install.go` - Selects explicit Kasetto config paths from matched files and skips lock-only detections.
- `cmd/pupdate/run_test.go` - Verifies explicit `--config` args for yaml/yml configs and the lock-only skip path.
