---
status: complete
trigger: "/gsd:quick Create a new follow-up maintenance phase after Phase 09 for the two remaining audit findings: 1) make root_directories matching respect case-sensitive filesystems instead of unconditional lowercasing in cmd/pupdate/config.go and related tests, 2) preserve actual on-disk matched lockfile paths during detection/freshness so mixed-case lockfiles detected case-insensitively do not fail later on case-sensitive filesystems. Create separate narrow plans, update ROADMAP.md and STATE.md, and keep scope limited to these two items."
---

# Quick Task

## Current Focus

### hypothesis
The two remaining audit findings should be tracked as a small post-Phase-09 maintenance phase with one plan for `root_directories` matching semantics and one plan for mixed-case lockfile path preservation.

### next_action
Done.

## Evidence

- timestamp: 2026-04-30T00:00:00Z
  note: `cmd/pupdate/config.go` still lowercases resolved paths in `normalizeDirectoryForComparison`, which makes `root_directories` matching unconditionally case-insensitive.
- timestamp: 2026-04-30T00:05:00Z
  note: `cmd/pupdate/run_test.go` still contains a mixed-case configured-root test that expects this case-insensitive match to succeed.
- timestamp: 2026-04-30T00:10:00Z
  note: `internal/detection/matrix.go` canonicalizes matched signal names while `internal/freshness/engine.go` later stats and hashes `MatchedFiles`, which can break mixed-case lockfiles on case-sensitive filesystems if the canonicalized path does not exist.
- timestamp: 2026-04-30T00:12:00Z
  note: Added Phase 10 planning artifacts plus roadmap and state updates for a two-plan filesystem case-sensitivity follow-up.

## Resolution

### root_cause
The remaining audit findings were understood but not yet represented as their own narrow maintenance phase, so the roadmap and project state still implied post-v1 follow-up work was complete.

### fix
Created Phase 10 planning artifacts with two tightly scoped plans, then synchronized `.planning/ROADMAP.md` and `.planning/STATE.md` so the remaining filesystem case-sensitivity work is visible from the standard planning entry points.
