---
phase: 09-post-v1-hardening-and-hermeticity
plan: 02
subsystem: command-tests
tags: [tests, config, hermeticity]
requires: [09-01]
provides:
  - Hermetic `cmd/pupdate` tests that do not read the developer's real config home
  - A reusable package-level test sandbox for config-path dependent command tests
affects: [cmd-pupdate, test-harness]
tech-stack:
  added: []
  patterns: [package-test-sandbox, minimal-test-seam]
key-files:
  created:
    - .planning/phases/09-post-v1-hardening-and-hermeticity/09-02-SUMMARY.md
    - cmd/pupdate/test_env_test.go
  modified: []
key-decisions:
  - "Isolate `cmd/pupdate` tests from ambient config with a package-level `TestMain` instead of refactoring command code that already has narrower plan-specific seams."
requirements-completed: []
duration: 8m
completed: 2026-04-29
---

# Phase 09 Plan 02: Hermetic command tests summary

Phase 09 plan 02 removes ambient user-config coupling from `cmd/pupdate` tests by forcing the package test process to use a temporary `XDG_CONFIG_HOME` unless a test explicitly overrides it.

## Verification

- `go test ./cmd/pupdate -count=1`

## Files Created/Modified

- `cmd/pupdate/test_env_test.go` - Adds a package `TestMain` that runs all command tests with an isolated config home.
