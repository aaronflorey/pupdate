# Phase 09: post-v1 hardening and hermeticity - Context

**Gathered:** 2026-04-29
**Status:** Ready for planning
**Mode:** Autonomous backfill from prior review and user-approved phase split

<domain>
## Phase Boundary

Close the next set of low-level maintenance gaps in detection, command-test hermeticity, freshness evaluation, state persistence, and config loading without expanding milestone scope.

</domain>

<decisions>
## Implementation Decisions

### phase structure
- Keep each requested hardening item isolated as its own plan so execution stays narrow and verification remains package-scoped.

### config behavior
- Missing user config should behave like implicit defaults.
- `pupdate run` must not create a config file as a side effect.

### execution posture
- Prefer the smallest correct fixes.
- Preserve existing user-visible behavior unless a plan explicitly changes it.
- Use focused package tests for each plan, then full-phase verification before phase closeout.

### the agent's Discretion
- Exact implementation details for detection, test seams, freshness metadata, and state durability are at the agent's discretion as long as behavior stays transparent and low-latency.

</decisions>

<code_context>
## Existing Code Insights

- `internal/detection/matrix.go` canonicalizes supported signal names and currently lowercases detected names.
- `internal/freshness/engine.go` hashes matched files on every run and shells out to `git submodule status --recursive`.
- `internal/state/store.go` uses temp-file write + rename for `.pupdate` persistence.
- `cmd/pupdate/config.go` currently creates a default config file during load.
- `cmd/pupdate/run_test.go` contains config-coupled tests that are not hermetic in ambient environments.

</code_context>

<specifics>
## Specific Ideas

- Preserve the canonical on-disk `Cargo.lock` filename through detection and freshness.
- Add explicit seams for git submodule status and config-path behavior where tests need isolation.
- Use metadata-first lockfile comparison to avoid unnecessary rehashing on unchanged files.

</specifics>

<deferred>
## Deferred Ideas

- Broad config-system redesigns beyond removing auto-create-on-run.
- Large architecture refactors in `cmd/pupdate` unless a narrow seam is insufficient.

</deferred>
