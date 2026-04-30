# Phase 11 Research - Kasetto and Freshness Correctness Follow-Ups

**Date:** 2026-04-30
**Phase:** 11-kasetto-and-freshness-correctness-follow-ups
**Question:** How should the latest audit findings be split so Kasetto execution becomes project-scoped and freshness skip decisions remain content-correct with minimal overlap?

## Confirmed Gaps

1. **Kasetto execution is not explicitly project-scoped**
   - `cmd/pupdate/run_install.go` currently maps Kasetto to `kst sync` with no explicit project/config argument.
   - That means detection of local Kasetto files does not by itself constrain execution to the detected project inputs.

2. **Kasetto detection and execution are misaligned around local config files**
   - `internal/detection/matrix.go` recognizes `kasetto.lock`, `kasetto.yaml`, and `kasetto.yml`, but execution does not use the detected config path explicitly.
   - A lock-only Kasetto detection can therefore still run through ambient/default Kasetto resolution instead of a confirmed local config.

3. **Freshness hash reuse is correct only if metadata implies content identity, which it does not**
   - `internal/freshness/engine.go` currently reuses a previous hash when size, modtime, and mode match.
   - That optimization reduces hot-path work but allows unchanged metadata to stand in for unchanged content, which is not a safe skip predicate.

## Recommended Phase Breakdown

### Plan 11-01 - Make Kasetto execution project-scoped

- Audit the Kasetto branch in `cmd/pupdate/run_install.go` and the available detection result inputs used during install planning.
- Update install planning so detected Kasetto projects execute in a project-scoped way instead of relying on global/default resolution.

### Plan 11-02 - Align Kasetto detection and execution around explicit local configs

- Use detected `kasetto.yaml` or `kasetto.yml` files as explicit execution inputs when present.
- Ensure lock-only Kasetto detections do not fall back to global/default config behavior.

### Plan 11-03 - Remove or replace metadata-only lockfile hash reuse

- Revisit the metadata-first fast path in `internal/freshness` and either remove it or replace it with an optimization that keeps skip decisions content-correct.
- Keep verification focused on unchanged, changed, renamed, and same-metadata/different-content cases.

## Research Outcome

Phase 11 should stay as a three-plan maintenance phase because the new findings split cleanly across install execution, detection/execution contract alignment, and freshness correctness, and each item can be verified with focused `go test` coverage.
