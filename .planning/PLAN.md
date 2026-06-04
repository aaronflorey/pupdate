# Active TODO Plan

## TODO ID and Objective

- ID: `P2-R1`
- Objective: Review the shared preflight and status-guidance phase for behavior drift, scope creep, and missing regression coverage.

## Source Documents Consulted

- `.planning/TODO.md`
- `.planning/update.md`
- `.planning/TODO-TRACE.md`
- `.planning/todos/P2-R1.md`
- `.planning/todos/P2-T1.md`
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

- Active queue state selected from `.planning/TODO.md`: `Current TODO: none`, `Next recommended TODO: P2-R1`, so `P2-R1` is now the active review TODO.
- Fresh `git status --short` inspection produced no reported tracked worktree changes before this planning-state sync.
- `.planning/PRD.md` does not exist in this repository; the active PRD source referenced by the queue and TODO briefs is `.planning/update.md`.
- `.planning/IMPLEMENTATION-PLAN.md` does not exist in this repository; approved implementation context for this TODO comes from `.planning/quick/phase-15-performance-diagnostics-config-hook-planning.md` and the detailed briefs.
- Phase-2 implementation scope to review is the completed `P2-T1` shared preflight extraction and `P2-T2` status-remediation guidance work in `cmd/pupdate/`.
- The required acceptance work for this TODO is review and verification of existing phase-2 code plus planning-state updates only if the phase is accepted or a focused repair mapping is needed.

## Expected Starting Files

- `.planning/TODO.md`
- `.planning/TODO-TRACE.md`
- `.planning/todos/P2-R1.md`

## Allowed Discovery Scope

- Inspect the existing phase-2 implementation and tests in `cmd/pupdate/`, especially `preflight.go`, `run_execution.go`, `status.go`, and `status_test.go`, plus any directly relevant adjacent test files.
- Update planning-state files outside the expected starting list only when directly required to record a focused repair mapping or acceptance evidence for `P2-R1`.

## Outside-Expected-File Policy

- The expected files above are starting scope, not a hard allowlist.
- A file outside that list may change only if it is directly required to satisfy `P2-R1`, is consistent with `.planning/update.md`, this plan, and the TODO brief, and is reported explicitly with rationale under `Outside expected files`.
- Do not modify product code during this review TODO unless a focused repair loop is triggered by review or verification findings and remains within approved phase-2 scope.

## Required Implementation Scope

- Compare the delivered `P2-T1` and `P2-T2` work against their detailed briefs, the PRD requirements for maintainability and action-oriented status guidance, and the approved phase-planning note.
- Confirm whether the shared preflight refactor preserved `run` and `status` behavior boundaries and whether the status guidance explains existing states without inventing new decision branches.
- Accept the phase or identify concrete, narrowly scoped repair defects with enough detail for a repair pass.

## Explicit Non-Goals And Forbidden Changes

- Do not start monorepo config, install-docs, toolchain-pinning, or later-phase work.
- Do not broaden CLI surface, detection policy, manager resolution behavior, config schema, or hook behavior in this review task.
- Do not mark any TODO complete and do not commit.

## Acceptance Criteria

- The review compares phase-2 code and tests against `P2-T1`, `P2-T2`, and the relevant PRD requirements.
- The review identifies any behavior drift, scope creep, regressions, missing checks, or incomplete acceptance criteria.
- The review either accepts the phase or records focused repair work consistent with the approved scope.
- Any outside-expected-file change is necessary, justified, and reported.

## Required Checks

- Review `cmd/pupdate` diffs or current code ownership for shared-preflight logic and status-guidance scope.
- `go test ./cmd/pupdate -count=1`
- Manual inspection of the relevant `cmd/pupdate` tests and output assertions.

## Blocker Policy

- Stop only for unapproved product, public API or CLI, persistence or migration, security, dependency or tooling, architecture, external integration, or UX decisions not already approved by the PRD, plan, TODO brief, or current repository conventions.
- Continue through normal review and repair issues within scope, including stale expected-file lists, directly required adjacent planning or test updates, missing helper or test seam files, reuse of generated artifacts or equivalents, and failing checks caused by in-scope repair work.

## Generated Artifact Policy

- Reuse existing generated classes, types, files, fixtures, schemas, and other generated equivalents when available instead of creating duplicates.
- Prefer existing tests, helpers, and planning artifacts over parallel review-only structures when they already capture the needed evidence.

## Stop Conditions

- The review uncovers a required fix that would need an unapproved CLI, product, persistence, security, dependency, tooling, architecture, integration, or UX decision.
- The review cannot determine phase-2 correctness from the existing code, tests, and approved planning materials.
- The work would require changes outside approved phase-2 scope for reasons not directly required by `P2-R1`.
