# Active TODO Plan

## TODO ID and Objective

- ID: `P4-R1`
- Objective: Review the installation, toolchain, and CLI integration-test phase for scope control and regression coverage.

## Source Documents Consulted

- `.planning/TODO.md`
- `.planning/update.md`
- `.planning/TODO-TRACE.md`
- `.planning/todos/P4-R1.md`
- `.planning/todos/P4-T1.md`
- `.planning/todos/P4-T2.md`
- `.planning/todos/P4-T3.md`
- `AGENTS.md`
- `README.md`
- `mise.toml`
- `go.mod`
- `cmd/pupdate/cli_integration_test.go`
- `cmd/pupdate/test_env_test.go`

## Current Repository State Summary

- Active queue state after planning sync: `Current TODO: P4-R1`, `Next recommended TODO: P4-R1`, so `P4-R1` is the active TODO for this run.
- `.planning/PRD.md` does not exist in this repository; the PRD source of truth for this queue is `.planning/update.md`.
- `.planning/IMPLEMENTATION-PLAN.md` does not exist in this repository; approved implementation context comes from the detailed TODO brief and `.planning/TODO-TRACE.md`.
- Fresh `git diff --stat` produced no output before planning-state sync, which is consistent with no pending product-code changes for this phase review.
- The phase under review spans the already-landed README install updates, `mise.toml` Go pin alignment with `go.mod`, and the dedicated CLI integration test layer in `cmd/pupdate/cli_integration_test.go`.
- This TODO is a review gate, so the primary deliverable is audited evidence and, only if needed, narrowly-scoped repair follow-up rather than new product behavior.

## Expected Starting Files

- `.planning/TODO.md`
- `.planning/PLAN.md`
- `.planning/TODO-TRACE.md`
- `.planning/todos/P4-R1.md`
- `README.md`
- `mise.toml`
- `go.mod`
- `cmd/pupdate/cli_integration_test.go`
- `cmd/pupdate/test_env_test.go`

## Allowed Discovery Scope

- Inspect adjacent `cmd/pupdate` production and test files needed to confirm the CLI integration suite exercises the intended command flows.
- Inspect release and contributor docs only as needed to validate installation and toolchain claims against current repository reality.
- Update planning-state files only for active-TODO synchronization, blocker recording, traceability repair mapping, or completion state after approval.

## Outside-Expected-File Policy

- The expected files above are expected starting files, not a hard allowlist.
- A file outside that list may change only if it is directly required to satisfy `P4-R1`, is consistent with `.planning/update.md`, this plan, and the detailed brief, and is reported explicitly with rationale under `Outside expected files`.
- For this TODO, new product features, broader docs rewrites, new tooling policy, CI changes, or unrelated test refactors are out of scope unless a concrete review failure proves a minimal repair is required.

## Required Implementation Scope

- Compare the implemented phase-4 work against the `P4-T1`, `P4-T2`, and `P4-T3` briefs.
- Re-check the related PRD requirements, repository constraints, and review whether the phase stayed within scope.
- Accept the phase if evidence is complete, or create narrowly-scoped repair follow-up if docs mismatch, toolchain drift, or deterministic-test gaps are found.

## Explicit Non-Goals And Forbidden Changes

- Do not add new product behavior, release-process changes, or broader tooling changes during this review task.
- Do not reopen earlier completed phases without concrete evidence tied to this phase review.
- Do not mark the TODO complete and do not commit until review, verification, audit, and explicit user approval are all complete.

## Acceptance Criteria

- The reviewer compared code against all implementation task briefs in the phase.
- The reviewer checked relevant PRD and approved-plan requirements.
- The reviewer identified scope drift, missing tests, regressions, and incomplete acceptance criteria.
- The reviewer either accepted the phase or created or updated focused repair TODOs.

## Required Checks

- Review diffs and current repository state for `README.md`, `mise.toml`, `go.mod`, `cmd/pupdate/cli_integration_test.go`, and any directly related helpers.
- `go test ./... -count=1`

## Correctness Evidence Plan

- Acceptance criterion 1: map each P4 implementation brief to concrete repository evidence in `README.md`, `mise.toml`, `go.mod`, and the CLI integration tests.
- Acceptance criterion 2: map the phase work back to `.planning/update.md` requirement areas 7, 9, and 10 plus the queue trace in `.planning/TODO-TRACE.md`.
- Acceptance criterion 3: use independent reviewer and verifier reports plus direct diff and file inspection to confirm there is no scope drift, missing regression coverage, or incomplete wiring.
- Acceptance criterion 4: if issues are found, record the focused repair path in planning state; otherwise acceptance is evidenced by substantive reviewer and verifier `PASS` reports.

## Risk Areas And Invariants

- README install guidance must match actual release and supported install paths.
- `mise.toml` must stay aligned with the project Go version in `go.mod` without broad toolchain-policy expansion.
- The CLI integration suite must remain deterministic, isolated from real manager binaries, and focused on the promised high-value command flows.
- This review must stay a review gate, not turn into unrelated product work.

## Blocker Policy

- Stop only for unapproved product, public API or CLI, persistence or migration, security, dependency or tooling, architecture, external integration, or UX decisions not already approved by the PRD, plan, TODO brief, or repository conventions.
- Continue through normal review and verification issues within scope, including stale expected-file lists, directly required adjacent inspection, missing focused evidence gathering, or failing checks caused by the active implementation.

## Generated Artifact Policy

- Reuse existing planning artifacts, test helpers, and repository configuration as evidence sources rather than duplicating state or inventing new schemas.

## Stop Conditions

- The integration suite appears flaky or would require an unapproved broader testability refactor to repair.
- README or install guidance would require release-process changes not already approved in this queue.
- A required source file is missing and cannot be reconstructed safely from the repository.
