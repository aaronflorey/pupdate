# Active TODO Plan

## TODO ID and Objective

- ID: `P3-T5`
- Objective: Apply the folder blacklist across all detection paths, including the default shallow scan and configured workspace expansions.

## Source Documents Consulted

- `.planning/TODO.md`
- `.planning/update.md`
- `.planning/TODO-TRACE.md`
- `.planning/todos/P3-T5.md`
- `.planning/quick/phase-15-performance-diagnostics-config-hook-planning.md`
- `AGENTS.md`
- `go.mod`
- `mise.toml`
- `internal/detection/detector.go`
- `cmd/pupdate/run_execution.go`
- `cmd/pupdate/status.go`

## Current Repository State Summary

- Active queue state selected from `.planning/TODO.md`: `Current TODO: none`, `Next recommended TODO: P3-T5`, so `P3-T5` is the active implementation TODO for this run.
- Fresh `git status --short` inspection produced no pre-existing tracked or staged worktree changes before this planning-state sync.
- `.planning/PRD.md` does not exist in this repository; the active PRD source referenced by the queue is `.planning/update.md`.
- `.planning/IMPLEMENTATION-PLAN.md` does not exist in this repository; the approved implementation context for this TODO comes from `.planning/quick/phase-15-performance-diagnostics-config-hook-planning.md`, `.planning/TODO-TRACE.md`, and the detailed brief.
- `internal/detection/detector.go` already supports `workspace_globs` via `DetectWithWorkspaceGlobs`, but there is no folder-blacklist input in the detector API or scan helpers yet.
- The current default detection path scans `.`, first-level child directories, and direct children of `packages/`, while `workspace_globs` expansion adds more directories through `scanWorkspaceGlob`.
- `run` and `status` both consume detection through shared preflight, so blacklist plumbing may require touching that shared seam or existing config plumbing even though the brief only lists `run_execution.go` and `status.go` as starting files.

## Expected Starting Files

- `internal/detection/detector.go`
- `internal/detection/detector_test.go`
- `internal/detection/detector_benchmark_test.go`
- `cmd/pupdate/run_execution.go`
- `cmd/pupdate/status.go`

## Allowed Discovery Scope

- Inspect adjacent detection helpers and tests under `internal/detection/` when needed to preserve ordering, ignore handling, and exact-match semantics.
- Update directly required command-side plumbing or tests in `cmd/pupdate/` when needed to thread resolved folder-blacklist config through the existing shared preflight and detection path without changing unrelated CLI behavior.
- Update planning-state files only for active-TODO synchronization, blocker recording, or completion state after review, verification, and approval.

## Outside-Expected-File Policy

- The expected files above are expected starting files, not a hard allowlist.
- A file outside that list may change only if it is directly required to satisfy `P3-T5`, is consistent with `.planning/update.md`, this plan, and the detailed brief, and is reported explicitly with rationale under `Outside expected files`.
- Outside-list edits are acceptable for directly required shared-preflight wiring or command tests, but not for docs, unrelated refactors, or later TODO work.

## Required Implementation Scope

- Extend detection inputs so the resolved folder-blacklist config reaches the detector from both `run` and `status`.
- Skip entering directories whose base name matches a configured blacklist entry anywhere in the detection walk, including default shallow scan candidates, `packages/*` children, and configured `workspace_globs` matches.
- Preserve current exact-match semantics, package-root detection behavior, ordering, `.gitignore` handling, and default latency expectations outside excluded directories.
- Add focused tests covering shallow exclusions, nested exclusions, and interaction with configured workspace expansions.
- Inspect or extend the benchmark guardrail only as needed to keep the default-path performance contract visible.

## Explicit Non-Goals And Forbidden Changes

- Do not change the folder blacklist contract into glob, regex, or path-pattern matching.
- Do not document the blacklist in the README yet.
- Do not broaden scanning beyond the already approved `workspace_globs` expansion model.
- Do not mark any TODO complete and do not commit.

## Acceptance Criteria

- Repositories without blacklist config behave exactly as they do today.
- A configured entry such as `blah` prevents detection from entering both `./blah` and `./foo/blah`.
- Non-matching sibling directories remain eligible for detection.
- Detection tests cover default-path exclusions, nested exclusions, and interaction with configured workspace expansions.
- Any outside-expected-file change is necessary, justified, and reported.

## Required Checks

- `go test ./internal/detection ./cmd/pupdate -count=1`
- `go test ./internal/detection -run '^$' -bench 'BenchmarkDetectProjectTree' -count=1`
- Manual inspection that default shallow-scan behavior remains unchanged when blacklist config is unset.

## Blocker Policy

- Stop only for unapproved product, public API or CLI, persistence or migration, security, dependency or tooling, architecture, external integration, or UX decisions not already approved by the PRD, plan, TODO brief, or repository conventions.
- Continue through normal implementation and repair issues within scope, including stale expected-file lists, directly required adjacent wiring, missing helper or test seam files, reuse of generated artifacts or equivalents, and failing checks caused by the in-scope implementation.

## Generated Artifact Policy

- Reuse existing generated classes, types, files, fixtures, schemas, and other generated equivalents when available instead of creating duplicates.
- Prefer existing detection helpers, config normalization, benchmark fixtures, and command test seams over introducing parallel structures when they already satisfy the TODO.

## Stop Conditions

- Applying the blacklist would require a broader path-pattern or glob-matching model not already approved.
- Benchmark results indicate a default-path regression that needs a new design decision.
- The work would require changing files outside the allowed list in a way that alters the approved detection contract.
- A required source file is missing and cannot be reconstructed safely from the repository.
