# Active TODO Plan

## TODO ID and Objective

- ID: `P2-T2`
- Objective: Add action-oriented remediation guidance to `pupdate status` for the blocked states already detected by the shared preflight flow.

## Source Documents Consulted

- `.planning/TODO.md`
- `.planning/update.md`
- `.planning/TODO-TRACE.md`
- `.planning/todos/P2-T2.md`
- `.planning/quick/phase-15-performance-diagnostics-config-hook-planning.md`
- `AGENTS.md`
- `go.mod`
- `mise.toml`
- `cmd/pupdate/preflight.go`
- `cmd/pupdate/run_execution.go`
- `cmd/pupdate/status.go`
- `cmd/pupdate/status_test.go`

## Current Repository State Summary

- Active queue state selected from `.planning/TODO.md`: `Current TODO: none`, `Next recommended TODO: P2-T2`, so `P2-T2` is now the active implementation TODO.
- Fresh `git status --short` inspection showed no tracked worktree changes before this planning-state sync.
- `.planning/PRD.md` does not exist in this repository; the active PRD source referenced by the queue and TODO brief is `.planning/update.md`.
- `.planning/IMPLEMENTATION-PLAN.md` does not exist in this repository; approved implementation context for this TODO comes from `.planning/quick/phase-15-performance-diagnostics-config-hook-planning.md` and the detailed brief.
- `cmd/pupdate/status.go` already exposes skip reasons, install readiness, manager-path discovery, and hook/config/state facts, but it does not yet emit targeted remediation guidance lines.
- `cmd/pupdate/preflight.go` centralizes the skip-state reasons that the new status guidance should reuse instead of re-deriving policy separately.
- Existing regression coverage in `cmd/pupdate/status_test.go` covers ready, skipped, blocked-manager, config-error, and hook-lock cases that can anchor deterministic guidance assertions.

## Expected Starting Files

- `cmd/pupdate/status.go`
- `cmd/pupdate/status_test.go`

## Allowed Discovery Scope

- Adjacent files under `cmd/pupdate/` may be added or updated only when directly required to express the approved status guidance or keep its tests deterministic.
- Reuse the shared preflight data already available in `cmd/pupdate/preflight.go` and existing status-formatting helpers instead of duplicating new decision plumbing.

## Outside-Expected-File Policy

- The expected files above are starting scope, not a hard allowlist.
- A file outside that list may change only if it is directly required to satisfy `P2-T2`, is consistent with `.planning/update.md`, this plan, and the TODO brief, and is reported explicitly with rationale under `Outside expected files`.
- Unrelated cleanup, renames, run-path behavior changes, config-schema changes, and broader refactors are forbidden.

## Required Implementation Scope

- Identify the currently blocked or confusing `status` states that should emit concrete remediation guidance based on the already-collected preflight and target data.
- Add concise, deterministic guidance text to `pupdate status` without changing the existing reported facts or the `run` command's behavior.
- Add or update tests in `cmd/pupdate/status_test.go` that pin representative guidance output for the supported states covered by the brief and PRD.

## Explicit Non-Goals And Forbidden Changes

- Do not add a new `doctor` command or any other CLI surface.
- Do not modify `run`, underlying decision logic, detection depth, config schema, manager resolution behavior, or hook behavior in this task.
- Do not broaden support policy, invent a new config model, or commit to unapproved UX beyond deterministic remediation text derived from existing status facts.
- Do not mark any TODO complete and do not commit.

## Acceptance Criteria

- `pupdate status` includes concrete next-step guidance for the supported blocked states covered by the PRD and TODO brief.
- The guidance text is deterministic and covered by tests.
- Existing status fields remain readable and do not change meaning.
- Any outside-expected-file change is necessary, justified, and reported.

## Required Checks

- `go test ./cmd/pupdate -count=1`
- Manual inspection of `status` output assertions in `cmd/pupdate/status_test.go`

## Blocker Policy

- Stop only for unapproved product, public API or CLI, persistence or migration, security, dependency or tooling, architecture, external integration, or UX decisions not already approved by the PRD, plan, TODO brief, or current repository conventions.
- Continue through normal implementation issues within scope, including stale expected-file lists, directly required adjacent-file edits, missing helper or test seam files, reuse of generated artifacts or equivalents, and failing checks caused by the active implementation.

## Generated Artifact Policy

- Reuse existing generated classes, types, files, fixtures, schemas, and other generated equivalents when available instead of creating duplicates.
- Prefer extending existing status snapshot and formatting structures over adding parallel status-reporting layers when the current data can support the approved guidance cleanly.

## Stop Conditions

- A proposed guidance string would commit the product to a new support policy or config model.
- The desired guidance cannot be expressed from the data already produced by the shared preflight layer and current status-target data.
- The work would require modifying files outside the allowed scope for reasons not directly required by `P2-T2`.
