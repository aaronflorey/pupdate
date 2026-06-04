---
status: complete
trigger: "/gsd:quick create planning phases for the codebase improvement items: 1) fix Cargo.lock case handling, 2) make cmd/pupdate tests hermetic and stop ambient config coupling, 3) add timeout/injection for git submodule freshness checks, 4) reduce hot-path lockfile hashing cost, 5) harden state-file persistence with parent-dir fsync, 6) remove auto-create-on-run behavior for user config and treat missing config as defaults"
---

# Quick Task

## Current Focus

### hypothesis
The six requested reliability and maintenance items can be staged as one post-v1 maintenance phase with one plan per item, because they touch distinct detection, config, freshness, hashing, and persistence code paths.

### next_action
Done.

## Evidence

- timestamp: 2026-04-29T00:00:00Z
  note: `internal/detection/matrix.go` already treats canonical Rust lockfiles as `Cargo.lock`, while detector tests still exercise lowercase `cargo.lock`, confirming a targeted case-handling follow-up is warranted.
- timestamp: 2026-04-29T00:05:00Z
  note: `cmd/pupdate/run_test.go` currently asserts `run` creates a default config file under `XDG_CONFIG_HOME`, which couples tests and runtime behavior to ambient user-config persistence.
- timestamp: 2026-04-29T00:10:00Z
  note: `internal/freshness/engine.go` shells out to `git submodule status --recursive` directly with `exec.Command` and no timeout or injectable runner.
- timestamp: 2026-04-29T00:12:00Z
  note: `internal/freshness/engine.go` computes full SHA-256 hashes for matched lockfiles on every freshness pass, making lockfile hashing a hot-path optimization candidate.
- timestamp: 2026-04-29T00:14:00Z
  note: `internal/state/store.go` syncs the temp file before rename but does not fsync the parent directory after replacing `.pupdate`.
- timestamp: 2026-04-29T00:18:00Z
  note: Added Phase 09 planning artifacts, roadmap entry, and state metadata for a six-plan post-v1 hardening phase.

## Resolution

### root_cause
The repo had no follow-up maintenance phase capturing the newly identified reliability, performance, and hermeticity improvements, so the work was not yet represented in the planning system.

### fix
Created a new post-v1 hardening phase with one plan per requested improvement item, plus validation and state/roadmap updates so the work is discoverable from the standard planning entry points.
