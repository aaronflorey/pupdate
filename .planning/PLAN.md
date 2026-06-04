# Active TODO Plan

## TODO ID and Objective

- ID: `P3-T3`
- Objective: Document `workspace_globs` behavior, defaults, and examples in the README.

## Source Documents Consulted

- `.planning/TODO.md`
- `.planning/update.md`
- `.planning/TODO-TRACE.md`
- `.planning/todos/P3-T3.md`
- `.planning/quick/phase-15-performance-diagnostics-config-hook-planning.md`
- `AGENTS.md`
- `README.md`
- `cmd/pupdate/config.go`
- `cmd/pupdate/config_cmd.go`

## Current Repository State Summary

- Active queue state selected from `.planning/TODO.md`: `Current TODO: none`, `Next recommended TODO: P3-T3`, so `P3-T3` is the active implementation TODO for this run.
- Fresh git-status inspection found no substantive pre-existing worktree evidence beyond the planning-state files being refreshed for this run.
- `.planning/PRD.md` does not exist in this repository; the active PRD source for this queue is `.planning/update.md`.
- `.planning/IMPLEMENTATION-PLAN.md` does not exist in this repository; approved implementation context comes from `.planning/quick/phase-15-performance-diagnostics-config-hook-planning.md`, `.planning/TODO-TRACE.md`, and the detailed brief.
- Runtime support for `workspace_globs` already exists in config parsing and detection; this TODO is documentation-only and should align README text with the implemented opt-in shallow-scan expansion behavior.
- The README currently describes only the default shallow traversal and does not yet document the `workspace_globs` config key or example monorepo layouts.

## Expected Starting Files

- `README.md`

## Allowed Discovery Scope

- Inspect adjacent config or command files only to confirm existing `workspace_globs` behavior and naming while keeping implementation changes scoped to README documentation.
- Update planning-state files only for active-TODO synchronization, blocker recording, or completion state after review, verification, and approval.

## Outside-Expected-File Policy

- The expected files above are expected starting files, not a hard allowlist.
- A file outside that list may change only if it is directly required to satisfy `P3-T3`, is consistent with `.planning/update.md`, this plan, and the detailed brief, and is reported explicitly with rationale under `Outside expected files`.
- For this TODO, outside-list product-code changes are not expected and would require a real stop-condition review.

## Required Implementation Scope

- Add `workspace_globs` to the README config documentation.
- Explain that the default detection remains shallow unless users opt in.
- Provide one or more concrete examples such as `apps/*` or `services/*` for common monorepo layouts.
- Keep the docs aligned with the existing command/config names and current traversal contract.

## Explicit Non-Goals And Forbidden Changes

- Do not change runtime behavior, config parsing, or tests in this task.
- Do not document unrelated install, platform, or blacklist topics beyond what is needed to explain `workspace_globs` accurately.
- Do not mark any TODO complete and do not commit.

## Acceptance Criteria

- README documents the `workspace_globs` setting and its default behavior.
- README examples demonstrate opt-in monorepo expansion concretely.
- README still describes the shallow default traversal accurately.
- Any outside-expected-file change is necessary, justified, and reported.

## Required Checks

- Manual inspection of README config and behavior sections for accuracy against existing implementation.

## Blocker Policy

- Stop only for unapproved product, public API or CLI, persistence or migration, security, dependency or tooling, architecture, external integration, or UX decisions not already approved by the PRD, plan, TODO brief, or repository conventions.
- Continue through normal implementation and repair issues within scope, including stale expected-file lists, directly required adjacent-file confirmation, and documentation wording fixes needed to match the implemented feature.

## Generated Artifact Policy

- Reuse existing command names, config keys, examples, and established README structure instead of inventing parallel documentation patterns.

## Stop Conditions

- Documenting `workspace_globs` would require describing behavior that is not already implemented and tested.
- The work would require editing files outside the allowed list in a way that changes runtime behavior or broadens scope.
- A required source file is missing and cannot be reconstructed safely from the repository.
