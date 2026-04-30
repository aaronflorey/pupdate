# Phase 16: backfill verification for hook visibility and script opt-in (phase 4) - Context

**Gathered:** 2026-04-30
**Status:** Ready for planning
**Mode:** Autonomous follow-up from milestone audit evidence gaps

<domain>
## Phase Boundary

Close the remaining audit evidence gap by adding the missing standalone Phase 4 verification artifact without reopening Phase 4 implementation, runtime behavior, or broader validation scope.

</domain>

<decisions>
## Implementation Decisions

### phase structure
- Keep Phase 16 as a verification-backfill phase with no product-code work and, unless new evidence gaps are discovered, a single narrow plan.

### verification scope
- The required deliverable is a standalone `.planning/phases/04-restore-hook-visibility-and-script-opt-in-controls/04-VERIFICATION.md`.
- That artifact must map `EXEC-03`, `STAT-01`, and `MILE-01` to existing Phase 4 code, tests, summary evidence, and validation coverage.
- Prefer evidence already present in Phase 4 implementation and validation artifacts instead of rerunning broader milestone analysis.

### evidence model
- Treat `04-01-SUMMARY.md` as the completion/source-of-change artifact and `04-VALIDATION.md` as the test/verification-path artifact.
- Cite the concrete command, init-snippet, and test files that prove visible stderr hook status and explicit lifecycle-script opt-in behavior.
- Keep milestone requirement mapping explicit so the milestone audit can upgrade the three partial requirements to satisfied.

### execution posture
- Keep this phase planning/documentation-only.
- Prefer the smallest artifact changes needed to make audit evidence complete and internally consistent.

</decisions>

<code_context>
## Existing Code Insights

- `.planning/v1.0-MILESTONE-AUDIT.md` identifies `EXEC-03`, `STAT-01`, and `MILE-01` as partial solely because `04-VERIFICATION.md` is missing.
- `.planning/phases/04-restore-hook-visibility-and-script-opt-in-controls/04-01-PLAN.md` already scoped those three requirements to one completed Phase 4 plan.
- `.planning/phases/04-restore-hook-visibility-and-script-opt-in-controls/04-01-SUMMARY.md` records the delivered behavior, affected files, and completed requirements.
- `.planning/phases/04-restore-hook-visibility-and-script-opt-in-controls/04-VALIDATION.md` already maps the task-level automated and manual verification paths for hook visibility, script opt-in, and interactive shell behavior.
- `cmd/pupdate/init_snippets.go`, `cmd/pupdate/init_test.go`, `cmd/pupdate/run.go`, `cmd/pupdate/run_install.go`, and `cmd/pupdate/run_test.go` contain the concrete implementation and regression evidence the missing verification report should reference.

</code_context>

<specifics>
## Specific Ideas

- Create one Phase 16 execution plan whose output is the standalone `04-VERIFICATION.md` artifact plus any minimal traceability/state sync needed after that file exists.
- Structure the verification report around one section per requirement so the audit can consume it directly.
- Reuse the audit's cited evidence where it is already specific enough, then tighten references with file-level or test-level pointers from the Phase 4 codebase.

</specifics>

<deferred>
## Deferred Ideas

- Any new product behavior, CLI flags, or shell-hook runtime changes.
- Any broader Nyquist backfill for Phases 10 through 15; Phase 16 only addresses the explicit Phase 4 verification gap.
- Any rewrite of the milestone audit beyond the minimum needed after the verification artifact exists.

</deferred>
