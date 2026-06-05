# Active TODO Plan

## TODO ID and Objective

- ID: `P4-T2`
- Objective: Pin the development Go version in `mise.toml` to match `go.mod` and contributor guidance.

## Source Documents Consulted

- `.planning/TODO.md`
- `.planning/update.md`
- `.planning/TODO-TRACE.md`
- `.planning/todos/P4-T2.md`
- `AGENTS.md`
- `go.mod`
- `mise.toml`
- `CONTRIBUTING.md`

## Current Repository State Summary

- Active queue state selected from `.planning/TODO.md`: `Current TODO: none`, `Next recommended TODO: P4-T2`, so `P4-T2` is the active TODO for this run and planning state is being synced accordingly.
- `.planning/PRD.md` does not exist in this repository; the PRD source of truth for this queue is `.planning/update.md`.
- `.planning/IMPLEMENTATION-PLAN.md` does not exist in this repository; approved implementation context comes from the detailed TODO brief and `.planning/TODO-TRACE.md`.
- Fresh `git status --short` evidence was clean before planning-state sync; there were no unrelated pre-existing worktree changes to separate from this TODO.
- `go.mod` declares `go 1.26`.
- `mise.toml` still uses `go = "latest"`, which is the direct drift called out by the PRD.
- `CONTRIBUTING.md` currently tells contributors to install Go 1.26 or the version declared in `go.mod`, which already aligns with the intended pinned toolchain guidance.

## Expected Starting Files

- `.planning/TODO.md`
- `.planning/PLAN.md`
- `.planning/todos/P4-T2.md`
- `mise.toml`
- `CONTRIBUTING.md`

## Allowed Discovery Scope

- Inspect adjacent toolchain and contributor docs needed to confirm the pinned Go version is truthful and consistent.
- Update planning-state files only for active-TODO synchronization, blocker recording, or completion state after approval.

## Outside-Expected-File Policy

- The expected files above are expected starting files, not a hard allowlist.
- A file outside that list may change only if it is directly required to satisfy `P4-T2`, is consistent with `.planning/update.md`, this plan, and the detailed brief, and is reported explicitly with rationale under `Outside expected files`.
- For this tooling-alignment TODO, CI, release automation, or broader developer workflow changes are out of scope unless a real blocker proves they must change to keep the version pin truthful.

## Required Implementation Scope

- Replace the floating Go tool declaration in `mise.toml` with the pinned project-supported version derived from `go.mod`.
- Confirm contributor guidance remains consistent with the pinned version and adjust only if a minimal wording change is directly required.
- Keep the scope limited to Go toolchain alignment.

## Explicit Non-Goals And Forbidden Changes

- Do not change `go.mod`.
- Do not add new tools, tasks, or policy to `mise.toml`.
- Do not change CI setup, release configuration, or unrelated contributor guidance.
- Do not start `P4-T3` or `P4-R1` work.
- Do not mark the TODO complete and do not commit until review, verification, audit, and explicit user approval are all complete.

## Acceptance Criteria

- `mise.toml` no longer uses `latest` for Go.
- The pinned version aligns with `go.mod`.
- Contributor docs remain consistent with the pinned toolchain guidance.

## Required Checks

- Manual inspection of `mise.toml`, `go.mod`, and `CONTRIBUTING.md`.

## Correctness Evidence Plan

- Acceptance criterion 1: Confirm `mise.toml` changes from `go = "latest"` to a pinned version string.
- Acceptance criterion 2: Confirm the pinned version matches the version declared in `go.mod`.
- Acceptance criterion 3: Confirm `CONTRIBUTING.md` still matches the pinned version guidance, either unchanged because it already aligns or updated minimally if needed.

## Risk Areas And Invariants

- The pinned `mise` version must reflect the supported project Go version rather than introducing a separate policy.
- Contributor guidance must not drift away from the actual toolchain declaration in `go.mod`.
- No new toolchain policy, CI behavior, or release behavior should be introduced by this TODO.

## Blocker Policy

- Stop only for unapproved product, public API or CLI, persistence or migration, security, dependency or tooling, architecture, external integration, or UX decisions not already approved by the PRD, plan, TODO brief, or repository conventions.
- Continue through normal implementation and verification issues within scope, including stale expected-file lists, directly required adjacent doc wording fixes, or small planning-state sync edits.

## Generated Artifact Policy

- Reuse the existing `go.mod` toolchain declaration and contributor docs as the source of truth instead of inventing new version-policy files or generated artifacts.

## Stop Conditions

- The repo uses a different Go patch-version policy than `go.mod` alone communicates.
- Pinning the version would require changing CI setup or release configuration beyond the allowed scope.
- Pinning the version would conflict with an existing documented contributor workflow.
- A required source file is missing and cannot be reconstructed safely from the repository.
