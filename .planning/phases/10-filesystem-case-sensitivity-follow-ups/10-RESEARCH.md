# Phase 10 Research - Filesystem Case-Sensitivity Follow-Ups

**Date:** 2026-04-30
**Phase:** 10-filesystem-case-sensitivity-follow-ups
**Question:** How should the two remaining audit findings be split so filesystem case-sensitivity behavior is corrected with minimal overlap?

## Confirmed Gaps

1. **`root_directories` matching still forces case-insensitive comparisons**
   - `cmd/pupdate/config.go` lowercases both the configured root and current path before `filepath.Rel` evaluation.
   - This allows mixed-case `root_directories` entries to match on case-sensitive filesystems even when the on-disk paths differ.

2. **Detection/freshness can lose the real matched lockfile path casing**
   - Detection uses case-insensitive signal matching, but canonicalized names can replace the actual filename discovered on disk.
   - Freshness later stats and hashes the detection result path, so mixed-case lockfiles found case-insensitively can fail on case-sensitive filesystems when the canonicalized filename does not exist.

## Recommended Phase Breakdown

### Plan 10-01 - Filesystem-aware `root_directories` matching

- Remove unconditional lowercasing from `cmd/pupdate/config.go` path comparison flow.
- Update command tests so case-sensitive behavior is explicit and still preserves current `~` and top-level-root rules.

### Plan 10-02 - Preserve actual matched lockfile paths through freshness

- Keep ecosystem signal detection case-insensitive while carrying forward the actual on-disk relative path that matched.
- Update freshness/state comparisons so mixed-case lockfiles can be hashed and compared without reopening a non-existent canonicalized path.

## Research Outcome

Phase 10 should stay as a two-plan maintenance phase because the remaining findings are narrow, touch different contracts, and can be verified with focused `go test` coverage.
