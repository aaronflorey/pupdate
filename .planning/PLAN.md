# Active TODO Plan

## TODO ID and Objective

- ID: `P1-R1`
- Objective: Review the phase-1 hook and platform contract changes against the PRD and resolved support-policy decisions.

## Source Documents Consulted

- `.planning/TODO.md`
- `.planning/update.md`
- `.planning/TODO-TRACE.md`
- `.planning/todos/P1-R1.md`
- `.planning/todos/P1-T1.md`
- `.planning/todos/P1-T2.md`
- `.planning/quick/phase-15-performance-diagnostics-config-hook-planning.md`
- `.planning/quick/post-v1-hardening-phase-planning.md`
- `AGENTS.md`
- `README.md`
- `.goreleaser.yaml`
- `cmd/pupdate/init.go`
- `cmd/pupdate/init_test.go`
- `cmd/pupdate/hook_test.go`
- `go.mod`

## Current Repository State Summary

- `git status --short` returned no entries at selection time.
- `.planning/PRD.md` does not exist in this repository; the queue's active PRD source is `.planning/update.md`.
- `.planning/IMPLEMENTATION-PLAN.md` does not exist in this repository; approved implementation context comes from the phase planning docs referenced by the TODO briefs.
- The phase-1 implementation TODOs already landed as commits `408e69f` (`P1-T1`) and `ea0fb4b` (`P1-T2`).
- `README.md` now documents Linux and macOS as the supported release platforms and explains the unsupported-OS freshness limitation.
- `.goreleaser.yaml` now targets `linux` and `darwin` only.
- The remaining work for this TODO is review, verification, and any planning-file follow-up needed if defects are found.

## Exact Files Expected To Change

- `.planning/PLAN.md`
- `.planning/TODO.md`
- `.planning/TODO-TRACE.md` only if review findings require additional repair mapping
- `.planning/todos/P1-R1.md` only if reviewer notes are required by repository convention

## Required Implementation Scope

- Review the phase-1 implementation represented by `P1-T1` and `P1-T2` against the queue, PRD, and approved plan.
- Confirm coverage for async-default hook behavior, supported-platform messaging, and unsupported-OS freshness documentation.
- Record only focused planning-state follow-ups if a concrete defect is found.

## Explicit Non-Goals And Forbidden Changes

- Do not change product runtime behavior in this TODO.
- Do not edit phase-1 product code or docs unless a repair loop is required from review or verification findings.
- Do not expand scope into phase 2 or later work.
- Do not mark the TODO complete without reviewer/verifier PASS and explicit user approval.
- Do not commit.

## Acceptance Criteria

- The phase-1 implementation is compared against `P1-T1`, `P1-T2`, and the related PRD requirements.
- Review explicitly checks for scope drift, missing tests, regressions, and unsupported Windows promises.
- Verification reruns or inspects the required checks for this phase.
- Any discovered defect is either repaired through the executor loop or recorded as a focused blocker.

## Required Checks

- Review diffs for `cmd/pupdate/init.go`, related tests, `README.md`, and `.goreleaser.yaml`.
- `go test ./cmd/pupdate -count=1`
- Manual inspection that no Windows release promises remain.

## Stop Conditions

- The phase reveals a new product or support-surface decision not covered by the resolved planning clarifications.
- Required repair work would extend beyond the phase-1 task scopes without explicit approval.
- Review or verification cannot substantively prove the phase-1 contract is correct.
