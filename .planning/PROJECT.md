# pupdate

## What This Is

pupdate is a fast Go CLI that detects package ecosystems in the current directory and runs the correct dependency install commands automatically when you enter a repo. It keeps dependency installs up to date without noticeably slowing shell navigation.

## Core Value

Keep project dependencies up to date automatically on directory entry without slowing down shell navigation.

## Requirements

### Validated

- [x] Detect supported ecosystems by manifest/lock files and select the correct package manager binary from current PATH. (Validated in Phases 1-2)
- [x] Run the appropriate install command for each detected manager, starting with composer and bun for MVP. (Validated in Phases 1-2)
- [x] Skip work when dependency state is unchanged by storing a lockfile hash in `.pupdate` and re-running only on change or first run. (Validated in Phase 1)
- [x] Respect `.pupignore` to disable automatic runs in a repository. (Validated in Phase 1)
- [x] Provide a lightweight `run` command and an `init` command to set up shell hooks for bash and zsh. (Validated in Phase 1)
- [x] Keep hook-driven runs non-blocking while preserving visible status output and explicit lifecycle-script opt-in control. (Validated in Phase 4)
- [x] Automate semver releases with Release Please + GoReleaser + GitHub Actions. (Validated in Phase 3)
- [x] Complete milestone-level verification and planning artifact synchronization for v1 closeout. (Validated in Phases 3 and 7)

### Active

(None)

### Out of Scope

- Background daemon or watcher service — higher complexity and latency risk for v1.
- Shell-specific hardcoded binary paths — must use current PATH to support env managers.

## Context

- v1.0 is complete with support for composer, bun, npm, pnpm, yarn, uv, poetry, pip, go mod, cargo, and git submodules.
- State stored locally in `.pupdate` with lockfile hashes to avoid unnecessary installs.
- Release automation is wired through Release Please, GitHub Actions, and GoReleaser for semver-tagged builds.
- Open question: track vendor directory drift (e.g., `vendor/` for composer) in addition to lockfile hashes.

## Constraints

- **Performance**: `run` must be low-latency on `cd` — no heavy scans when nothing changed.
- **Runtime detection**: package manager binaries must be resolved via current PATH for `nvm`, `asdf`, and `mise` compatibility.
- **Shell compatibility**: v1 supports bash and zsh hook setup via `init`.
- **Safety**: auto-update behavior must be transparent and non-blocking with visible status/errors.
- **Release automation**: use Release Please + GoReleaser + GitHub Actions.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go + Cobra for CLI | Fast startup, single binary, standard CLI ergonomics | Adopted |
| Release Please + GoReleaser + GitHub Actions | Automated semver releases and cross-platform binaries | Adopted |
| Preserve stderr in quiet shell hooks and gate scripts behind an explicit flag | Restores visible status without reintroducing noisy command output or unsafe defaults | Adopted |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-04-08 after v1.0 milestone completion*
