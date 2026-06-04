# TODO Traceability

## PRD Requirement Coverage

| PRD area | Requirement | Task IDs | Status |
| --- | --- | --- | --- |
| Highest ROI / 1 | Make async hook mode the default instead of leaving foreground as the default. | P1-T1, P1-R1 | covered |
| Product Gaps / 2 | Add opt-in broader monorepo scanning without changing the shallow default latency profile. | P3-T1, P3-T2, P3-T3, P3-R1 | covered |
| User inserted idea / 2026-06-04 | Allow blacklisting folders by exact directory name so detection does not enter matching paths such as `./blah` or `./foo/blah`. | P3-T4, P3-T5, P3-T6, P3-R1 | covered |
| Reliability / 3 | Resolve the Windows async-hook stale-lock risk. | P1-T2, P1-R1 | covered |
| Reliability / 4 | Align released platforms with tested platforms. | P1-T2, P1-R1 | covered |
| Performance / 5 | Address unsupported-OS freshness hash reuse behavior. | P1-T2, P1-R1 | covered |
| Maintainability / 6 | Centralize shared `run` and `status` preflight flow. | P2-T1, P2-R1 | covered |
| UX and Adoption / 7 | Expand installation guidance in the README. | P4-T1, P4-R1 | covered |
| UX and Adoption / 8 | Make `pupdate status` more action-oriented. | P2-T2, P2-R1 | covered |
| Developer Experience / 9 | Pin the development Go version in `mise.toml`. | P4-T2, P4-R1 | covered |
| Developer Experience / 10 | Add a small end-to-end CLI test layer. | P4-T3, P4-R1 | covered |

## Phase Dependencies

| Phase | Depends on | Unlocks |
| --- | --- | --- |
| P1 | none | P2, P3, P4 |
| P2 | P1 | P4 |
| P3 | P1 | P4 |
| P4 | P1, P2, P3 | final verification and delivery |

## Unresolved Decisions

| ID | Question | Blocks | Status |
| --- | --- | --- | --- |
| none | none | none | none |

## Resolved Planning Decisions

| Topic | Decision | Source |
| --- | --- | --- |
| Hook default | Make async the default hook mode. | User clarification during PRD-to-TODO planning on 2026-05-21 |
| Windows support | Drop Windows release artifacts instead of planning Windows bug fixes and CI coverage. | User clarification during PRD-to-TODO planning on 2026-05-21 |
| Monorepo expansion | Use opt-in `workspace_globs` config rather than named roots or max-depth recursion. | User clarification during PRD-to-TODO planning on 2026-05-21 |
| Folder blacklist scope | Apply folder blacklist entries across all detection paths, not only `workspace_globs` expansion. | User clarification during TODO insertion on 2026-06-04 |
| Folder blacklist semantics | Treat each folder blacklist entry as an exact directory-name match, not a glob or path pattern. | User clarification during TODO insertion on 2026-06-04 |
| Unsupported-OS freshness | Document the degraded unsupported-OS path instead of adding more platform-specific file identity implementations now. | User clarification during PRD-to-TODO planning on 2026-05-21 |
| Install guidance | Document Homebrew and `go install` as supported paths alongside `bin`. | User clarification during PRD-to-TODO planning on 2026-05-21 |

## Existing TODO Preservation Notes

- No existing `.planning/TODO.md`, `.planning/TODO-TRACE.md`, or `.planning/todos/` queue existed, so no prior queue state needed to be merged.
- Existing planning notes under `.planning/quick/` and `.planning/debug/` were preserved untouched.
- `.planning/quick/phase-15-performance-diagnostics-config-hook-planning.md` and `.planning/quick/post-v1-hardening-phase-planning.md` were used as approved-plan context where they overlap this PRD.
- `.planning/debug/unchanged-project-installs.md` was preserved as separate investigation context and not merged into this implementation queue because it is a blocked repro-driven debug thread, not approved PRD scope.
