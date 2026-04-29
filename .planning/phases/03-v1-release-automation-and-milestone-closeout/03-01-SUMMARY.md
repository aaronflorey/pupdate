---
phase: 03-v1-release-automation-and-milestone-closeout
plan: 01
subsystem: release
tags: [github-actions, release-please, goreleaser, ci, docs]
requires: []
provides:
  - Release Please manifest/config for root Go module
  - CI workflow running `go test ./... -count=1` on PRs and main pushes
  - Tag-based GoReleaser workflow for `v*` releases
affects: [release-process, ci]
tech-stack:
  added: []
  patterns: [manifest-mode release-please, tag-triggered goreleaser, go-version-file workflow setup]
key-files:
  created:
    - .github/workflows/ci.yml
    - .github/workflows/release-please.yml
    - .github/workflows/release.yml
    - release-please-config.json
    - .release-please-manifest.json
  modified:
    - README.md
key-decisions:
  - "Use release-please manifest mode with a single root package for explicit semver state."
  - "Trigger GoReleaser from `v*` tags so release builds are tied directly to version tags."
requirements-completed: [REL-01, REL-02, REL-03]
duration: 6m
completed: 2026-04-07
---

# Phase 3 Plan 01: Release automation wiring summary

Release automation is now wired end-to-end with CI checks on `main`/PRs, Release Please release PR generation, and GoReleaser execution on `v*` tags.

## Verification

- `go test ./... -count=1` passes.
- Workflow trigger scopes now map to plan intent:
  - `ci.yml`: pull_request + push to `main`
  - `release-please.yml`: push to `main`
  - `release.yml`: push tags `v*`

## Files Created/Modified

- `.github/workflows/ci.yml` - Adds test gate on PR/main pushes.
- `.github/workflows/release-please.yml` - Runs Release Please against manifest/config files.
- `.github/workflows/release.yml` - Runs GoReleaser on semver-style tags.
- `release-please-config.json` - Defines root Go package release strategy.
- `.release-please-manifest.json` - Tracks current root package version state.
- `README.md` - Documents release flow entry points.
