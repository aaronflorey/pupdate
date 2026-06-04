# Active TODO Plan

## TODO ID and Objective

- ID: `P2-T1`
- Objective: Extract a shared preflight collection layer used by both `run` and `status` without changing their distinct behaviors.

## Source Documents Consulted

- `.planning/TODO.md`
- `.planning/update.md`
- `.planning/TODO-TRACE.md`
- `.planning/todos/P2-T1.md`
- `.planning/quick/phase-15-performance-diagnostics-config-hook-planning.md`
- `AGENTS.md`
- `go.mod`
- `mise.toml`
- `cmd/pupdate/run_execution.go`
- `cmd/pupdate/status.go`
- `cmd/pupdate/run_test.go`
- `cmd/pupdate/status_test.go`

## Current Repository State Summary

- Active queue state selected from `.planning/TODO.md`: `Current TODO: none`, `Next recommended TODO: P2-T1`, so `P2-T1` is now the active implementation TODO.
- Fresh `git status --short` inspection showed no tracked worktree changes before this planning-state sync.
- `.planning/PRD.md` does not exist in this repository; the active PRD source referenced by the queue and TODO brief is `.planning/update.md`.
- `.planning/IMPLEMENTATION-PLAN.md` does not exist in this repository; approved implementation context for this TODO comes from `.planning/quick/phase-15-performance-diagnostics-config-hook-planning.md` and the detailed brief.
- `cmd/pupdate/run_execution.go` and `cmd/pupdate/status.go` currently duplicate preflight logic for config loading, repo gating, detection, state loading, and freshness evaluation.
- Existing regression coverage already exercises several skip and status cases in `cmd/pupdate/run_test.go` and `cmd/pupdate/status_test.go`.

## Expected Starting Files

- `cmd/pupdate/run_execution.go`
- `cmd/pupdate/status.go`
- `cmd/pupdate/run_test.go`
- `cmd/pupdate/status_test.go`
- `cmd/pupdate/preflight.go`

## Allowed Discovery Scope

- Adjacent files under `cmd/pupdate/` may be added or updated only when directly required to complete the shared preflight refactor or its regression coverage.
- Reuse existing helpers, seams, and generated equivalents instead of duplicating logic in new files.

## Outside-Expected-File Policy

- The expected files above are starting scope, not a hard allowlist.
- A file outside that list may change only if it is directly required to satisfy `P2-T1`, is consistent with `.planning/update.md`, this plan, and the TODO brief, and is reported explicitly with rationale under `Outside expected files`.
- Unrelated cleanup, renames, API changes, config-schema changes, and broader refactors are forbidden.

## Required Implementation Scope

- Define a shared preflight collector or normalized result structure for the setup and decision flow used by both `run` and `status`.
- Refactor `run` and `status` to consume that shared path while preserving their current output, branching, and side-effect differences.
- Add or update regression coverage proving unchanged behavior for home-directory checks, configured-root exclusions, `.pupignore`, detection, state loading, and freshness evaluation.

## Explicit Non-Goals And Forbidden Changes

- Do not add the later status-remediation UX planned for `P2-T2`.
- Do not change detection depth, config schema, manager resolution behavior, or hook behavior.
- Do not move the refactor outside `cmd/pupdate` unless an approved blocker decision says otherwise.
- Do not mark any TODO complete and do not commit.

## Acceptance Criteria

- `run` and `status` both use a common preflight collection layer.
- Existing skip/readiness semantics remain unchanged apart from internal refactoring.
- Regression tests cover the shared preflight path for both commands.
- Any outside-expected-file change is necessary, justified, and reported.

## Required Checks

- `go test ./cmd/pupdate -count=1`
- Manual inspection that the shared helper now owns the duplicated setup logic while `run` and `status` keep distinct behavior only where intended.

## Blocker Policy

- Stop only for unapproved product, public API or CLI, persistence or migration, security, dependency or tooling, architecture, external integration, or UX decisions not already approved by the PRD, plan, TODO brief, or current repository conventions.
- Continue through normal implementation issues within scope, including stale expected-file lists, directly required adjacent-file edits, missing helper or test seam files, reuse of generated artifacts or equivalents, and failing checks caused by the active implementation.

## Generated Artifact Policy

- Reuse existing generated classes, types, files, fixtures, schemas, and other generated equivalents when available instead of creating duplicates.
- Do not introduce duplicate helpers or parallel preflight result types when an existing local pattern can be extended cleanly.

## Stop Conditions

- The refactor would require moving logic into packages outside `cmd/pupdate`.
- Existing tests reveal a behavior difference that requires a product decision rather than a refactor fix.
- The work would require modifying unrelated command surfaces beyond what is directly required for `P2-T1`.
