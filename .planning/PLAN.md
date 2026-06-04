# Active TODO Plan

## Completion Status

- Status: `P3-T2` implementation complete, reviewed PASS, verified PASS.
- User approval received to mark the TODO complete in planning state and prepare the repo for commit without creating the commit yet.

## TODO ID and Objective

- ID: `P3-T2`
- Objective: Apply `workspace_globs` during detection while keeping the current shallow default scan unchanged.

## Source Documents Consulted

- `.planning/TODO.md`
- `.planning/update.md`
- `.planning/TODO-TRACE.md`
- `.planning/todos/P3-T2.md`
- `.planning/quick/phase-15-performance-diagnostics-config-hook-planning.md`
- `AGENTS.md`
- `go.mod`
- `mise.toml`
- `cmd/pupdate/config.go`
- `cmd/pupdate/preflight.go`
- `cmd/pupdate/run.go`
- `cmd/pupdate/run_execution.go`
- `cmd/pupdate/status.go`
- `cmd/pupdate/status_test.go`
- `internal/detection/detector.go`
- `internal/detection/detector_test.go`
- `internal/detection/detector_benchmark_test.go`

## Current Repository State Summary

- Active queue state selected from `.planning/TODO.md`: `Current TODO: none`, `Next recommended TODO: P3-T2`, so `P3-T2` is now the active implementation TODO.
- Fresh `git status --short` inspection did not report any pre-existing tracked or staged worktree entries before this planning-state sync.
- `.planning/PRD.md` does not exist in this repository; the active PRD source referenced by the queue and TODO briefs is `.planning/update.md`.
- `.planning/IMPLEMENTATION-PLAN.md` does not exist in this repository; approved implementation context for this TODO comes from `.planning/quick/phase-15-performance-diagnostics-config-hook-planning.md` and the detailed brief.
- `userConfig` already carries normalized `workspace_globs` values, but `collectPreflight` still calls `detectFn(".")`, and `detectFn` still points to `detection.Detect(dir string)`.
- Default detection currently scans `.`, depth-1 directories, and `packages/*`, with `.gitignore` filtering and a benchmark guardrail already covering that shallow path.
- Existing command tests stub `detectFn` with a single `dir string` argument, so any plumbing change may require directly related command-test updates in addition to detector tests.

## Expected Starting Files

- `internal/detection/detector.go`
- `internal/detection/detector_test.go`
- `internal/detection/detector_benchmark_test.go`
- `cmd/pupdate/preflight.go`
- `cmd/pupdate/run.go`
- `cmd/pupdate/status.go`
- `cmd/pupdate/status_test.go`
- `.planning/TODO.md`
- `.planning/todos/P3-T2.md`
- `.planning/PLAN.md`

## Allowed Discovery Scope

- Inspect adjacent detection helpers under `internal/detection/` when needed to preserve current ordering, ignore handling, and matching semantics.
- Update directly related command tests or small seams in `cmd/pupdate/` when required to thread resolved `workspace_globs` through preflight without changing unrelated CLI behavior.
- Update planning-state files outside the expected starting list only when directly required to record a blocker or a scope-consistent repair note for `P3-T2`.

## Outside-Expected-File Policy

- The expected files above are expected starting files, not a hard allowlist.
- A file outside that list may change only if it is directly required to satisfy `P3-T2`, is consistent with `.planning/update.md`, this plan, and the detailed brief, and is reported explicitly with rationale under `Outside expected files`.
- Outside-list edits are acceptable for directly required adjacent wiring or tests, but not for docs, unrelated refactors, or later TODO work.

## Required Implementation Scope

- Extend the detection call path so resolved `workspace_globs` reach the detector from shared preflight used by both `run` and `status`.
- Expand candidate scanning only for directories matched by configured workspace globs while preserving the current default scan exactly when `workspace_globs` is unset.
- Keep existing detection semantics for matched roots, including ordering, `.gitignore` filtering, matched-file reporting, and ecosystem/state-key behavior.
- Add focused tests covering unchanged default behavior and opt-in discovery for common layouts like `apps/*` or `services/*`.
- Inspect or extend the benchmark guardrail only as needed to keep the default-path performance contract visible.

## Explicit Non-Goals And Forbidden Changes

- Do not change the default traversal depth for users who do not configure `workspace_globs`.
- Do not implement folder blacklist behavior in this TODO.
- Do not update README or other user-facing docs in this TODO.
- Do not broaden scanning into unbounded recursion or a new detection model beyond the approved glob expansion.
- Do not mark any TODO complete and do not commit.

## Acceptance Criteria

- Repositories without `workspace_globs` behave exactly as they do today.
- Configured glob patterns can include additional monorepo roots such as `apps/*` or `services/*` without broad recursive scanning elsewhere.
- Detection tests cover both the unchanged default path and the configured opt-in path.
- Any outside-expected-file change is necessary, justified, and reported.

## Required Checks

- `go test ./internal/detection ./cmd/pupdate -count=1`
- `go test ./internal/detection -run '^$' -bench 'BenchmarkDetectProjectTree' -count=1`
- Manual inspection that default shallow-scan behavior remains unchanged when `workspace_globs` is unset.

## Blocker Policy

- Stop only for unapproved product, public API or CLI, persistence or migration, security, dependency or tooling, architecture, external integration, or UX decisions not already approved by the PRD, plan, TODO brief, or current repository conventions.
- Continue through normal implementation and repair issues within scope, including stale expected-file lists, directly required adjacent wiring, missing helper or test seam files, reuse of generated artifacts or equivalents, and failing checks caused by the in-scope implementation.

## Generated Artifact Policy

- Reuse existing generated classes, types, files, fixtures, schemas, and other generated equivalents when available instead of creating duplicates.
- Prefer existing detection helpers, config normalization, benchmark fixtures, and command test seams over introducing parallel structures when they already satisfy the TODO.

## Stop Conditions

- The implementation would require recursive scanning beyond the approved glob model.
- Benchmark results indicate a default-path regression that needs a new design decision.
- A required source file is missing and cannot be reconstructed safely from the repository.
