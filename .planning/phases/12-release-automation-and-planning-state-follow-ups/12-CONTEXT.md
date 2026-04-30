# Phase 12: release automation and planning state follow-ups - Context

**Gathered:** 2026-04-30
**Status:** Ready for planning
**Mode:** Autonomous follow-up from latest milestone audit findings

<domain>
## Phase Boundary

Close the two newly identified post-Phase-11 maintenance gaps without reopening runtime package-manager behavior or broader release-process redesign.

</domain>

<decisions>
## Implementation Decisions

### phase structure
- Keep each follow-up isolated as its own plan so release automation cleanup and planning-metadata resynchronization can be executed and verified independently.

### release automation path
- Release automation should have one clearly supported GitHub Actions path.
- The surviving workflow must satisfy the Homebrew token contract already declared in `.goreleaser.yaml`.

### roadmap and state metadata
- `.planning/ROADMAP.md` and `.planning/STATE.md` should agree that Phase 11 is complete while Phase 12 is the next pending maintenance phase.
- Progress counters, current-focus fields, and pending todo lists should all describe the same execution state.

### execution posture
- Prefer the smallest planning and workflow changes that restore a single auditable release path.
- Keep verification focused on workflow/config review and planning-artifact consistency.

</decisions>

<code_context>
## Existing Code Insights

- `.github/workflows/release.yaml` and `.github/workflows/release-please.yml` currently implement overlapping Release Please plus GoReleaser flows.
- `.github/workflows/release.yaml` passes `HOMEBREW_TAP_GITHUB_TOKEN` to GoReleaser, while `.github/workflows/release-please.yml` currently does not.
- `.goreleaser.yaml` declares the Homebrew tap publish token as `{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}`.
- `.planning/ROADMAP.md` marks Phase 11 complete, but `.planning/STATE.md` still presents Phase 11 as the current focus and leaves the project in a completed state with no next phase.

</code_context>

<specifics>
## Specific Ideas

- Reconcile the duplicate release workflows down to one documented, validated path before future release work proceeds.
- Update planning metadata so completed Phase 11 closeout and pending Phase 12 maintenance work are both visible without contradiction.

</specifics>

<deferred>
## Deferred Ideas

- Broader release-process redesign beyond removing duplication and aligning token wiring.
- New milestone or product-scope requirements; this phase stays maintenance-only.

</deferred>
