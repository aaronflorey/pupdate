---
phase: 03
slug: v1-release-automation-and-milestone-closeout
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-31
---

# Phase 03 - Validation Strategy

## Test Infrastructure

| Property | Value |
|----------|-------|
| Framework | go test + GitHub Actions workflow validation |
| Quick run command | `go test ./... -count=1` |
| Full run command | `go test ./... -count=1` |
| Estimated runtime | ~15-30 seconds |

## Sampling Rate

- After every task commit: `go test ./... -count=1`
- Before phase sign-off: validate all workflow/config files parse and trigger scopes are correct
- Max feedback latency: 60 seconds

## Per-Task Verification Map

| Task ID | Plan | Requirement | Test Type | Automated Command | Status |
|---------|------|-------------|-----------|-------------------|--------|
| 03-01-01 | 01 | REL-01, REL-02 | unit/config | `go test ./... -count=1` | complete |
| 03-01-02 | 01 | REL-03 | static review | `go test ./... -count=1` | complete |
| 03-02-01 | 02 | MILE-01 | integration | `go test ./... -count=1` | complete |
| 03-02-02 | 02 | MILE-02 | manual + docs | manual shell-hook check + doc sync review | complete |

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Hook behavior remains non-blocking in interactive shell | MILE-01 | Requires real prompt lifecycle | `eval "$(pupdate init --shell bash)"` in interactive shell, `cd` across repos, confirm prompt remains responsive and status appears |

## Validation Sign-Off

- [x] All tasks have verification paths
- [x] No watch-mode commands
- [x] Feedback loop remains sub-minute for automated checks

Approval: complete (2026-04-07)
