# Active TODO Plan

## TODO ID and Objective

- ID: `P3-T1`
- Objective: Define validated `workspace_globs` config support and surface it through `pupdate config`.

## Source Documents Consulted

- `.planning/TODO.md`
- `.planning/update.md`
- `.planning/TODO-TRACE.md`
- `.planning/todos/P3-T1.md`
- `.planning/quick/phase-15-performance-diagnostics-config-hook-planning.md`
- `AGENTS.md`
- `go.mod`
- `mise.toml`
- `cmd/pupdate/config.go`
- `cmd/pupdate/config_test.go`
- `cmd/pupdate/config_cmd.go`
- `cmd/pupdate/config_cmd_test.go`

## Current Repository State Summary

- Active queue state selected from `.planning/TODO.md`: `Current TODO: none`, `Next recommended TODO: P3-T1`, so `P3-T1` is now the active implementation TODO.
- Fresh `git status --short` inspection produced no reported tracked worktree entries before this planning-state sync.
- `.planning/PRD.md` does not exist in this repository; the active PRD source referenced by the queue and TODO briefs is `.planning/update.md`.
- `.planning/IMPLEMENTATION-PLAN.md` does not exist in this repository; approved implementation context for this TODO comes from `.planning/quick/phase-15-performance-diagnostics-config-hook-planning.md` and the detailed brief.
- Existing config support currently includes `root_directories`, `quiet`, and `allow_scripts` in `cmd/pupdate/config.go`, with `pupdate config` printing raw and resolved config values.
- `P3-T1` is limited to defining and validating the config contract for `workspace_globs`; detection traversal and README changes remain later TODOs.

## Expected Starting Files

- `cmd/pupdate/config.go`
- `cmd/pupdate/config_test.go`
- `cmd/pupdate/config_cmd.go`
- `cmd/pupdate/config_cmd_test.go`
- `.planning/TODO.md`
- `.planning/todos/P3-T1.md`
- `.planning/PLAN.md`

## Allowed Discovery Scope

- Inspect adjacent config consumers in `cmd/pupdate/` when needed to preserve current config-loading and output patterns.
- Update directly related test helpers or adjacent command tests only when required to satisfy `P3-T1` acceptance criteria.
- Update planning-state files outside the expected starting list only when directly required to record a blocker or accepted scope-consistent repair mapping for `P3-T1`.

## Outside-Expected-File Policy

- The expected files above are expected starting files, not a hard allowlist.
- A file outside that list may change only if it is directly required to satisfy `P3-T1`, is consistent with `.planning/update.md`, this plan, and the detailed brief, and is reported explicitly with rationale under `Outside expected files`.
- No detection behavior files should change in this TODO unless a directly required config-surface seam already approved by the brief must be touched; if that becomes necessary, report it explicitly.

## Required Implementation Scope

- Add `workspace_globs` to the user config model with deterministic default behavior when the field is missing.
- Validate and normalize configured `workspace_globs` values so later detection logic receives a safe input shape.
- Expose the configured and resolved `workspace_globs` values through `pupdate config` output and tests.

## Explicit Non-Goals And Forbidden Changes

- Do not change detection traversal behavior in this task.
- Do not document `workspace_globs` in the README yet.
- Do not add folder blacklist support in this task.
- Do not mark any TODO complete and do not commit.

## Acceptance Criteria

- Missing config keeps the current shallow scan default with no implicit globs.
- Invalid `workspace_globs` values are handled consistently and are covered by tests.
- `pupdate config` reports configured and resolved `workspace_globs` values in a deterministic format.
- Any outside-expected-file change is necessary, justified, and reported.

## Required Checks

- `go test ./cmd/pupdate -count=1`
- Manual inspection of config parsing behavior and config-command output assertions.

## Blocker Policy

- Stop only for unapproved product, public API or CLI, persistence or migration, security, dependency or tooling, architecture, external integration, or UX decisions not already approved by the PRD, plan, TODO brief, or current repository conventions.
- Continue through normal implementation and repair issues within scope, including stale expected-file lists, directly required adjacent wiring, missing helper or test seam files, reuse of generated artifacts or equivalents, and failing checks caused by the in-scope implementation.

## Generated Artifact Policy

- Reuse existing generated classes, types, files, fixtures, schemas, and other generated equivalents when available instead of creating duplicates.
- Prefer existing config helpers and test utilities over introducing parallel structures when they already satisfy the TODO.

## Stop Conditions

- The chosen glob validation or normalization shape would require an unapproved CLI, product, persistence, security, dependency, tooling, architecture, integration, or UX decision.
- The work would require changing detection files before the config contract is settled.
- A required source file is missing and cannot be reconstructed safely from the repository.
