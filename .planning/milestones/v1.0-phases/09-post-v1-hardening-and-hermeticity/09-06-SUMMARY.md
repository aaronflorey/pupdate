---
phase: 09-post-v1-hardening-and-hermeticity
plan: 06
subsystem: config
tags: [config, defaults, side-effects]
requires: [09-05]
provides:
  - Side-effect-free config loading for `run`
  - Read-only `config` command behavior when the config file is missing
affects: [cmd-pupdate]
tech-stack:
  added: []
  patterns: [missing-config-defaults, read-only-config-inspection]
key-files:
  created:
    - .planning/phases/09-post-v1-hardening-and-hermeticity/09-06-SUMMARY.md
  modified:
    - cmd/pupdate/config.go
    - cmd/pupdate/config_cmd.go
    - cmd/pupdate/config_cmd_test.go
    - cmd/pupdate/run_test.go
key-decisions:
  - "Remove config auto-creation everywhere instead of only in `run` so missing config consistently behaves like defaults without hidden filesystem writes."
  - "Have `pupdate config` report `exists: false` for missing files while still showing the resolved path and default-effective values."
requirements-completed: []
duration: 7m
completed: 2026-04-29
---

# Phase 09 Plan 06: Missing config defaults summary

Phase 09 plan 06 removes automatic user-config creation and treats a missing config file as implicit default settings across both `run` and `config` flows.

## Verification

- `go test ./cmd/pupdate -count=1`
- `go test ./... -count=1`

## Files Created/Modified

- `cmd/pupdate/config.go` - Stops creating config files during normal config loading.
- `cmd/pupdate/config_cmd.go` - Makes the config command read-only for missing files and reports the actual `exists` status.
- `cmd/pupdate/config_cmd_test.go` - Verifies missing config stays missing while output still shows default-effective values.
- `cmd/pupdate/run_test.go` - Verifies `pupdate run` no longer creates a config file as a side effect.
