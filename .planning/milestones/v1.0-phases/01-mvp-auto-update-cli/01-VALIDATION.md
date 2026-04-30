---
phase: 1
slug: mvp-auto-update-cli
status: complete
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-31
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — standard Go test layout |
| **Quick run command** | `go test ./cmd/pupdate ./internal/freshness ./internal/state -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~20-40 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./cmd/pupdate ./internal/freshness ./internal/state -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 1-01-01 | 01 | 1 | DET-01, DET-02, ECO-01 | unit | `go test ./internal/detection ./cmd/pupdate -count=1` | ✅ | ✅ green |
| 1-01-02 | 01 | 1 | EXEC-01, EXEC-02, EXEC-03 | unit | `go test ./cmd/pupdate -count=1` | ✅ | ✅ green |
| 1-02-01 | 02 | 1 | STATE-01, STATE-02, SHELL-02 | unit | `go test ./internal/freshness ./internal/state ./cmd/pupdate -count=1` | ✅ | ✅ green |
| 1-02-02 | 02 | 1 | DET-03, STAT-01 | unit | `go test ./cmd/pupdate -count=1` | ✅ | ✅ green |
| 1-03-01 | 03 | 2 | SHELL-01 | unit | `go test ./cmd/pupdate -count=1` | ✅ | ✅ green |
| 1-03-02 | 03 | 2 | STAT-01 | integration/unit | `go test ./... -count=1` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Entering a repo via shell hook triggers run output in real shell startup context | SHELL-01, STAT-01 | Requires interactive shell startup files (`.bashrc` / `.zshrc`) and real prompt lifecycle | Run `eval "$(pupdate init --shell bash)"` in an interactive shell, `cd` between dirs, confirm status appears and command does not block prompt |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 60s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** complete (2026-04-08)
