---
phase: 1
slug: mvp-auto-update-cli
status: complete
created: 2026-03-31
---

# Phase 1 Research — MVP Auto-Update CLI

## Scope

Implement and harden the `pupdate run` + `pupdate init` MVP for composer and bun with safe defaults, PATH-based manager resolution, `.pupignore` skip behavior, lockfile-hash skip logic, and concise status output.

## Current Codebase Findings

- `cmd/pupdate/run.go` already wires detection → freshness decision → optional install → state save.
- Runtime manager resolution already uses `exec.LookPath` (`lookPath` var), satisfying PATH-driven behavior pattern.
- `.pupignore` short-circuit exists (`hasPupIgnore`).
- State model/store already supports lockfile hash persistence in `.pupdate`.
- `init` already emits bash/zsh snippets in `cmd/pupdate/init.go`.

## Gaps vs Phase Requirements

1. Install commands are currently non-frozen and do not disable lifecycle scripts by default.
2. Status output is JSON-centric and not optimized for concise run/skip/error operator feedback.
3. Fast-skip behavior exists but should be explicit and test-verified in CLI output contract.

## Recommended Implementation Pattern

1. Keep detection/freshness/state modules intact (low-risk existing architecture).
2. Tighten manager command plans in `selectManagerPlan`:
   - composer: `install --no-interaction --prefer-dist --no-scripts`
   - bun: `install --frozen-lockfile --ignore-scripts`
3. Add explicit per-ecosystem status lines in stderr for update/skip/failure outcomes.
4. Maintain opt-in script execution via flag/environment gate (default remains no scripts).
5. Keep tests at command-layer and module-layer (`go test ./...`) for fast local verification.

## Common Pitfalls to Avoid

- Do not cache absolute manager paths across runs (violates PATH runtime detection intent).
- Do not use heavy directory recursion for detection (performance constraint on `cd` hook).
- Do not write `.pupdate` when all ecosystems skip/fail (preserve meaningful last-success semantics).

## Validation Architecture

Use `go test` with targeted package runs after each task and full suite after each wave.

- Quick: `go test ./cmd/pupdate ./internal/freshness ./internal/state -count=1`
- Full: `go test ./... -count=1`

All planned tasks include automated verification commands and grep-verifiable acceptance criteria.
