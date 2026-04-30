---
phase: 15-performance-diagnostics-and-config-hook-follow-ups
plan: 02
subsystem: diagnostics
tags: [diagnostics, cli, config, state, troubleshooting]
requires: []
provides:
  - Read-only `pupdate status` command for current-directory diagnostics
  - Per-target reporting for freshness decisions, manager readiness, and install blocking reasons
  - Command coverage for ready, repo-skip, config-parse-error, and missing-manager scenarios
affects: [cmd-pupdate, docs]
tech-stack:
  added: []
  patterns: [read-only-diagnostics, reuse-run-pipeline, per-target-status-reporting]
key-files:
  created:
    - .planning/phases/15-performance-diagnostics-and-config-hook-follow-ups/15-02-SUMMARY.md
    - cmd/pupdate/status.go
    - cmd/pupdate/status_test.go
  modified:
    - cmd/pupdate/root.go
    - README.md
key-decisions:
  - "Use `pupdate status` as the first dedicated troubleshooting surface because it matches the existing command naming and can describe current state without implying repair or mutation."
  - "Reuse the existing run/config/freshness/PATH logic so status output reflects real execution behavior instead of a parallel diagnostic implementation."
  - "Keep the command read-only and stdout-oriented while reporting repo-level skip reasons, state warnings, and per-target install blockers."
requirements-completed: []
duration: 26m
completed: 2026-04-30
---

# Phase 15 Plan 02: Diagnostic command summary

Phase 15 plan 02 adds a read-only `pupdate status` command that explains what `pupdate run` would do in the current directory by surfacing repo-level skip conditions, config and state context, freshness decisions, and per-target install readiness or blocking reasons.

## Verification

- `go test ./cmd/pupdate -count=1`
- `go test ./... -count=1`

## Files Created/Modified

- `cmd/pupdate/status.go` - Collects and prints the diagnostic snapshot using the real run/config/state/freshness logic without mutating anything.
- `cmd/pupdate/status_test.go` - Covers ready, repo-skip, invalid-config, and missing-manager troubleshooting flows.
- `cmd/pupdate/root.go` - Registers the new `status` command.
- `README.md` - Documents the new command and its troubleshooting output.
