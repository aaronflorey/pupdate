---
phase: 09-post-v1-hardening-and-hermeticity
plan: 03
subsystem: freshness
tags: [git, submodules, timeout, testability]
requires: [09-02]
provides:
  - Timeout-bounded git submodule freshness checks
  - Injectable command execution seam for submodule status tests
affects: [internal-freshness, cmd-pupdate]
tech-stack:
  added: []
  patterns: [context-timeout, injectable-command-runner]
key-files:
  created:
    - .planning/phases/09-post-v1-hardening-and-hermeticity/09-03-SUMMARY.md
  modified:
    - internal/freshness/engine.go
    - internal/freshness/engine_test.go
key-decisions:
  - "Keep the existing `Evaluate` contract and inject only the git submodule command runner so the non-blocking behavior stays local to `internal/freshness`."
  - "Surface timeouts as explicit freshness errors (`timed out after ...`) so hook output stays visible without hanging the run path."
requirements-completed: []
duration: 8m
completed: 2026-04-29
---

# Phase 09 Plan 03: Git submodule timeout summary

Phase 09 plan 03 prevents submodule freshness checks from hanging by running `git submodule status --recursive` behind a context timeout and a testable command runner seam.

## Verification

- `go test ./internal/freshness ./cmd/pupdate -count=1`

## Files Created/Modified

- `internal/freshness/engine.go` - Adds the timeout-bounded submodule status runner and preserves stderr-rich command failures.
- `internal/freshness/engine_test.go` - Covers timeout handling, stderr propagation, and the surfaced skip reason for submodule status failures.
