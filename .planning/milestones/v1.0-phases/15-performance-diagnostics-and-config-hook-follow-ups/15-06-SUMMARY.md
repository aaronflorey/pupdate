---
phase: 15-performance-diagnostics-and-config-hook-follow-ups
plan: 06
subsystem: hook
tags: [shell, hook, async, diagnostics]
requires: []
provides:
  - Opt-in `pupdate init --mode async` shell snippets
  - Hidden hook command that can detach `run` execution in the background
  - Repo-local overlap protection and status visibility for active background hook runs
affects: [cmd-pupdate, docs]
tech-stack:
  added: []
  patterns: [opt-in-async-hook, detached-child-command, repo-local-hook-lock]
key-files:
  created:
    - .planning/phases/15-performance-diagnostics-and-config-hook-follow-ups/15-06-SUMMARY.md
    - cmd/pupdate/hook.go
    - cmd/pupdate/hook_test.go
  modified:
    - README.md
    - .planning/ROADMAP.md
    - .planning/STATE.md
    - cmd/pupdate/init.go
    - cmd/pupdate/init_shell.go
    - cmd/pupdate/init_snippets.go
    - cmd/pupdate/init_snippet_bash.go
    - cmd/pupdate/init_snippet_zsh.go
    - cmd/pupdate/init_snippet_fish.go
    - cmd/pupdate/init_test.go
    - cmd/pupdate/root.go
    - cmd/pupdate/status.go
    - cmd/pupdate/status_test.go
key-decisions:
  - "Make async hook execution an `init`-time mode switch so the default foreground contract stays unchanged and explicit opt-in remains easy to understand."
  - "Route hook execution through a hidden `pupdate hook` command so shell snippets stay simple while background launch, overlap control, and cleanup remain in Go code."
  - "Use a repo-local `.pupdate.hook.lock` with stale-lock expiry to skip overlapping background runs without introducing a daemon or broader process manager."
requirements-completed: []
duration: 21m
completed: 2026-04-30
---

# Phase 15 Plan 06: Async hook mode summary

Phase 15 plan 06 adds an explicit `pupdate init --mode async` opt-in that detaches hook-triggered runs into the background while leaving the existing foreground hook as the default, and it protects each repo from overlapping detached runs with a local lock file that also shows up in `pupdate status`.

## Verification

- `go test ./cmd/pupdate -count=1`
- `go test ./... -count=1`

## Files Created/Modified

- `cmd/pupdate/hook.go` - Adds the hidden hook command, detached-child launch path, and repo-local overlap lock handling.
- `cmd/pupdate/init*.go` - Extends shell snippet generation with the new `--mode` contract while keeping foreground mode as the default.
- `cmd/pupdate/status.go` - Reports the background hook lock path and active/idle state in diagnostics.
- `cmd/pupdate/*_test.go` - Covers init mode selection, overlap skipping, detached launch behavior, and status visibility.
- `README.md` - Documents the async init mode, overlap semantics, and new status output.
