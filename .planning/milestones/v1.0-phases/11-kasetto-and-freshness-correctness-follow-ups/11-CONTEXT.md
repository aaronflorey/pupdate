# Phase 11: Kasetto and freshness correctness follow-ups - Context

**Gathered:** 2026-04-30
**Status:** Completed
**Mode:** Autonomous follow-up from latest milestone audit findings

<domain>
## Phase Boundary

Close the three newly identified Kasetto execution and freshness-correctness gaps without reopening broader runtime behavior, config-system design, or performance work.

</domain>

<decisions>
## Implementation Decisions

### phase structure
- Keep each audit finding isolated as its own plan so Kasetto scoping, Kasetto config alignment, and freshness correctness can be changed and verified independently.

### kasetto execution scoping
- Kasetto execution should be anchored to the detected project inputs rather than ambient/global tool state.
- Project detection must not allow `kst sync` to read or mutate global state when a project-scoped signal was detected.

### kasetto config alignment
- Detection and execution must agree on what constitutes a runnable local Kasetto project.
- If `kasetto.yaml` or `kasetto.yml` is the detected local config, execution should use that exact file explicitly.
- Lock-only Kasetto detection should not silently fall back to a global/default config outside the detected project.

### freshness correctness
- Skip decisions must remain content-correct even when file metadata is unchanged.
- Any retained optimization must prove correctness first; otherwise the optimization should be removed.

### execution posture
- Prefer the smallest correct fixes.
- Keep verification package-scoped to `cmd/pupdate`, `internal/detection`, and `internal/freshness`.

</decisions>

<code_context>
## Existing Code Insights

- `cmd/pupdate/run_install.go` currently executes Kasetto as `kst sync` with no explicit local config or project argument.
- `internal/detection/matrix.go` and `internal/detection/detector.go` detect `kasetto.lock`, `kasetto.yaml`, and `kasetto.yml` as Kasetto signals and preserve matched files in results.
- `internal/freshness/engine.go` currently reuses stored lockfile hashes when prior metadata exactly matches current metadata.
- `cmd/pupdate/run_test.go`, `internal/detection/detector_test.go`, and `internal/freshness/engine_test.go` already provide narrow seams for the three affected subsystems.

</code_context>

<specifics>
## Specific Ideas

- Thread detected Kasetto signal details into install planning so execution can stay project-scoped.
- Differentiate config-backed Kasetto detections from lock-only detections when deciding whether and how to execute `kst sync`.
- Replace metadata-only hash reuse with a correctness-preserving alternative, or fall back to full hashing if no safe optimization exists.

</specifics>

<deferred>
## Deferred Ideas

- Broader Kasetto feature support beyond the currently detected local config files.
- Larger freshness/state schema redesigns beyond what is strictly needed to restore content-correct skip behavior.

</deferred>
