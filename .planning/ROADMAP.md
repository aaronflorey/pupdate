# Roadmap: pupdate

## Overview

Deliver a low-latency CLI that safely detects supported dependency ecosystems, runs the correct install commands only when needed, and can be wired into bash/zsh on directory entry with clear status output and automated releases.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: MVP Auto-Update CLI** - Safe detect → run pipeline with skip logic, shell hooks, and status output for composer/bun.
- [x] **Phase 2: Implement Other Package Managers from IDEA.md** - Extend safe, PATH-resolved updates to npm/pnpm/yarn, Python, Go, Rust, and git submodules.
- [x] **Phase 3: v1 Release Automation and Milestone Closeout** - Add release pipeline automation and close milestone verification/documentation gaps.
- [x] **Phase 4: Restore Hook Visibility and Script Opt-In Controls** - Close integration and flow gaps by restoring hook status visibility and adding explicit lifecycle script opt-in.
- [x] **Phase 5: Backfill Verification for MVP Core (Phase 1)** - Add missing phase verification artifacts and requirement evidence for MVP core requirements.
- [x] **Phase 6: Backfill Verification for Ecosystem Expansion (Phase 2)** - Add missing verification artifacts for ecosystem expansion requirements.
- [x] **Phase 7: Backfill Verification for Release and Milestone Closeout (Phase 3)** - Add missing verification artifacts for release automation and milestone closeout requirements.
- [x] **Phase 8: Optional Audit Tech Debt Cleanup (INSERTED)** - Add regression tests and clean up low-value exports identified by the audit.

## Phase Details

### Phase 1: MVP Auto-Update CLI
**Goal**: Users can safely auto-update dependencies on directory entry for composer and bun projects without slowing shell navigation.
**Depends on**: Nothing (first phase)
**Requirements**: DET-01, DET-02, DET-03, EXEC-01, EXEC-02, EXEC-03, STATE-01, STATE-02, SHELL-01, SHELL-02, STAT-01, ECO-01
**Success Criteria** (what must be TRUE):
  1. User can run `pupdate run` in a composer or bun repo and it resolves the correct package manager from current PATH and executes the matching install command.
  2. Installs run in safe/frozen mode by default and lifecycle scripts are skipped unless the user explicitly opts in.
  3. A repo containing `.pupignore` is skipped with a clear status message and no install command executed.
  4. After the first run, unchanged lockfiles result in a fast skip that records state in `.pupdate` and reports a concise status.
  5. User can run `pupdate init` to get bash/zsh hook snippets, and when installed, entering a repo triggers `pupdate run` with visible status output.
**Plans**: 3 plans

Plans:
- [x] 01-01-PLAN.md — Lock MVP detection and safe install command defaults for bun/composer.
- [x] 01-02-PLAN.md — Harden hash-based skip/state persistence and concise run/skip/error status output.
- [x] 01-03-PLAN.md — Finalize shell init contract, docs, and interactive hook verification.

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 1.1 → 1.2 → 2

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. MVP Auto-Update CLI | 3/3 | Complete | 2026-03-31 |
| 2. Implement Other Package Managers from IDEA.md | 3/3 | Complete | 2026-03-31 |
| 3. v1 Release Automation and Milestone Closeout | 2/2 | Complete | 2026-04-07 |
| 4. Restore Hook Visibility and Script Opt-In Controls | 1/1 | Complete | 2026-04-08 |
| 5. Backfill Verification for MVP Core (Phase 1) | 1/1 | Complete | 2026-04-08 |
| 6. Backfill Verification for Ecosystem Expansion (Phase 2) | 1/1 | Complete | 2026-04-08 |
| 7. Backfill Verification for Release and Milestone Closeout (Phase 3) | 1/1 | Complete | 2026-04-08 |
| 8. Optional Audit Tech Debt Cleanup | 1/1 | Complete | 2026-04-08 |

### Phase 2: implement other package managers from IDEA.md

**Goal:** Users can run `pupdate run` in Node, Python, Go, Rust, and git-submodule repos and get safe, PATH-resolved update behavior with the same fast skip/status semantics as MVP.
**Requirements**: ECO-02, ECO-03, ECO-04, ECO-05
**Depends on:** Phase 1
**Plans:** 3 plans

Plans:
- [x] 02-01-PLAN.md — Expand deterministic detection contracts for npm/pnpm/yarn, Python, Go, and Rust ecosystems.
- [x] 02-02-PLAN.md — Implement safe manager execution plans and status-tested run behavior for newly supported ecosystems.
- [x] 02-03-PLAN.md — Add git submodule drift-aware freshness + execution wiring and document full phase-2 ecosystem behavior.

### Phase 3: v1 release automation and milestone closeout

**Goal:** pupdate can be released via automated semver flow, and milestone artifacts accurately reflect validated v1 delivery status.
**Requirements**: REL-01, REL-02, REL-03, MILE-01, MILE-02
**Depends on:** Phase 1, Phase 2
**Plans:** 2 plans

Plans:
- [x] 03-01-PLAN.md — Add Release Please + GitHub Actions release workflow wired to GoReleaser.
- [x] 03-02-PLAN.md — Complete milestone verification pass and sync planning/docs state for v1 closeout.

### Phase 4: restore hook visibility and script opt-in controls

**Goal:** Hook-driven runs stay non-blocking while preserving visible status output, and lifecycle scripts remain disabled by default with explicit user opt-in.
**Requirements**: EXEC-03, STAT-01, MILE-01
**Depends on:** Phase 1, Phase 3
**Gap Closure:** Closes audit integration and flow gaps for hook stderr suppression and missing lifecycle script opt-in control.
**Plans:** 1 plan

Plans:
- [x] 04-01-PLAN.md — Restore visible hook stderr status and add explicit lifecycle-script opt-in.

### Phase 5: backfill verification for MVP core (phase 1)

**Goal:** Re-establish requirement verification evidence for MVP core delivery by producing required verification artifacts for phase 1 scope.
**Requirements**: DET-01, DET-02, DET-03, EXEC-01, EXEC-02, STATE-01, STATE-02, SHELL-01, SHELL-02, ECO-01
**Depends on:** Phase 4
**Gap Closure:** Closes orphaned requirement gaps for phase 1 requirements identified by milestone audit.
**Plans:** 1 plan

Plans:
- [x] 05-01-PLAN.md — Backfill Phase 1 verification artifact and finalize validation evidence.

### Phase 6: backfill verification for ecosystem expansion (phase 2)

**Goal:** Re-establish requirement verification evidence for ecosystem expansion delivery by producing required verification artifacts for phase 2 scope.
**Requirements**: ECO-02, ECO-03, ECO-04, ECO-05
**Depends on:** Phase 5
**Gap Closure:** Closes orphaned requirement gaps for phase 2 requirements identified by milestone audit.
**Plans:** 1 plan

Plans:
- [x] 06-01-PLAN.md — Backfill Phase 2 verification artifact and finalize validation evidence.

### Phase 7: backfill verification for release and milestone closeout (phase 3)

**Goal:** Re-establish requirement verification evidence for release and milestone closeout delivery by producing required verification artifacts for phase 3 scope.
**Requirements**: REL-01, REL-02, REL-03, MILE-02
**Depends on:** Phase 4, Phase 6
**Gap Closure:** Closes orphaned requirement gaps for phase 3 requirements identified by milestone audit.
**Plans:** 1 plan

Plans:
- [x] 07-01-PLAN.md — Backfill Phase 3 verification artifact and finalize release/milestone evidence.

### Phase 8: optional audit tech debt cleanup (inserted)

**Goal:** Reduce recurrence risk for audit findings by adding targeted regression coverage and cleaning identified low-value exports.
**Requirements**: None (tech debt)
**Depends on:** Phase 4
**Gap Closure:** Addresses optional audit tech-debt items not required for milestone pass gates.
**Plans:** 1 plan

Plans:
- [x] 08-01-PLAN.md — Remove low-value exported helper and record existing hook-visibility regression coverage.
