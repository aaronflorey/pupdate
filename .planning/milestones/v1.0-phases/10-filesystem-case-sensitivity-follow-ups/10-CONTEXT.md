# Phase 10: filesystem case-sensitivity follow-ups - Context

**Gathered:** 2026-04-30
**Status:** Completed
**Mode:** Autonomous follow-up from remaining audit findings

<domain>
## Phase Boundary

Close the two remaining filesystem case-sensitivity gaps without reopening broader detection, config, or freshness work.

</domain>

<decisions>
## Implementation Decisions

### phase structure
- Keep each remaining audit finding isolated as its own plan so matching behavior and lockfile-path behavior can be changed and verified independently.

### root directory matching
- `root_directories` matching must respect the host filesystem contract instead of forcing unconditional case-insensitive comparison.
- Existing `~` expansion and top-level-only root semantics should remain unchanged.

### matched lockfile paths
- Detection and freshness should preserve the actual on-disk matched lockfile path when a signal is discovered case-insensitively.
- Compatibility with previously stored lowercase freshness state keys may still be preserved if it can be done without losing the real matched filename.

### execution posture
- Prefer the smallest correct fixes.
- Keep verification package-scoped to `cmd/pupdate`, `internal/detection`, and `internal/freshness`.

</decisions>

<code_context>
## Existing Code Insights

- `cmd/pupdate/config.go` lowercases normalized directories before `filepath.Rel` comparisons.
- `cmd/pupdate/run_test.go` currently encodes case-insensitive `root_directories` expectations that are wrong on case-sensitive filesystems.
- `internal/detection/matrix.go` canonicalizes matched signal names, which can discard the real on-disk filename casing.
- `internal/freshness/engine.go` still hashes matched files using the detection result path and lowercases lockfile map keys.

</code_context>

<specifics>
## Specific Ideas

- Gate root-directory normalization on platform/filesystem-sensitive behavior instead of unconditional lowercasing.
- Return the actual matched relative path from detection while keeping ecosystem identification case-insensitive.
- Ensure freshness/state comparison can reuse previous hashes without requiring mixed-case lockfiles to be reopened through a canonicalized path that does not exist on disk.

</specifics>

<deferred>
## Deferred Ideas

- Broader root-selection redesigns or recursive root matching changes.
- State schema redesign beyond what is strictly needed for mixed-case lockfile compatibility.

</deferred>
