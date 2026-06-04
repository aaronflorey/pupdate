# Project Improvement Opportunities

This document captures the highest-value improvements identified from the current `pupdate` codebase, with emphasis on product fit, reliability, performance, and maintainability.

## Highest ROI

### 1. Make hook execution non-blocking by default, or align the docs

`pupdate` is positioned as a low-friction tool for directory-entry usage, but the default `init` mode is still `foreground`.

- `cmd/pupdate/init.go:44` defaults `--mode` to `foreground`
- `cmd/pupdate/hook.go:73-77` executes synchronously unless async is explicitly enabled
- `README.md:13-14` and `README.md:223-224` describe non-blocking behavior as part of the safety model

Why this matters:

- Blocking shell transitions on installs conflicts with the project’s core value proposition
- The current default creates surprise when users adopt the hook from the README

Recommendation:

- Either switch the default hook mode to async
- Or tighten the README and command help so foreground behavior is clearly the default and async is the recommended mode

## Product Gaps

### 2. Revisit detection depth for common monorepo layouts

Detection currently scans only:

- `.`
- depth-1 child directories
- direct children of `packages/`

Evidence:

- `internal/detection/detector.go:40-87`
- `internal/detection/detector_test.go:325-379`
- `README.md:7-11`

Why this matters:

- It misses common layouts like `apps/web`, `apps/api`, or `services/backend`
- The current behavior is fast, but it limits usefulness in modern monorepos

Recommendation:

- Keep the current shallow scan as the default for latency
- Add an opt-in config for broader scanning, additional roots, or configurable workspace patterns

## Reliability

### 3. Fix Windows background hook lock handling

Windows binaries are released, but stale async hook locks are not reliably cleaned up.

Evidence:

- `.goreleaser.yaml:10-18` builds Windows binaries
- `cmd/pupdate/hook_process_windows.go:7-16` uses `os.FindProcess`, which does not confirm whether the PID is still alive
- `cmd/pupdate/hook.go:152-180` and `cmd/pupdate/hook.go:239-264` rely on accurate process liveness for stale-lock cleanup and status reporting

Why this matters:

- A stale `.pupdate.hook.lock` can permanently suppress async runs on Windows
- Published platform support should match real behavior

Recommendation:

- Use a Windows-specific process liveness check that can distinguish active vs dead PIDs
- Add tests around stale-lock recovery semantics for Windows logic where feasible

### 4. Add Windows CI coverage or stop releasing Windows artifacts

There is a mismatch between tested platforms and released platforms.

Evidence:

- `.github/workflows/ci.yml:17-20` tests only Linux and macOS
- `.goreleaser.yaml:10-13` publishes Linux, macOS, and Windows binaries

Why this matters:

- Windows regressions can ship unnoticed
- The current Windows hook bug is an example of this gap

Recommendation:

- Add Windows CI coverage for at least targeted unit tests and CLI smoke tests
- If Windows is not a supported target, remove it from release builds until support is deliberate

## Performance

### 5. Improve lockfile hash reuse on non-Linux/macOS platforms

Freshness checks reuse stored lockfile hashes only when file identity metadata is available.

Evidence:

- `internal/freshness/engine.go:236-247` requires `FileID` and `ChangeTimeUnixNano`
- `internal/freshness/file_identity_other.go:11-12` is a no-op on non-Linux/macOS targets
- `internal/freshness/performance_guardrail_test.go:11-62` shows stored-hash reuse is expected to be materially faster than rehashing

Why this matters:

- Unchanged lockfiles may be rehashed more often on unsupported OSes
- That pushes cost into the hot path the project is trying to keep fast

Recommendation:

- Add platform-specific metadata support where possible
- If not possible, document the degraded path and constrain supported platforms more explicitly

## Maintainability

### 6. Centralize shared `run` and `status` preflight flow

`run` and `status` currently duplicate a large portion of the same setup and decision pipeline.

Evidence:

- `cmd/pupdate/run_execution.go:33-79`
- `cmd/pupdate/status.go:66-146`

Shared logic includes:

- config loading and option resolution
- home-directory and configured-root checks
- `.pupignore` checks
- detection
- state loading
- freshness evaluation

Why this matters:

- Behavior drift becomes more likely as features evolve
- Fixes or new policy checks may be added in one path and forgotten in the other

Recommendation:

- Extract a shared collection/preflight layer that returns normalized repo state and decisions
- Keep `run` and `status` separate only where behavior truly diverges

## UX and Adoption

### 7. Expand installation guidance in the README

The README currently documents installation via `bin` only.

Evidence:

- `README.md:21-27`
- `.goreleaser.yaml:47-56` already supports Homebrew publishing

Why this matters:

- Users may expect `go install` or Homebrew as standard entry points for a Go CLI
- Requiring a separate `bin` tool increases adoption friction

Recommendation:

- Add supported install paths such as Homebrew and `go install`, if those are intended to work
- If `bin` is the preferred path, explain why briefly

### 8. Make `status` more action-oriented

`pupdate status` already provides a strong diagnostic snapshot, but it could do more to guide fixes.

Evidence:

- `cmd/pupdate/status.go` already computes install readiness, manager paths, hook lock status, config state, and freshness reasons

Why this matters:

- The command is close to being a self-serve troubleshooting tool
- More explicit remediation text would reduce trial and error

Recommendation:

- Add targeted guidance for common blocked states, such as missing managers on `PATH`, foreground hook mode, or configured-root exclusions

## Developer Experience

### 9. Pin the development Go version in `mise.toml`

Evidence:

- `mise.toml:1-2` uses `go = "latest"`
- `go.mod:3` declares `go 1.26`

Why this matters:

- `latest` can drift unexpectedly for contributors and CI reproduction
- A pinned version better matches the project’s intended toolchain

Recommendation:

- Pin `mise` to the same major/minor Go version used by the project

### 10. Add a small end-to-end CLI test layer

The repository has good unit coverage in several areas, but a few lightweight integration tests would protect the highest-value flows.

Good candidates:

- `pupdate run`
- `pupdate status`
- `pupdate init`
- async hook behavior around lock lifecycle

Why this matters:

- These commands span config, detection, freshness, manager selection, and status formatting
- End-to-end tests would catch cross-module regressions that unit tests can miss

Recommendation:

- Add a focused CLI integration suite that uses temporary directories and stubbed command resolution

## Suggested Order

1. Decide whether async hook mode should be the default
2. Fix Windows hook lock behavior and align Windows release/test support
3. Centralize shared `run` and `status` preflight logic
4. Decide on a monorepo scanning strategy that preserves low latency
5. Improve installation docs and developer toolchain pinning
