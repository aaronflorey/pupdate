# Phase 13 Research - Final Milestone Audit Documentation Drift Follow-Ups

**Date:** 2026-04-30
**Phase:** 13-final-milestone-audit-documentation-drift-follow-ups
**Question:** How should the remaining post-Phase-12 documentation drift be split so release-planning references and CI platform claims match the repository's surviving workflow files exactly?

## Confirmed Gaps

1. **Release-planning documentation still references the old workflow filename or an outdated release-path description**
   - `.github/workflows/release.yaml` is now the sole surviving release workflow.
   - Some planning artifacts still refer to `.github/workflows/release.yml` or describe the older layout instead of the current push-to-`main`/`master` plus conditional created-tag checkout model.

2. **README overstates the CI platform matrix**
   - `.github/workflows/ci.yml` runs on `ubuntu-latest` and `macos-latest`.
   - `README.md` still claims CI runs across Linux, macOS, and Windows.

## Recommended Phase Breakdown

### Plan 13-01 - Update release-planning documentation and state text

- Audit the narrow set of planning and state artifacts that still mention the old release workflow filename or stale release-flow wording.
- Update those references so they consistently describe `.github/workflows/release.yaml` and its current trigger and created-tag checkout behavior.

### Plan 13-02 - Correct the README CI platform claim

- Update the README CI section so the platform claim matches the current `ci.yml` matrix and trigger scope.
- Keep the fix limited to the inaccurate CI claim unless nearby wording must change for consistency.

## Research Outcome

Phase 13 should stay as a two-plan documentation-only follow-up because the remaining drift divides cleanly between release-planning/state references and the README CI platform statement, and neither issue requires runtime or workflow changes.
