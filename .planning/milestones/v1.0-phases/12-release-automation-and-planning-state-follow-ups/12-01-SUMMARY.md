---
phase: 12-release-automation-and-planning-state-follow-ups
plan: 01
subsystem: release-automation
tags: [github-actions, release-please, goreleaser]
requires: []
provides:
  - One supported release automation workflow path
  - Homebrew token wiring aligned with `.goreleaser.yaml`
affects: [release-process, planning-docs]
tech-stack:
  added: []
  patterns: [single-release-workflow, release-created-tag-goreleaser]
key-files:
  created:
    - .planning/phases/12-release-automation-and-planning-state-follow-ups/12-01-SUMMARY.md
  modified:
    - .github/workflows/release.yaml
    - .planning/phases/03-v1-release-automation-and-milestone-closeout/03-01-SUMMARY.md
    - .planning/phases/03-v1-release-automation-and-milestone-closeout/03-VERIFICATION.md
  deleted:
    - .github/workflows/release-please.yml
key-decisions:
  - "Keep `release.yaml` as the single supported release automation path because it already carries the Homebrew tap token required by `.goreleaser.yaml`."
  - "Update the phase-3 release documentation to describe the surviving Release Please plus GoReleaser flow instead of the old two-workflow layout."
requirements-completed: []
duration: 7m
completed: 2026-04-30
---

# Phase 12 Plan 01: Release automation summary

Phase 12 plan 01 removes the duplicate release workflow and leaves `release.yaml` as the single supported Release Please plus GoReleaser path, with Homebrew token wiring matching `.goreleaser.yaml`.

## Verification

- Manual review that only `.github/workflows/release.yaml` remains for release automation.
- Manual review that `.github/workflows/release.yaml` passes `HOMEBREW_TAP_GITHUB_TOKEN` to GoReleaser.

## Files Created/Modified

- `.github/workflows/release.yaml` - Retained as the single supported release workflow.
- `.github/workflows/release-please.yml` - Removed to eliminate the duplicate release path.
- `.planning/phases/03-v1-release-automation-and-milestone-closeout/03-01-SUMMARY.md` - Updated release-flow documentation to describe the surviving workflow.
- `.planning/phases/03-v1-release-automation-and-milestone-closeout/03-VERIFICATION.md` - Updated verification evidence to reference the surviving workflow.
