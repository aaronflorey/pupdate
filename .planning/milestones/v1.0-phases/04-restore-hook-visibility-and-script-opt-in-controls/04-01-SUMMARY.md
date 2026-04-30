---
phase: 04-restore-hook-visibility-and-script-opt-in-controls
plan: 01
subsystem: cli
tags: [shell-hooks, safety, docs, tests]
requires: []
provides:
  - Hook snippets that preserve stderr status visibility during quiet runs
  - `run --allow-scripts` opt-in for supported managers
  - Regression coverage for hook visibility and script opt-in behavior
affects: [shell-integration, install-safety, docs]
tech-stack:
  added: []
  patterns: [quiet-stdout-only hook mode, explicit script opt-in flag]
key-files:
  created:
    - .planning/phases/04-restore-hook-visibility-and-script-opt-in-controls/04-01-PLAN.md
    - .planning/phases/04-restore-hook-visibility-and-script-opt-in-controls/04-01-SUMMARY.md
    - .planning/phases/04-restore-hook-visibility-and-script-opt-in-controls/04-VALIDATION.md
  modified:
    - cmd/pupdate/init.go
    - cmd/pupdate/run.go
    - cmd/pupdate/init_test.go
    - cmd/pupdate/run_test.go
    - README.md
    - .planning/ROADMAP.md
    - .planning/REQUIREMENTS.md
    - .planning/PROJECT.md
    - .planning/STATE.md
key-decisions:
  - "Keep `--quiet` limited to stdout/child-process noise suppression so concise stderr status remains visible in shell hooks."
  - "Expose lifecycle scripts only through an explicit `--allow-scripts` flag to preserve safe defaults."
requirements-completed: [EXEC-03, STAT-01, MILE-01]
duration: 12m
completed: 2026-04-08
---

# Phase 4 Plan 01: Hook visibility and script opt-in summary

Hook-generated `pupdate run --quiet` snippets now preserve stderr status lines, which restores visible run/skip/error feedback without re-enabling noisy child command output.

`pupdate run --allow-scripts` now provides the missing explicit opt-in for lifecycle scripts by removing script-blocking flags only for package managers that support them.

## Verification

- `go test ./cmd/pupdate -count=1`
- `go test ./... -count=1`

## Files Created/Modified

- `cmd/pupdate/init.go` - Removes shell-level stderr redirection from generated hook snippets.
- `cmd/pupdate/run.go` - Adds `--allow-scripts` and threads it into manager plan selection.
- `cmd/pupdate/init_test.go` - Guards visible hook stderr behavior in generated snippets.
- `cmd/pupdate/run_test.go` - Adds quiet stderr visibility and script opt-in regression coverage.
- `README.md` - Documents visible hook status output and the new opt-in flag.
