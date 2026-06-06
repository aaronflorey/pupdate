# Active TODO Plan

## TODO ID and Objective

- ID: `P4-T3`
- Objective: Add a focused CLI integration test layer for `run`, `status`, `init`, and async hook lock lifecycle behavior.

## Source Documents Consulted

- `.planning/TODO.md`
- `.planning/update.md`
- `.planning/TODO-TRACE.md`
- `.planning/todos/P4-T3.md`
- `AGENTS.md`
- `go.mod`
- `cmd/pupdate/run_test.go`
- `cmd/pupdate/status_test.go`
- `cmd/pupdate/init_test.go`
- `cmd/pupdate/hook_test.go`
- `cmd/pupdate/test_env_test.go`

## Current Repository State Summary

- Active queue state selected from `.planning/TODO.md`: `Current TODO: none`, `Next recommended TODO: P4-T3`, so `P4-T3` is the active TODO for this run and planning state is being synced accordingly.
- `.planning/PRD.md` does not exist in this repository; the PRD source of truth for this queue is `.planning/update.md`.
- `.planning/IMPLEMENTATION-PLAN.md` does not exist in this repository; approved implementation context comes from the detailed TODO brief and `.planning/TODO-TRACE.md`.
- Fresh `git status --short --untracked-files=all` produced no file entries before planning-state sync, so there were no unrelated pre-existing worktree changes to separate from this TODO.
- `cmd/pupdate` already has command-focused tests for `run`, `status`, `init`, and `hook`, plus shared helper setup in `test_env_test.go`, but there is no dedicated `cli_integration_test.go` layer yet.
- Existing tests already use temporary directories, environment isolation, and seam overrides such as `lookPath`, `execCommand`, `detectFn`, `evaluateFreshnessFn`, `resolveExecutablePath`, `startBackgroundProcess`, and `executeRunFn`, which should allow deterministic CLI-level coverage without real package-manager binaries.

## Expected Starting Files

- `.planning/TODO.md`
- `.planning/PLAN.md`
- `.planning/todos/P4-T3.md`
- `cmd/pupdate/cli_integration_test.go`
- `cmd/pupdate/test_env_test.go`

## Allowed Discovery Scope

- Inspect adjacent `cmd/pupdate` production and test files needed to understand current command wiring, shared seams, and lock lifecycle behavior.
- Update planning-state files only for active-TODO synchronization, blocker recording, or completion state after approval.
- Allow directly required adjacent test or command files to change only if the existing seams are insufficient and the change is the smallest PRD-consistent way to keep the new integration suite deterministic.

## Outside-Expected-File Policy

- The expected files above are expected starting files, not a hard allowlist.
- A file outside that list may change only if it is directly required to satisfy `P4-T3`, is consistent with `.planning/update.md`, this plan, and the detailed brief, and is reported explicitly with rationale under `Outside expected files`.
- For this TODO, broad command refactors, shell harnesses, CI changes, and unrelated test cleanups are out of scope unless a real blocker proves a minimal adjacent seam change is required for deterministic integration coverage.

## Required Implementation Scope

- Add a focused CLI integration-style test layer for `pupdate run`, `pupdate status`, and `pupdate init` using isolated temporary directories and command execution through the Cobra root command.
- Add explicit CLI-level regression coverage for the async hook lock lifecycle without relying on real detached background installs or real package-manager binaries.
- Reuse or minimally extend existing command-test seams so the suite remains deterministic and lightweight under normal `go test` execution.

## Explicit Non-Goals And Forbidden Changes

- Do not add a shell-based end-to-end harness, external fixtures, or real package-manager dependencies.
- Do not broaden coverage into unrelated commands or product areas beyond `run`, `status`, `init`, and async hook lock lifecycle behavior.
- Do not make behavior changes to production command flows unless a minimal testability seam adjustment is directly required and PRD-consistent.
- Do not change release, CI, dependency, or toolchain configuration.
- Do not mark the TODO complete and do not commit until review, verification, audit, and explicit user approval are all complete.

## Acceptance Criteria

- The repository has a lightweight CLI integration test layer covering `run`, `status`, `init`, and async hook lock lifecycle behavior.
- The tests run under `go test ./... -count=1` without requiring real dependency-manager binaries.
- The async hook lock lifecycle has explicit regression coverage at the CLI level.

## Required Checks

- `go test ./... -count=1`
- Manual inspection that the tests isolate filesystem and command-resolution side effects.

## Correctness Evidence Plan

- Acceptance criterion 1: point to the new integration-style test file and specific test cases covering `run`, `status`, `init`, and async hook behavior via the real command entrypoints.
- Acceptance criterion 2: confirm the test implementation uses temporary directories plus stubbed command resolution or child-process seams rather than real manager binaries, and verify with `go test ./... -count=1`.
- Acceptance criterion 3: point to explicit test cases that exercise lock creation, active-lock handling, stale-lock replacement, or child cleanup behavior at the CLI layer rather than only unit-level helper coverage.

## Risk Areas And Invariants

- The suite must stay deterministic across local environments and CI by isolating `HOME`, `XDG_CONFIG_HOME`, cwd, lock files, and manager resolution.
- The tests must protect user-visible command behavior rather than only re-testing internal helpers in isolation.
- Existing command output expectations and shared test helpers must not regress or become coupled to real background execution.
- New seams should stay minimal and should not broaden the public or production behavior surface.

## Blocker Policy

- Stop only for unapproved product, public API or CLI, persistence or migration, security, dependency or tooling, architecture, external integration, or UX decisions not already approved by the PRD, plan, TODO brief, or repository conventions.
- Continue through normal implementation and verification issues within scope, including stale expected-file lists, directly required adjacent test-seam wiring, missing focused helper coverage, generated artifacts that should be reused, or failing checks caused by the active implementation.

## Generated Artifact Policy

- Reuse existing shared test helpers, seam variables, and command constructors instead of introducing duplicate harnesses or generated fixtures.

## Stop Conditions

- Deterministic CLI integration coverage cannot be added with the existing or minimally extended seams and would require a broader testability refactor.
- The work would require real network access or real package-manager binaries.
- A required source file is missing and cannot be reconstructed safely from the repository.
