# Phase 12 Research - Release Automation and Planning State Follow-Ups

**Date:** 2026-04-30
**Phase:** 12-release-automation-and-planning-state-follow-ups
**Question:** How should the latest audit findings be split so release automation has one reliable path and planning metadata consistently reflects completed Phase 11 closeout?

## Confirmed Gaps

1. **Release automation currently has duplicate GitHub Actions entry points**
   - `.github/workflows/release.yaml` and `.github/workflows/release-please.yml` both run Release Please and then conditionally run GoReleaser.
   - Keeping both workflows creates ambiguity about the intended release path and increases drift risk when one workflow is updated without the other.

2. **Homebrew publish token wiring is inconsistent with `.goreleaser.yaml` requirements**
   - `.goreleaser.yaml` expects `HOMEBREW_TAP_GITHUB_TOKEN` for the brew tap repository token.
   - `.github/workflows/release.yaml` passes that environment variable, but `.github/workflows/release-please.yml` currently does not.

3. **Planning state still presents stale post-Phase-11 metadata**
   - `.planning/ROADMAP.md` correctly marks Phase 11 complete.
   - `.planning/STATE.md` still lists Phase 11 as the current focus, still says the milestone is complete with no active follow-up, and still carries Phase 11 completion text as if no next phase exists.

## Recommended Phase Breakdown

### Plan 12-01 - Reconcile duplicate release automation workflows

- Audit the two release workflows against the original Phase 3 release intent and pick one supported release automation path.
- Update workflow and release-planning artifacts so the surviving path is the only supported one and it passes the Homebrew tap token required by `.goreleaser.yaml`.

### Plan 12-02 - Resynchronize stale roadmap and state metadata

- Update `.planning/ROADMAP.md` and `.planning/STATE.md` so they agree that Phase 11 is complete and Phase 12 is the next pending maintenance phase.
- Clear stale completion-only metadata that now conflicts with the newly planned follow-up work.

## Research Outcome

Phase 12 should stay as a two-plan maintenance phase because the new findings divide cleanly into one release-automation-path fix and one planning-state resynchronization fix, and neither requires reopening broader runtime scope.
