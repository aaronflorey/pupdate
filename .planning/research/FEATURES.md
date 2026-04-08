# Feature Research

**Domain:** Dependency update tooling (CLI + automation)
**Researched:** 2026-03-31
**Confidence:** MEDIUM

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Ecosystem detection by manifest/lockfiles | Renovate and Dependabot auto-detect supported package files and ecosystems | MEDIUM | For CLI, detect lockfiles/manifests reliably and map to correct package manager | 
| Run update/install and keep lockfiles consistent | Renovate/Dependabot update dependency files and lockfiles by default | MEDIUM | For CLI, execute the package manager’s install/update to refresh lockfiles | 
| Update scheduling / frequency control | Renovate and Dependabot let users control when updates run | MEDIUM | For CLI, schedule = “only on change/interval” via local state + cooldown | 
| Ignore/allow rules for dependencies/versions | Dependabot supports allow/ignore and semver-level control | MEDIUM | For CLI, ignore patterns + semver gating reduces surprise updates | 
| Reduce update noise via grouping | Dependabot supports grouped updates; Renovate supports grouped PRs | MEDIUM | For CLI, grouping = one run per repo/manager vs per-dependency | 
| Visible status and logs | Dependabot/Remediate tools surface PR info and update history | LOW | For CLI, print concise status and record last run in `.pupdate` | 

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required, but valuable.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Multi-ecosystem grouping | Consolidate updates across ecosystems into fewer runs | MEDIUM | Dependabot supports multi-ecosystem grouping; CLI can run a single “all ecosystems” pass | 
| Cooldown windows for new releases | Avoid churn from fresh releases while still updating | MEDIUM | Dependabot supports cooldown; local state can implement “delay majors X days” | 
| Preset/shareable configs | Reuse standardized org policies | MEDIUM | Renovate supports shared presets; CLI can support global config + repo overrides | 
| Deprecated dependency replacement hints | “Out with old, in with new” replacement PRs | HIGH | Renovate highlights replacements; CLI could suggest alternates in output | 
| Security/vulnerability-driven updates | Auto-fix PRs for vulnerabilities | HIGH | Snyk/Dependabot have security updates; CLI can optionally gate “update if vuln” via external scanner | 

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Always-on background daemon | “Keep everything updated continuously” | Adds latency, permissions, and platform-specific complexity | Run on `cd` with fast checks and local state (current plan) |
| Auto-merge/auto-commit updates | “No human review needed” | Risk of breaking changes and silent regressions | Print status + let users decide or integrate with existing CI workflows |
| Heavy dependency scanning on every `cd` | “Absolute certainty” | Violates low-latency constraint | Hash/mtime checks and only run when files change |

## Feature Dependencies

```
Ecosystem detection
    └──requires──> Manifest/lockfile scan
                       └──requires──> Filesystem access + path rules

Update execution
    └──requires──> Manager binary resolution (PATH)

Skip unchanged updates
    └──requires──> Local state file + hash/mtime tracking

Scheduling / cooldowns
    └──requires──> Local state file + timestamp tracking

Multi-ecosystem grouping
    └──requires──> Ecosystem detection + per-ecosystem update execution
```

### Dependency Notes

- **Ecosystem detection requires manifest/lockfile scan:** can’t decide a manager without reliable file signatures.
- **Update execution requires PATH-based manager resolution:** aligns with env managers (nvm/asdf/mise).
- **Skip unchanged updates requires local state:** lockfile hash/mtime stored in `.pupdate`.
- **Scheduling/cooldown requires timestamps:** needs state to compare last run vs window.
- **Multi-ecosystem grouping requires detection + execution:** only works after per-ecosystem logic exists.

## MVP Definition

### Launch With (v1)

Minimum viable product — what's needed to validate the concept.

- [ ] Ecosystem detection (composer + bun initially) — core value depends on correct manager selection
- [ ] Update execution + lockfile refresh — primary user outcome
- [ ] Skip unchanged updates using `.pupdate` state — preserves `cd` latency
- [ ] `.pupignore` / repo opt-out — required for safety and user control
- [ ] Clear status output + error visibility — users must trust what happened

### Add After Validation (v1.x)

Features to add once core is working.

- [ ] Configurable update windows / cooldowns — reduce noise without disabling automation
- [ ] Additional ecosystems (pnpm/npm/yarn, uv/poetry/pip, go mod, cargo) — expand coverage after MVP success
- [ ] Grouped multi-ecosystem runs — reduce multiple update passes in mixed repos

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] Security/vulnerability-driven updates — requires scanner integration and policy decisions
- [ ] Deprecated dependency replacement suggestions — complex mapping and maintenance burden

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Ecosystem detection (composer + bun) | HIGH | MEDIUM | P1 |
| Update execution + lockfile refresh | HIGH | MEDIUM | P1 |
| Skip unchanged updates via local state | HIGH | MEDIUM | P1 |
| `.pupignore` opt-out | MEDIUM | LOW | P1 |
| Status output + error visibility | HIGH | LOW | P1 |
| Additional ecosystem support | HIGH | MEDIUM | P2 |
| Update windows/cooldowns | MEDIUM | MEDIUM | P2 |
| Multi-ecosystem grouping | MEDIUM | MEDIUM | P2 |
| Security/vulnerability-driven updates | HIGH | HIGH | P3 |
| Deprecated replacement hints | LOW | HIGH | P3 |

**Priority key:**
- P1: Must have for launch
- P2: Should have, add when possible
- P3: Nice to have, future consideration

## Competitor Feature Analysis

| Feature | Dependabot | Renovate | Our Approach |
|---------|------------|----------|--------------|
| Scheduled updates | `schedule.interval` in dependabot.yml | Renovate scheduling with timezone + cron | Local state + cooldown + run-on-change |
| Grouped updates | `groups` and multi-ecosystem grouping | Group presets and package rules | Group per manager; optional multi-ecosystem pass |
| Ignore/allow rules | allow/ignore, semver gating | packageRules + presets | `.pupdate` config for ignore/allow + semver gates |
| Security/vuln-driven PRs | Security updates supported | (via Renovate + scanners) | Defer; optional future scanner integration |
| Config sharing | YAML config in repo | Shared presets | Global config + repo overrides (future) |

## Sources

- Renovate docs (automatic updates, scheduling, configuration presets): https://docs.renovatebot.com/ , https://docs.renovatebot.com/key-concepts/scheduling/
- Dependabot options reference (schedule, allow/ignore, grouping, cooldown): https://docs.github.com/en/code-security/dependabot/working-with-dependabot/dependabot-options-reference
- Dependabot grouping across ecosystems: https://docs.github.com/en/code-security/dependabot/working-with-dependabot/configuring-multi-ecosystem-updates
- Dependabot allow/ignore examples: https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/controlling-dependencies-updated
- Snyk automatic Fix/Upgrade PRs: https://docs.snyk.io/scan-with-snyk/pull-requests/snyk-pull-or-merge-requests

---
*Feature research for: dependency update tools*
*Researched: 2026-03-31*
