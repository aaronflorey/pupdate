---
phase: 15-performance-diagnostics-and-config-hook-follow-ups
plan: 05
subsystem: config
tags: [config, run, status, docs]
requires: []
provides:
  - Config-backed defaults for `quiet` and `allow_scripts`
  - Shared run-option resolution across `run` and `status`
  - Documentation and regression coverage for expanded config surface
affects: [cmd-pupdate, docs]
tech-stack:
  added: []
  patterns: [config-defaults, flag-overrides-config, shared-run-option-resolution]
key-files:
  created:
    - .planning/phases/15-performance-diagnostics-and-config-hook-follow-ups/15-05-SUMMARY.md
  modified:
    - README.md
    - .planning/ROADMAP.md
    - .planning/STATE.md
    - cmd/pupdate/config.go
    - cmd/pupdate/config_cmd.go
    - cmd/pupdate/config_cmd_test.go
    - cmd/pupdate/config_test.go
    - cmd/pupdate/run.go
    - cmd/pupdate/run_execution.go
    - cmd/pupdate/run_test.go
    - cmd/pupdate/status.go
    - cmd/pupdate/status_test.go
key-decisions:
  - "Expand the existing config file with the already-supported run knobs `quiet` and `allow_scripts` instead of inventing a broader new config hierarchy."
  - "Let explicit CLI flags override config values so manual one-off runs can still opt back into verbose output or safe script blocking."
  - "Drive `status` from the same effective run-option resolution as `run` so diagnostics reflect real configured behavior."
requirements-completed: []
duration: 18m
completed: 2026-04-30
---

# Phase 15 Plan 05: Config expansion summary

Phase 15 plan 05 broadens the user config surface in the smallest useful way by promoting the existing `run` behavior knobs `quiet` and `allow_scripts` into `config.yaml`, then reusing those effective defaults in both `pupdate run` and `pupdate status`.

## Verification

- `go test ./cmd/pupdate`
- `go test ./...`

## Files Created/Modified

- `cmd/pupdate/config.go` - Adds `quiet` and `allow_scripts` to the user config schema.
- `cmd/pupdate/run_execution.go` - Resolves effective run options from config first, with explicit flags taking precedence.
- `cmd/pupdate/status.go` - Shows configured/effective run defaults and uses them when rendering install readiness.
- `cmd/pupdate/config_cmd.go` - Prints the expanded config surface for operator visibility.
- `cmd/pupdate/*_test.go` - Covers config defaults, explicit flag overrides, and status/config output.
- `README.md` - Documents the new config keys and where they appear in diagnostics.
