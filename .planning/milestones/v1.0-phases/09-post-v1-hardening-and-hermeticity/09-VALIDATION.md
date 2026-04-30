---
phase: 09
slug: post-v1-hardening-and-hermeticity
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-04-29
---

# Phase 09 - Validation Strategy

## Test Infrastructure

| Property | Value |
|----------|-------|
| Framework | go test |
| Quick run command | `go test ./internal/detection ./internal/freshness ./internal/state ./cmd/pupdate -count=1` |
| Full run command | `go test ./... -count=1` |
| Estimated runtime | ~20-45 seconds |

## Sampling Rate

- After detection/config/freshness tasks: run the narrowest affected package tests.
- Before phase sign-off: `go test ./... -count=1`.
- Max feedback latency: 60 seconds for focused checks.

## Per-Task Verification Map

| Task ID | Plan | Focus | Test Type | Automated Command | Status |
|---------|------|-------|-----------|-------------------|--------|
| 09-01-01 | 01 | Rust lockfile case handling | unit/regression | `go test ./internal/detection -count=1` | complete |
| 09-02-01 | 02 | Hermetic config-dependent command tests | unit/regression | `go test ./cmd/pupdate -count=1` | complete |
| 09-03-01 | 03 | Timeout-bound injectable git submodule freshness checks | unit/regression | `go test ./internal/freshness ./cmd/pupdate -count=1` | complete |
| 09-04-01 | 04 | Reduced unchanged-lockfile hashing cost | unit/regression | `go test ./internal/freshness ./cmd/pupdate -count=1` | complete |
| 09-05-01 | 05 | Parent-directory fsync after state-file rename | unit | `go test ./internal/state -count=1` | complete |
| 09-06-01 | 06 | Missing config treated as defaults without file creation | unit/regression | `go test ./cmd/pupdate -count=1` | complete |

## Manual-Only Verifications

| Behavior | Why Manual | Test Instructions |
|----------|------------|-------------------|
| Hook-driven runs remain responsive when git submodule status blocks or fails slowly | Needs a real shell/repo environment and timing observation | Run `pupdate run` inside a repo with submodules and confirm timeout-bounded stderr behavior remains non-blocking |

## Validation Sign-Off

- [x] All requested items have a verification path.
- [x] Quick checks remain package-scoped.
- [x] Full phase sign-off is a single `go test ./... -count=1` pass.

Approval: complete (2026-04-29)
