# Active TODO Plan

## TODO ID and Objective

- ID: `P4-T1`
- Objective: Expand the README install section to support Homebrew and `go install` alongside the existing `bin` workflow.

## Source Documents Consulted

- `.planning/TODO.md`
- `.planning/update.md`
- `.planning/TODO-TRACE.md`
- `.planning/todos/P4-T1.md`
- `AGENTS.md`
- `README.md`
- `.goreleaser.yaml`
- `go.mod`

## Current Repository State Summary

- Active queue state selected from `.planning/TODO.md`: `Current TODO: P4-T1`, `Next recommended TODO: P4-T1`, so `P4-T1` is the active TODO for this run.
- `.planning/PRD.md` does not exist in this repository; the PRD source of truth for this queue is `.planning/update.md`.
- `.planning/IMPLEMENTATION-PLAN.md` does not exist in this repository; approved implementation context comes from the detailed TODO brief and `.planning/TODO-TRACE.md`.
- Fresh `git status --short` evidence was clean before planning-state sync; current worktree changes are limited to active planning-state files for `P4-T1` activation.
- `README.md` currently documents installation via `bin` only.
- `.goreleaser.yaml` already defines a Homebrew tap publication for `pupdate` under `aaronflorey/homebrew-tap`.
- `go.mod` declares module path `github.com/aaronflorey/pupdate`, which is the path needed for truthful `go install` documentation.

## Expected Starting Files

- `.planning/TODO.md`
- `.planning/PLAN.md`
- `.planning/todos/P4-T1.md`
- `README.md`

## Allowed Discovery Scope

- Inspect adjacent release and module metadata needed to keep install instructions truthful and aligned with shipped behavior.
- Update planning-state files only for active-TODO synchronization, blocker recording, or completion state after approval.

## Outside-Expected-File Policy

- The expected files above are expected starting files, not a hard allowlist.
- A file outside that list may change only if it is directly required to satisfy `P4-T1`, is consistent with `.planning/update.md`, this plan, and the detailed brief, and is reported explicitly with rationale under `Outside expected files`.
- For this docs TODO, product or release-automation changes are out of scope unless a real blocker proves the existing release story cannot support truthful install docs.

## Required Implementation Scope

- Expand the README install section with concrete Homebrew instructions.
- Expand the README install section with concrete `go install` instructions.
- Preserve the existing `bin` workflow while making clear it is one supported path rather than the only supported path.
- Keep the install wording aligned with the existing Homebrew release configuration and module path.

## Explicit Non-Goals And Forbidden Changes

- Do not change release automation, GoReleaser behavior, or package publishing.
- Do not modify unrelated README sections unless a minimal adjacent wording change is directly required for coherence.
- Do not start `P4-T2`, `P4-T3`, or `P4-R1` work.
- Do not mark the TODO complete and do not commit until review, verification, audit, and explicit user approval are all complete.

## Acceptance Criteria

- README contains concrete Homebrew install instructions.
- README contains concrete `go install` instructions.
- README still documents `bin` without implying it is the only supported path.
- The install section stays consistent with the current release configuration.

## Required Checks

- Manual inspection of README install steps against `.goreleaser.yaml` and `go.mod`.

## Correctness Evidence Plan

- Acceptance criterion 1: Confirm the README adds an install command that matches the published Homebrew tap information in `.goreleaser.yaml`.
- Acceptance criterion 2: Confirm the README adds a truthful `go install` command using the module path from `go.mod` and an appropriate package target.
- Acceptance criterion 3: Confirm the README still documents `bin` as an available workflow without exclusive wording.
- Acceptance criterion 4: Confirm reviewer and verifier both map the install text to repo evidence and flag any release-story mismatch.

## Risk Areas And Invariants

- Homebrew instructions must match the configured tap owner/repo and formula name.
- `go install` instructions must not imply unsupported versioning or packaging behavior.
- The install section should remain concise and not drift into unrelated setup guidance.
- No release automation or publishing behavior should change in this TODO.

## Blocker Policy

- Stop only for unapproved product, public API or CLI, persistence or migration, security, dependency or tooling, architecture, external integration, or UX decisions not already approved by the PRD, plan, TODO brief, or repository conventions.
- Continue through normal docs and verification issues within scope, including stale expected-file lists, directly required adjacent wording fixes, or README phrasing cleanup needed to make the install section accurate.

## Generated Artifact Policy

- Reuse existing README structure and existing release metadata as the source of truth instead of introducing new generated docs or release assets.

## Stop Conditions

- The exact Homebrew tap/formula install command cannot be stated truthfully from current repository evidence.
- The exact `go install` command or package target cannot be stated truthfully from current repository evidence.
- Required work would need to edit files outside the allowed scope for reasons not already approved.
- A required source file is missing and cannot be reconstructed safely from the repository.
