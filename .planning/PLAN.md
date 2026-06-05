# Active TODO Plan

## TODO ID and Objective

- ID: `P3-R1`
- Objective: Review the monorepo-detection phase against the latency, exclusion, and low-surprise constraints.

## Source Documents Consulted

- `.planning/TODO.md`
- `.planning/update.md`
- `.planning/TODO-TRACE.md`
- `.planning/todos/P3-R1.md`
- `.planning/todos/P3-T1.md`
- `.planning/todos/P3-T2.md`
- `.planning/todos/P3-T3.md`
- `.planning/todos/P3-T4.md`
- `.planning/todos/P3-T5.md`
- `.planning/todos/P3-T6.md`
- `.planning/quick/phase-15-performance-diagnostics-config-hook-planning.md`
- `AGENTS.md`
- `README.md`
- `cmd/pupdate/config.go`
- `cmd/pupdate/config_cmd.go`
- `cmd/pupdate/config_cmd_test.go`
- `cmd/pupdate/config_test.go`
- `cmd/pupdate/run_execution.go`
- `cmd/pupdate/status.go`
- `internal/detection/detector.go`
- `internal/detection/detector_benchmark_test.go`
- `internal/detection/detector_test.go`
- `go.mod`
- `mise.toml`

## Current Repository State Summary

- Active queue state selected from `.planning/TODO.md`: `Current TODO: P3-R1`, `Next recommended TODO: P3-R1`, so `P3-R1` is the active TODO for this run.
- `.planning/PRD.md` does not exist in this repository; the PRD source of truth for this queue is `.planning/update.md`.
- `.planning/IMPLEMENTATION-PLAN.md` does not exist in this repository; approved implementation context comes from `.planning/quick/phase-15-performance-diagnostics-config-hook-planning.md`, the detailed TODO briefs, and `.planning/TODO-TRACE.md`.
- Fresh `git status --short` evidence shows planning-only worktree modifications in `.planning/PLAN.md`, `.planning/TODO.md`, `.planning/todos/P3-R1.md`, `.planning/todos/P3-T1.md`, and `.planning/todos/P3-T4.md`.
- The current implementation already exposes `workspace_globs` and `folder_blacklist` in config parsing and `pupdate config`, and detection threads both settings into the shared run/status flows.
- `normalizeWorkspaceGlob` enforces repo-relative glob inputs and rejects absolute paths, root matches, and parent traversal.
- `normalizeFolderBlacklistEntry` enforces exact directory-name entries and rejects globs, `.`/`..`, and path separators.
- Detection still starts from the historical shallow default (`.`, depth-1 directories, direct `packages/*` children) and only expands beyond that when configured `workspace_globs` are present.
- Directory blacklist checks are applied before entering shallow candidates, `packages/*` children, and workspace-glob traversal segments, using exact last-path-segment matching.
- The README currently documents the opt-in `workspace_globs` behavior, the exact-match `folder_blacklist` behavior, and the distinction from `.pupignore`.
- Planning-state status alignment confirmed: `.planning/TODO.md`, `.planning/todos/P3-T1.md`, and `.planning/todos/P3-T4.md` now all show `P3-T1` and `P3-T4` as `complete`.

## Expected Starting Files

- `.planning/TODO.md`
- `.planning/PLAN.md`
- `.planning/TODO-TRACE.md`
- `.planning/todos/P3-R1.md`

## Allowed Discovery Scope

- Inspect adjacent config, detection, run/status, benchmark, and README files to collect evidence for the phase review.
- Inspect tests and planning briefs for `P3-T1` through `P3-T6` to verify the phase against the approved plan.
- Update planning-state files only for active-TODO synchronization, reviewer notes, blocker recording, or completion state after approval.

## Outside-Expected-File Policy

- The expected files above are expected starting files, not a hard allowlist.
- A file outside that list may change only if it is directly required to satisfy `P3-R1`, is consistent with `.planning/update.md`, this plan, and the detailed brief, and is reported explicitly with rationale under `Outside expected files`.
- For this review TODO, product-code changes are not expected. If review or verification finds concrete defects, repair must be delegated as a separate executor pass that changes only the directly required files.

## Required Implementation Scope

- Compare the phase-3 config, detection, and README work delivered by `P3-T1` through `P3-T6` against the PRD, inserted user idea, and approved plan.
- Confirm `workspace_globs` remains opt-in and that `folder_blacklist` entries behave as exact directory-name matches.
- Confirm the default scan path remains unchanged except for explicitly blacklisted directories.
- Confirm README wording matches implemented behavior for `workspace_globs`, `folder_blacklist`, and `.pupignore`.
- Accept the phase or route concrete defects into a narrow repair loop.

## Explicit Non-Goals And Forbidden Changes

- Do not implement new monorepo scanning or blacklist behavior as part of the review itself.
- Do not start phase-4 install, toolchain, or integration-test work.
- Do not broaden `folder_blacklist` into path-pattern, glob, or regex semantics.
- Do not mark the TODO complete and do not commit until review, verification, audit, and explicit user approval are all complete.

## Acceptance Criteria

- The reviewer compares code against all phase implementation task briefs (`P3-T1` through `P3-T6`).
- The reviewer checks the relevant PRD, inserted-idea, and approved-plan requirements.
- The reviewer identifies scope drift, missing tests, regressions, and incomplete acceptance criteria if any exist.
- The phase is either accepted with evidence or sent through focused repair for concrete issues.

## Required Checks

- Review config, detection, and README diffs and current implementations against the phase briefs.
- `go test ./cmd/pupdate ./internal/detection -count=1`
- `go test ./internal/detection -run '^$' -bench 'BenchmarkDetectProjectTree' -count=1`
- Manual inspection of README config examples for `workspace_globs`, `folder_blacklist`, and `.pupignore` distinctions.

## Correctness Evidence Plan

- Acceptance criterion 1: Map each `P3-T1` through `P3-T6` objective and acceptance criterion to current code/tests in config, detection, and README files.
- Acceptance criterion 2: Validate PRD alignment by checking that broader scanning is opt-in, exact-name blacklist semantics are preserved, and the shallow default remains the fast path.
- Acceptance criterion 3: Look for regressions or drift through targeted test results, benchmark output, and manual README/code comparison.
- Acceptance criterion 4: Require both independent reviewer and verifier reports to return substantive `PASS` with concrete evidence, or send only the identified defects into repair.

## Risk Areas And Invariants

- The default detection path must remain shallow when `workspace_globs` is unset.
- Blacklist behavior must stay exact-name based and must not silently become path-pattern matching.
- `.pupignore` must remain a repo-wide short-circuit distinct from traversal-only blacklist behavior.
- `run` and `status` must stay aligned because both depend on the shared preflight/detection flow.
- Any benchmark signal suggesting default-path regression is a stop condition, not a silent acceptance.

## Blocker Policy

- Stop only for unapproved product, public API or CLI, persistence or migration, security, dependency or tooling, architecture, external integration, or UX decisions not already approved by the PRD, plan, TODO brief, or repository conventions.
- Continue through normal review and repair issues within scope, including stale expected-file lists, adjacent evidence gathering, stale planning-state notes, directly required test updates, or failing checks caused by the active implementation.

## Generated Artifact Policy

- Reuse existing task briefs, tests, benchmark coverage, and README structure as the evidence source instead of creating duplicate planning or verification artifacts unless a focused repair requires it.

## Stop Conditions

- Benchmark or test results suggest the default path regressed materially.
- Folder blacklist semantics appear to require a broader model than exact directory-name entries.
- Required repair work would change files outside the phase-3 task scope in a way that alters the approved contract.
- A required source file is missing and cannot be reconstructed safely from the repository.
