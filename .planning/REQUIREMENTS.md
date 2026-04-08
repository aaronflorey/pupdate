# Requirements: pupdate

**Defined:** 2026-03-31
**Core Value:** Keep project dependencies up to date automatically on directory entry without slowing down shell navigation.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Detection

- [ ] **DET-01**: Detect supported ecosystems by manifest/lockfiles (composer, bun) in the current repo.
- [ ] **DET-02**: Resolve the package manager binary from the current PATH at runtime.
- [ ] **DET-03**: Skip execution when a `.pupignore` file exists in the repo.

### Execution & Safety

- [ ] **EXEC-01**: Run the correct install command for each detected ecosystem (composer install, bun install).
- [ ] **EXEC-02**: Use safe/frozen install modes to avoid lockfile mutation by default.
- [ ] **EXEC-03**: Avoid running lifecycle scripts by default where the ecosystem supports it, with opt-in control.

### State & Skipping

- [ ] **STATE-01**: Store per-ecosystem lockfile hashes in `.pupdate` state.
- [ ] **STATE-02**: Skip installs when lockfile hashes are unchanged; run on first run or hash change.

### Shell Integration

- [ ] **SHELL-01**: `init` outputs bash and zsh hook snippets to run `pupdate run` on directory entry.
- [ ] **SHELL-02**: `run` completes quickly when no changes are detected.

### Status

- [ ] **STAT-01**: Print concise status for run/skip/error outcomes.

### Ecosystem Support

- [ ] **ECO-01**: MVP supports composer and bun.

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Ecosystem Support

- [ ] **ECO-02**: Support npm, pnpm, and yarn.
- [ ] **ECO-03**: Support uv, poetry, and pip.
- [ ] **ECO-04**: Support go mod and cargo.
- [ ] **ECO-05**: Validate git submodules against `.gitmodules` and update when stale.

### Release Automation

- [ ] **REL-01**: Configure Release Please for semver release PR generation.
- [ ] **REL-02**: Add GitHub Actions workflows for CI and release orchestration.
- [ ] **REL-03**: Execute GoReleaser on `v*` tags using repository release config.

### Milestone Closeout

- [ ] **MILE-01**: Verify shell hook behavior remains non-blocking with visible status output.
- [ ] **MILE-02**: Synchronize planning/project artifacts to reflect validated v1 completion.

### Policy & Scheduling

- **POL-01**: Configurable cooldown windows or schedules to reduce update noise.
- **POL-02**: Multi-ecosystem grouping to minimize repeated runs in mixed repos.

### Security Signals

- **SEC-01**: Optional security/vulnerability checks per ecosystem.

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Always-on daemon | Violates low-latency constraint and adds complexity. |
| Auto-commit/auto-merge updates | Outside local CLI scope; risks unwanted repo changes. |
| Heavy file watchers on every `cd` | Conflicts with performance constraint. |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| DET-01 | Phase 5 | Complete |
| DET-02 | Phase 5 | Complete |
| DET-03 | Phase 5 | Complete |
| EXEC-01 | Phase 5 | Complete |
| EXEC-02 | Phase 5 | Complete |
| EXEC-03 | Phase 4 | Complete |
| STATE-01 | Phase 5 | Complete |
| STATE-02 | Phase 5 | Complete |
| SHELL-01 | Phase 5 | Complete |
| SHELL-02 | Phase 5 | Complete |
| STAT-01 | Phase 4 | Complete |
| ECO-01 | Phase 5 | Complete |
| ECO-02 | Phase 6 | Complete |
| ECO-03 | Phase 6 | Complete |
| ECO-04 | Phase 6 | Complete |
| ECO-05 | Phase 6 | Complete |
| REL-01 | Phase 7 | Complete |
| REL-02 | Phase 7 | Complete |
| REL-03 | Phase 7 | Complete |
| MILE-01 | Phase 4 | Complete |
| MILE-02 | Phase 7 | Complete |

**Coverage:**
- Requirements in traceability: 21 total
- Mapped to phases: 21
- Pending: 0

---
*Requirements defined: 2026-03-31*
*Last updated: 2026-04-08 after v1.0 milestone completion*
