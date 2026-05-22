# pupdate TODO

This file is the queue and status tracker for the `.planning/update.md` improvement PRD. It is planning state only. Completion is proven by implementation, review, verification, and traceability updates.

## Status Values

- `pending`: not started.
- `in_progress`: currently active.
- `blocked`: waiting on a decision or failed verification.
- `complete`: accepted and verified.

## Current Progress

Current TODO: none

Next recommended TODO: P2-T1

Last update: 2026-05-22

## Operating Rules

- Work on one TODO per run.
- Before editing code, read the matching detailed brief in `.planning/todos/`.
- Keep changes scoped to the files listed in the active brief unless a stop condition is triggered.
- Use focused executor, reviewer, and verifier passes even when the main agent performs them directly.
- Do not skip the review task at the end of each phase.
- Do not mark a TODO complete until the listed checks pass or a failure is documented as unrelated.
- Stop and ask if a TODO requires a new product, API, storage, security, or tooling decision.
- Do not stage or commit unless the user explicitly requests it in that run.

## Sub-Agent Responsibilities

- Executor: implement only the active TODO, make the smallest correct change, and report changed files plus checks run.
- Reviewer: compare the implementation against the brief, PRD, existing planning notes, and regression risk.
- Verifier: run the listed checks and confirm the acceptance criteria are satisfied.
- Repair: fix only the focused defects found by review or verification, then return the work to review.

## Queue Summary

| ID | Status | Objective | Brief | Verification |
| --- | --- | --- | --- | --- |
| P1-T1 | complete | Change `pupdate init` so async hook mode is the default and the generated hook help/tests match that contract. | `.planning/todos/P1-T1.md` | `go test ./cmd/pupdate -count=1` |
| P1-T2 | complete | Remove Windows from the shipped support surface and document the unsupported-OS freshness boundary. | `.planning/todos/P1-T2.md` | Manual inspection of README and `.goreleaser.yaml` support targets. |
| P1-R1 | complete | Review the phase-1 hook and platform contract changes against the PRD and support-policy decisions. | `.planning/todos/P1-R1.md` | Review changed files, rerun referenced checks, confirm no Windows promises remain. |
| P2-T1 | pending | Extract a shared preflight collection layer used by both `run` and `status` without changing their distinct behaviors. | `.planning/todos/P2-T1.md` | `go test ./cmd/pupdate -count=1` |
| P2-T2 | pending | Add action-oriented remediation guidance to `pupdate status` for the blocked states already detected by the shared preflight flow. | `.planning/todos/P2-T2.md` | `go test ./cmd/pupdate -count=1` |
| P2-R1 | pending | Review the shared preflight and status-guidance phase for behavior drift, scope creep, and missing regression coverage. | `.planning/todos/P2-R1.md` | Review `cmd/pupdate` diffs and rerun `go test ./cmd/pupdate -count=1`. |
| P3-T1 | pending | Define validated `workspace_globs` config support and surface it through `pupdate config`. | `.planning/todos/P3-T1.md` | `go test ./cmd/pupdate -count=1` |
| P3-T2 | pending | Apply `workspace_globs` during detection while keeping the current shallow default scan unchanged. | `.planning/todos/P3-T2.md` | `go test ./internal/detection ./cmd/pupdate -count=1` and benchmark guardrail inspection. |
| P3-T3 | pending | Document `workspace_globs` behavior, defaults, and examples in the README. | `.planning/todos/P3-T3.md` | Manual inspection of README examples and config docs. |
| P3-R1 | pending | Review the monorepo-detection phase against the latency and low-surprise constraints. | `.planning/todos/P3-R1.md` | Review diffs and rerun the listed detection/config checks. |
| P4-T1 | pending | Expand the README install section to support Homebrew and `go install` alongside the existing `bin` workflow. | `.planning/todos/P4-T1.md` | Manual inspection of README install steps against release config. |
| P4-T2 | pending | Pin the development Go version in `mise.toml` to match `go.mod` and contributor guidance. | `.planning/todos/P4-T2.md` | Manual inspection of `mise.toml`, `go.mod`, and contributing docs. |
| P4-T3 | pending | Add a focused CLI integration test layer for `run`, `status`, `init`, and async hook lock lifecycle behavior. | `.planning/todos/P4-T3.md` | `go test ./... -count=1` |
| P4-R1 | pending | Review the installation, toolchain, and CLI integration-test phase for scope control and regression coverage. | `.planning/todos/P4-R1.md` | Review diffs and rerun `go test ./... -count=1`. |

## Detailed Brief Index

- `P1-T1`: `.planning/todos/P1-T1.md`
- `P1-T2`: `.planning/todos/P1-T2.md`
- `P1-R1`: `.planning/todos/P1-R1.md`
- `P2-T1`: `.planning/todos/P2-T1.md`
- `P2-T2`: `.planning/todos/P2-T2.md`
- `P2-R1`: `.planning/todos/P2-R1.md`
- `P3-T1`: `.planning/todos/P3-T1.md`
- `P3-T2`: `.planning/todos/P3-T2.md`
- `P3-T3`: `.planning/todos/P3-T3.md`
- `P3-R1`: `.planning/todos/P3-R1.md`
- `P4-T1`: `.planning/todos/P4-T1.md`
- `P4-T2`: `.planning/todos/P4-T2.md`
- `P4-T3`: `.planning/todos/P4-T3.md`
- `P4-R1`: `.planning/todos/P4-R1.md`

## Traceability

Requirement coverage, phase dependencies, resolved recommendation branches, and preservation notes are tracked in `.planning/TODO-TRACE.md`.
