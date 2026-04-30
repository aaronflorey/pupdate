# Phase 15: performance, diagnostics, and config/hook follow-ups - Context

**Gathered:** 2026-04-30
**Status:** Ready for planning
**Mode:** Autonomous follow-up from approved maintenance improvements

<domain>
## Phase Boundary

Track the next approved maintenance work after Phase 14 without implementing product code in this planning step and without reopening broader milestone or ecosystem-expansion scope.

</domain>

<decisions>
## Implementation Decisions

### phase structure
- Keep each approved improvement isolated as its own Phase 15 plan so performance, diagnostics, state cleanup, config expansion, and hook-mode work can be sequenced and verified independently.

### freshness hot path
- Any new lockfile-metadata reuse must preserve the content-correctness guarantees restored in Phase 11 plan 03.
- The phase should explicitly distinguish safe metadata reuse from the previously rejected metadata-only shortcut.

### operational visibility
- Diagnostics should help users understand why `pupdate` skipped, ran, or failed without forcing manual state-file inspection.
- State-pruning work should keep `.pupdate` truthful when tracked repos, ecosystems, or lockfiles disappear.

### config and hook scope
- Config expansion should build on the existing config model rather than introducing a separate configuration source.
- Background hook execution must remain opt-in and preserve transparent foreground defaults.

### execution posture
- Keep this step planning-only.
- Prefer narrow plan boundaries that let future execution stay incremental and package-scoped.

</decisions>

<code_context>
## Existing Code Insights

- Phase 11 plan 03 removed unsafe metadata-only lockfile hash reuse, so any new fast path must be stricter about when stored metadata can safely bypass rehashing.
- The current roadmap ends at Phase 14 and does not yet represent the newly approved maintenance items.
- `.planning/STATE.md` currently presents the repo as fully complete, so it needs to move back to a planned maintenance follow-up state once Phase 15 is added.
- User config today is intentionally narrow around `root_directories`, leaving additional operator controls unplanned.
- Hook behavior today is foreground and visibility-oriented, so any async mode needs explicit planning around status transparency and non-blocking semantics.

</code_context>

<specifics>
## Specific Ideas

- Keep the freshness optimization, stale-state pruning, and performance-guardrail plans separate so correctness, cleanup, and measurement do not get conflated.
- Treat diagnostics, config expansion, and async hook mode as separate user-facing follow-ups because each changes CLI behavior and documentation surface in different ways.

</specifics>

<deferred>
## Deferred Ideas

- Any runtime implementation work in `cmd/` or `internal/`.
- Any new milestone definition or release-process redesign.
- Any shell integration changes beyond planning the opt-in async mode.

</deferred>
