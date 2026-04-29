# Phase 03 Research - v1 Release Automation and Milestone Closeout

**Date:** 2026-03-31
**Phase:** 03-v1-release-automation-and-milestone-closeout
**Question:** What milestone gaps block a clean v1.0 closeout?

## Confirmed Gaps

1. **Release automation is incomplete**
   - `.goreleaser.yaml` exists, but there is no `.github/workflows` pipeline and no Release Please configuration.
   - This misses the project constraint requiring semver/tagged release automation.

2. **Milestone planning artifacts are out of sync**
   - Earlier phase work is complete, but planning metadata still reports stale status in places.
   - v1 completion evidence should be normalized in roadmap/project/state docs.

3. **Validation handoff is not fully codified for milestone closeout**
   - Automated tests are green, but milestone-level verification/sign-off checklist is not captured as a dedicated phase artifact.

## Existing Patterns to Reuse

- Keep GoReleaser as the packaging backend (`.goreleaser.yaml` already present).
- Use GitHub-native release tooling (Release Please + Actions) to align with stack guidance.
- Keep all verification commands in `go test` style and avoid long-running watch flows.

## Recommended Phase Breakdown

### Plan 03-01 - Release automation wiring

- Add Release Please manifest/config files.
- Add CI workflow for test/build checks.
- Add release workflow that runs Release Please and GoReleaser on tags.

### Plan 03-02 - Milestone closeout sync

- Run milestone verification matrix (tests + shell-hook manual check).
- Update `.planning/PROJECT.md` requirement states based on validated behavior.
- Finalize `.planning/STATE.md`/roadmap progress for milestone completion.

## Risks and Mitigations

1. **Risk:** Release workflow triggers on wrong branches/tags.
   - **Mitigation:** Restrict workflow triggers to `main` and release tags.

2. **Risk:** Release Please versioning and GoReleaser tag expectations drift.
   - **Mitigation:** Use consistent `v*` tag format and test with dry-run commands where possible.

3. **Risk:** Milestone docs report completion without reproducible evidence.
   - **Mitigation:** Require explicit verify commands and checklist outputs in summary artifacts.

## Research Outcome

Phase 03 should focus on release pipeline enablement first, then milestone verification and artifact consistency for v1 closeout.
