# Phase 09 Research - Post-v1 Hardening and Hermeticity

**Date:** 2026-04-29
**Phase:** 09-post-v1-hardening-and-hermeticity
**Question:** How should the next maintenance phase be split to address the requested reliability and performance improvements with minimal scope overlap?

## Confirmed Gaps

1. **Rust lockfile detection is not consistently case-aware**
   - `internal/detection/matrix.go` uses canonical `Cargo.lock` entries, while existing detector tests still cover lowercase `cargo.lock` fixtures.
   - The next plan should make detector behavior and tests agree on the intended case-handling contract.

2. **`cmd/pupdate` tests still depend on ambient user-config behavior**
   - Current tests assert config-file creation in `run`, which ties command behavior to user-config side effects.
   - Hermetic tests should inject config resolution/loading seams instead of relying on process-global config paths.

3. **Git submodule freshness checks are not bounded or injectable**
   - `internal/freshness/engine.go` runs `git submodule status --recursive` directly with no timeout and no test seam.
   - This is a reliability and testability gap in a hook-path code path.

4. **Lockfile hashing is a hot-path cost center**
   - Freshness evaluation currently computes full file hashes for matched lockfiles on every pass.
   - A metadata-first fast path should be explored without weakening correctness.

5. **State persistence is not fully durable across directory metadata loss**
   - `internal/state/store.go` syncs the temp file but not the parent directory after rename.
   - Parent-directory fsync should be added where supported.

6. **Missing user config should not be treated as a file-creation event**
   - `cmd/pupdate/config.go` exposes `ensureUserConfig`, and `cmd/pupdate/run_test.go` currently expects `run` to create a default config file.
   - The desired follow-up is to treat a missing config file as implicit defaults and reserve file creation for explicit user actions only.

## Recommended Phase Breakdown

### Plan 09-01 - Cargo.lock case handling

- Align Rust detection behavior with the intended canonical lockfile contract.
- Add regression coverage for the chosen case-handling behavior.

### Plan 09-02 - Hermetic command tests and config decoupling

- Remove `cmd/pupdate` test dependence on ambient config paths and auto-created config files.
- Add injection seams or test helpers so config-dependent tests stay isolated.

### Plan 09-03 - Git submodule freshness timeout/injection

- Add an injectable command runner and timeout-bound execution for submodule status checks.
- Preserve non-blocking skip/error semantics for hook-driven runs.

### Plan 09-04 - Hot-path lockfile hashing reduction

- Introduce a cheaper unchanged-file fast path for lockfile freshness decisions.
- Re-run focused regression tests around changed vs unchanged lockfiles.

### Plan 09-05 - Durable state-file persistence

- Fsync the parent directory after renaming the temp state file into place.
- Cover the new persistence step with unit tests where feasible.

### Plan 09-06 - Missing-config defaults without auto-create

- Remove `run`-path config creation.
- Treat absent user config as default values while keeping config inspection behavior explicit.

## Research Outcome

Phase 09 should stay as a single maintenance phase with six focused plans, because each requested item maps cleanly to one concrete subsystem and can be verified with targeted `go test` coverage.
