# Active TODO Plan

## TODO ID and Objective

- ID: `P3-T6`
- Objective: Document the folder blacklist behavior, exact-match semantics, and examples in the README.

## Source Documents Consulted

- `.planning/TODO.md`
- `.planning/update.md`
- `.planning/TODO-TRACE.md`
- `.planning/todos/P3-T6.md`
- `.planning/quick/phase-15-performance-diagnostics-config-hook-planning.md`
- `AGENTS.md`
- `README.md`
- `cmd/pupdate/config.go`
- `cmd/pupdate/config_cmd.go`
- `internal/detection/detector_test.go`

## Current Repository State Summary

- Active queue state selected from `.planning/TODO.md`: `Current TODO: none`, `Next recommended TODO: P3-T6`, so `P3-T6` is the active implementation TODO for this run.
- `.planning/PRD.md` does not exist in this repository; the active PRD source for this queue is `.planning/update.md`.
- `.planning/IMPLEMENTATION-PLAN.md` does not exist in this repository; approved implementation context comes from `.planning/quick/phase-15-performance-diagnostics-config-hook-planning.md`, `.planning/TODO-TRACE.md`, and the detailed brief.
- Fresh repository inspection was limited to the planning-state files being refreshed for this run plus documentation and config/detection evidence for the already-implemented `folder_blacklist` feature.
- Runtime support for `folder_blacklist` already exists in config parsing, `pupdate config` output, and detection behavior. Validation enforces exact directory-name entries only: no globs and no path separators.
- Detection tests cover the intended examples: a blacklist entry like `blah` suppresses traversal into both a top-level `./blah` directory and nested paths such as `./foo/blah/...` when matched during default or workspace-glob-assisted traversal.
- The README currently documents `workspace_globs` and `.pupignore`, but it does not yet document `folder_blacklist`, its exact-match semantics, or how it differs from `.pupignore`.

## Expected Starting Files

- `README.md`

## Allowed Discovery Scope

- Inspect adjacent config, command, and detection files only to confirm the implemented `folder_blacklist` naming and semantics while keeping implementation changes scoped to README documentation.
- Update planning-state files only for active-TODO synchronization, blocker recording, or completion state after review, verification, and approval.

## Outside-Expected-File Policy

- The expected files above are expected starting files, not a hard allowlist.
- A file outside that list may change only if it is directly required to satisfy `P3-T6`, is consistent with `.planning/update.md`, this plan, and the detailed brief, and is reported explicitly with rationale under `Outside expected files`.
- For this TODO, outside-list product-code changes are not expected and would require a real stop-condition review.

## Required Implementation Scope

- Add `folder_blacklist` to the README config documentation.
- Explain that each entry is an exact directory-name match, not a glob and not a path pattern.
- Provide concrete examples showing that `blah` excludes both `./blah` and `./foo/blah`.
- Clarify how `folder_blacklist` interacts with the default shallow scan and optional `workspace_globs` expansions.
- Distinguish traversal skipping via `folder_blacklist` from repo-wide suppression via `.pupignore`.

## Explicit Non-Goals And Forbidden Changes

- Do not change runtime behavior, config parsing, or tests in this task.
- Do not broaden the docs into unrelated install, platform, or future blacklist-pattern topics.
- Do not document unimplemented glob or path-pattern blacklist behavior.
- Do not mark any TODO complete and do not commit.

## Acceptance Criteria

- README documents the `folder_blacklist` setting and its exact directory-name semantics.
- README examples demonstrate that `blah` excludes `./blah` and `./foo/blah` concretely.
- README explains the difference between `folder_blacklist` traversal skips and `.pupignore` repo-wide skipping.
- README explains the interaction with the default shallow scan and configured `workspace_globs` expansion without implying broader pattern support.
- Any outside-expected-file change is necessary, justified, and reported.

## Required Checks

- Manual inspection of README config docs and examples for accuracy against existing config validation, config output, and detection behavior.

## Blocker Policy

- Stop only for unapproved product, public API or CLI, persistence or migration, security, dependency or tooling, architecture, external integration, or UX decisions not already approved by the PRD, plan, TODO brief, or repository conventions.
- Continue through normal implementation and repair issues within scope, including stale expected-file lists, directly required adjacent-file confirmation, and documentation wording fixes needed to match the implemented feature.

## Generated Artifact Policy

- Reuse existing command names, config keys, examples, and established README structure instead of inventing parallel documentation patterns.

## Stop Conditions

- Documenting `folder_blacklist` would require describing behavior that is not already implemented and tested.
- The work would require editing files outside the allowed list in a way that changes runtime behavior or broadens scope.
- A required source file is missing and cannot be reconstructed safely from the repository.
