---
phase: 04
slug: restore-hook-visibility-and-script-opt-in-controls
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-04-08
---

# Phase 04 - Validation Strategy

## Test Infrastructure

| Property | Value |
|----------|-------|
| Framework | go test |
| Quick run command | `go test ./cmd/pupdate -count=1` |
| Full run command | `go test ./... -count=1` |
| Estimated runtime | ~20-40 seconds |

## Sampling Rate

- After every task commit: `go test ./cmd/pupdate -count=1`
- Before phase sign-off: `go test ./... -count=1`
- Max feedback latency: 60 seconds

## Per-Task Verification Map

| Task ID | Plan | Requirement | Test Type | Automated Command | Status |
|---------|------|-------------|-----------|-------------------|--------|
| 04-01-01 | 01 | STAT-01, MILE-01 | unit | `go test ./cmd/pupdate -count=1` | complete |
| 04-01-02 | 01 | EXEC-03 | unit | `go test ./cmd/pupdate -count=1` | complete |
| 04-01-03 | 01 | EXEC-03, STAT-01, MILE-01 | regression | `go test ./... -count=1` | complete |

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Hook-driven `cd` remains responsive in a real interactive shell while stderr status stays visible | MILE-01 | Requires real prompt lifecycle | Run `eval "$(pupdate init --shell bash)"` in an interactive shell, `cd` between repos, confirm prompt remains responsive and status appears |

## Validation Sign-Off

- [x] All tasks have verification paths
- [x] No watch-mode commands
- [x] Feedback loop remains sub-minute for automated checks

Approval: complete (2026-04-08)
