---
status: blocked
trigger: "/gsd:debug pupdate still tries to install for projects that have not changed"
---

# Debug Session

## Current Focus

### hypothesis
The remaining issue is likely tied to a project-specific state shape or environment condition that is not covered by the current tests.

### next_action
Get one concrete repro from an affected project, including `pupdate status` output and its `.pupdate` state contents.

## Evidence

- timestamp: 2026-05-18T00:00:00Z
  note: Current freshness tests pass for current-format state, but they do not cover legacy basename-only lockfile keys for namespaced subdirectory ecosystems.
- timestamp: 2026-05-18T00:10:00Z
  note: Tried a compatibility fallback for basename-only subdirectory lockfile keys, but it also hid a legitimate rename/update signal, so it was reverted.
- timestamp: 2026-05-18T00:15:00Z
  note: Targeted Go tests for `./internal/freshness` and `./cmd/pupdate` pass after reverting the unsafe fallback.
- timestamp: 2026-05-18T00:20:00Z
  note: Clean sandbox repros for both a root Node project and a nested Node project run once, persist `.pupdate`, and skip on the second run under isolated HOME/XDG config.

## Resolution

### root_cause
Not reproduced in current code. Likely project-specific state or environment drift.

### fix
No product-code change landed yet because the safe failing case has not been isolated.
