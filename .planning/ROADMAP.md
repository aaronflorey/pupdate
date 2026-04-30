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
- [x] **Phase 9: Post-v1 Hardening and Hermeticity (INSERTED)** - Address targeted reliability, durability, and hot-path performance follow-ups identified after milestone closeout.
- [x] **Phase 10: Filesystem Case-Sensitivity Follow-Ups (INSERTED)** - Close the remaining case-sensitive root matching and matched-lockfile path follow-ups.
- [x] **Phase 11: Kasetto and Freshness Correctness Follow-Ups** - Close the latest Kasetto execution scoping and freshness correctness audit findings. (completed 2026-04-30)
- [x] **Phase 12: Release Automation and Planning State Follow-Ups** - Reconcile duplicate release workflows and resynchronize stale roadmap/state metadata after Phase 11 closeout. (completed 2026-04-30)
- [x] **Phase 13: Final Milestone Audit Documentation Drift Follow-Ups** - Correct the remaining release-planning and CI-platform documentation drift after the Phase 12 workflow cleanup. (completed 2026-04-30)
- [x] **Phase 14: Final Documentation and Process Cleanup** - Align README config behavior docs with current missing-config semantics and reconcile Phase 10-13 validation/process metadata drift. (completed 2026-04-30)
- [x] **Phase 15: Performance, Diagnostics, and Config/Hook Follow-Ups** - Capture the next approved maintenance improvements for freshness performance, diagnostics, state cleanup, config breadth, and opt-in hook execution modes. (completed 2026-04-30)
- [x] **Phase 16: Backfill Verification for Hook Visibility and Script Opt-In (Phase 4)** - Re-establish the missing phase-level verification evidence for the Phase 4 hook visibility and lifecycle-script opt-in requirements. (completed 2026-04-30)

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
| 9. Post-v1 Hardening and Hermeticity | 6/6 | Complete | 2026-04-29 |
| 10. Filesystem Case-Sensitivity Follow-Ups | 2/2 | Complete | 2026-04-30 |
| 11. Kasetto and Freshness Correctness Follow-Ups | 3/3 | Complete    | 2026-04-30 |
| 12. Release Automation and Planning State Follow-Ups | 2/2 | Complete    | 2026-04-30 |
| 13. Final Milestone Audit Documentation Drift Follow-Ups | 2/2 | Complete    | 2026-04-30 |
| 14. Final Documentation and Process Cleanup | 2/2 | Complete | 2026-04-30 |
| 15. Performance, Diagnostics, and Config/Hook Follow-Ups | 6/6 | Complete | 2026-04-30 |
| 16. Backfill Verification for Hook Visibility and Script Opt-In (Phase 4) | 1/1 | Complete | 2026-04-30 |

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

### Phase 9: post-v1 hardening and hermeticity (inserted)

**Goal:** Close the next set of low-level maintenance gaps by hardening detection, config, freshness, hashing, and state persistence behavior without reopening milestone-scale scope.
**Requirements**: None (maintenance hardening)
**Depends on:** Phase 8
**Gap Closure:** Captures post-v1 reliability and performance follow-ups for Rust lockfile handling, command-test hermeticity, git submodule freshness, lockfile hashing cost, state durability, and missing-config defaults.
**Plans:** 6 plans

Plans:
- [x] 09-01-PLAN.md — Fix `Cargo.lock` case handling.
- [x] 09-02-PLAN.md — Make `cmd/pupdate` tests hermetic and remove ambient config coupling.
- [x] 09-03-PLAN.md — Add timeout/injection for git submodule freshness checks.
- [x] 09-04-PLAN.md — Reduce hot-path lockfile hashing cost.
- [x] 09-05-PLAN.md — Harden state-file persistence with parent-directory fsync.
- [x] 09-06-PLAN.md — Remove auto-create-on-run config behavior and treat missing config as defaults.

### Phase 10: filesystem case-sensitivity follow-ups (inserted)

**Goal:** Close the two remaining filesystem case-sensitivity audit findings without widening the maintenance scope beyond root matching and lockfile path preservation.
**Requirements**: None (maintenance hardening)
**Depends on:** Phase 9
**Gap Closure:** Captures the remaining follow-up work for filesystem-aware `root_directories` matching and preserving actual on-disk matched lockfile paths through detection and freshness.
**Plans:** 2/2 plans complete

Plans:
- [x] 10-01-PLAN.md — Make `root_directories` matching filesystem-aware.
- [x] 10-02-PLAN.md — Preserve matched lockfile paths through detection and freshness.

### Phase 11: kasetto and freshness correctness follow-ups

**Goal:** Close the latest audit findings by making Kasetto execution project-scoped, aligning Kasetto detection/execution around explicit local configs, and restoring content-correct freshness skip behavior.
**Requirements**: None (maintenance hardening)
**Depends on:** Phase 10
**Gap Closure:** Captures the new post-Phase-10 audit follow-ups for Kasetto project scoping, Kasetto local-config execution, and unsafe metadata-only lockfile hash reuse.
**Plans:** 3/3 plans complete

Plans:
- [x] 11-01-PLAN.md — Make Kasetto execution project-scoped.
- [x] 11-02-PLAN.md — Align Kasetto detection and execution around explicit local configs.
- [x] 11-03-PLAN.md — Remove or replace metadata-only lockfile hash reuse.

### Phase 12: release automation and planning state follow-ups

**Goal:** Close the latest post-Phase-11 audit findings by leaving one reliable release automation path and resynchronizing roadmap/state metadata around completed Phase 11 closeout.
**Requirements**: None (maintenance hardening)
**Depends on:** Phase 11
**Gap Closure:** Captures the new maintenance follow-ups for duplicate release workflow drift, Homebrew token wiring alignment, and stale roadmap/state metadata after Phase 11 completion.
**Plans:** 2/2 plans complete

Plans:
- [x] 12-01-PLAN.md — Reconcile duplicate release automation workflows.
- [x] 12-02-PLAN.md — Resynchronize stale roadmap and state metadata.

### Phase 13: final milestone audit documentation drift follow-ups

**Goal:** Close the remaining milestone-audit documentation drift by aligning release-planning references with the surviving `release.yaml` workflow and correcting the README CI platform claim to match the actual CI matrix.
**Requirements**: None (documentation maintenance)
**Depends on:** Phase 12
**Gap Closure:** Captures the last post-Phase-12 documentation-only follow-ups for stale release-workflow references in planning/state artifacts and the README CI platform mismatch.
**Plans:** 2/2 plans complete

Plans:
- [x] 13-01-PLAN.md — Update release-planning documentation and state text for the surviving `release.yaml` model.
- [x] 13-02-PLAN.md — Correct the README CI platform claim to match `ci.yml`.

### Phase 14: final documentation and process cleanup

**Goal:** Close the remaining documentation and planning-process drift by aligning README config behavior with the current non-auto-creating config semantics and reconciling Phase 10-13 validation/process metadata with the planning artifacts that actually exist.
**Requirements**: None (documentation/process maintenance)
**Depends on:** Phase 13
**Gap Closure:** Captures the remaining post-Phase-13 cleanup for stale README config behavior wording and stale validation/process metadata claims around completed Phases 10 to 13.
**Plans:** 2 plans

Plans:
- [x] 14-01-PLAN.md — Update README config behavior docs for non-auto-creating missing-config defaults.
- [x] 14-02-PLAN.md — Reconcile Phase 10-13 validation/process metadata with the actual planning artifacts.

### Phase 15: performance, diagnostics, and config/hook follow-ups

**Goal:** Land the next approved maintenance improvements by tightening safe hot-path freshness reuse, improving operator diagnostics, pruning stale state, broadening config surface area, and adding optional asynchronous hook execution without reopening milestone-scale scope.
**Requirements**: None (maintenance hardening)
**Depends on:** Phase 14
**Gap Closure:** Captures the newly approved post-Phase-14 maintenance work for safe lockfile-metadata reuse, a diagnostic command, stale-state pruning, latency guardrails, broader user config, and an opt-in background hook mode.
**Plans:** 6 plans

Plans:
- [x] 15-01-PLAN.md — Reuse stored lockfile metadata to short-circuit hot-path freshness hashing where safe.
- [x] 15-02-PLAN.md — Add a `pupdate status` or `pupdate doctor` diagnostic command.
- [x] 15-03-PLAN.md — Prune stale `.pupdate` entries when tracked ecosystems or directories disappear.
- [x] 15-04-PLAN.md — Add benchmarks and CI latency/performance guardrails.
- [x] 15-05-PLAN.md — Expand user config beyond `root_directories`.
- [x] 15-06-PLAN.md — Add an opt-in async/background shell hook mode.

### Phase 16: backfill verification for hook visibility and script opt-in (phase 4)

**Goal:** Re-establish requirement verification evidence for hook visibility and lifecycle-script opt-in delivery by producing the missing standalone verification artifact for Phase 4.
**Requirements**: EXEC-03, STAT-01, MILE-01
**Depends on:** Phase 15
**Gap Closure:** Closes the milestone audit evidence gaps for the missing `04-VERIFICATION.md` mapping of Phase 4 requirements.
**Plans:** 1 plan

Plans:
- [x] 16-01-PLAN.md — Add the missing Phase 4 verification artifact and close the remaining requirement traceability gap.
