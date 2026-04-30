---
phase: 02-implement-other-package-managers-from-idea-md
plan: 01
subsystem: api
tags: [go, detection, npm, pnpm, yarn, python, rust]
requires:
  - phase: 01-mvp-auto-update-cli
    provides: deterministic detection and run payload contracts
provides:
  - Expanded deterministic signal detection for Node, Python, Go, and Rust ecosystems
  - Manager mapping in detection output for bun/npm/pnpm/yarn plus pip/go/cargo
  - Run JSON assertions locked to phase-2 detection payload shape
affects: [02-02-PLAN, 02-03-PLAN, manager-selection]
tech-stack:
  added: []
  patterns: [signal-to-ecosystem mapping, deterministic manager ordering, JSON contract tests]
key-files:
  created: [.planning/phases/02-implement-other-package-managers-from-idea-md/02-01-SUMMARY.md]
  modified: [internal/detection/matrix.go, internal/detection/detector.go, internal/detection/detector_test.go, cmd/pupdate/run_test.go]
key-decisions:
  - "Mapped managers for python/go/rust at detection time to keep run payload deterministic across ecosystems."
  - "Kept directory-level os.ReadDir detection flow unchanged to preserve AGENTS.md latency constraints."
patterns-established:
  - "Detection expands by editing matrix maps first, then extending detector tests, then run payload assertions."
requirements-completed: [ECO-02, ECO-03, ECO-04]
duration: 3m 7s
completed: 2026-03-31
---

# Phase 2 Plan 01: Expand deterministic detection contracts summary

**Deterministic multi-ecosystem detection now maps npm/pnpm/yarn and python/go/rust manager identities into run payload output for phase-2 execution work.**

## Performance

- **Duration:** 3m 7s
- **Started:** 2026-03-31T01:16:00Z
- **Completed:** 2026-03-31T01:19:07Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Expanded detection signals for Node lockfiles and Python/Go/Rust canonical files in `ecosystemSignals`.
- Added manager mapping for node/python/go/rust signals and wired detector manager extraction through shared map lookups.
- Updated and locked run JSON tests so output contracts include new managers/ecosystems without changing command execution behavior.

## Task Commits

1. **Task 1 (RED): Extend detection matrix and manager mapping for ECO-02/ECO-03/ECO-04** - `db94297` (test)
2. **Task 1 (GREEN): Extend detection matrix and manager mapping for ECO-02/ECO-03/ECO-04** - `224f914` (feat)
3. **Task 2: Lock run JSON detection payload expectations for expanded matrix** - `230dbf7` (test)

## Files Created/Modified
- `internal/detection/matrix.go` - Added phase-2 signals and manager maps.
- `internal/detection/detector.go` - Generalized manager resolution by ecosystem signal map.
- `internal/detection/detector_test.go` - Added deterministic node manager and python/go/rust detection coverage.
- `cmd/pupdate/run_test.go` - Expanded run JSON payload assertions for new ecosystem/manager coverage.

## Decisions Made
- Mapped non-node manager identifiers during detection so run payload consumers can treat manager fields consistently.
- Preserved non-recursive root directory scanning (`os.ReadDir`) to maintain low-latency shell hook behavior.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Corrected rust test signal to canonical lockfile**
- **Found during:** Task 1 (GREEN)
- **Issue:** Multi-ecosystem deterministic test used `Cargo.toml` instead of required `cargo.lock`, causing false failure against intended phase contract.
- **Fix:** Updated test fixture to use `cargo.lock`.
- **Files modified:** `internal/detection/detector_test.go`
- **Verification:** `go test ./internal/detection -count=1`
- **Committed in:** `224f914`

---

**Total deviations:** 1 auto-fixed (Rule 1)
**Impact on plan:** No scope creep; deviation was required to align tests with the explicit ECO-04 detection contract.

## Auth Gates Encountered

None.

## Issues Encountered

None.

## Known Stubs

None.

## Next Phase Readiness

- Detection and payload contracts are now in place for 02-02 manager execution planning.
- No blockers identified for continuing phase 2 implementation.

## Self-Check: PASSED

- FOUND: `.planning/phases/02-implement-other-package-managers-from-idea-md/02-01-SUMMARY.md`
- FOUND: commits `db94297`, `224f914`, `230dbf7`
