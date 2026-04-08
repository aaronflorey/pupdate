# Project Research Summary

**Project:** pupdate
**Domain:** Dependency update CLI (cd hook, multi-ecosystem)
**Researched:** 2026-03-31
**Confidence:** MEDIUM

## Executive Summary

pupdate is a low-latency CLI that runs on directory entry to keep dependencies current without slowing shell navigation. Research converges on a Go + Cobra single-binary approach with a staged detect → plan → execute pipeline, backed by a local `.pupdate` state file to skip unnecessary work. The product’s core value hinges on fast filesystem checks, PATH-based manager resolution, and clear status output so users trust what ran (or was skipped).

The recommended build path prioritizes safety and predictability: frozen/readonly installs by default, opt‑in for lifecycle scripts, and lockfile/hash based change detection. Start with a minimal ecosystem set (composer + bun) to prove the pipeline, then expand to more managers and scheduling features. Avoid daemons and heavy file watchers that violate the low-latency constraint.

Key risks are (1) unintended script execution, (2) lockfile churn on `cd`, and (3) wrong-root detection in monorepos. Mitigation requires default safety flags, strict skip logic, explicit root selection, and visible status summaries before scaling to additional ecosystems.

## Key Findings

### Recommended Stack

Stack research supports a lean Go CLI with minimal dependencies, relying on stdlib for filesystem checks, hashing, and process execution. Go 1.26.1 and Cobra v1.10.2 give a fast single binary with subcommands (`init`, `run`, `status`) and shell completion support. Release automation should use Release Please + GoReleaser via GitHub Actions; optional tooling includes Viper for layered config and Testify for test ergonomics.

**Core technologies:**
- **Go 1.26.1**: single-binary, low-latency runtime — critical for `cd` hooks.
- **spf13/cobra v1.10.2**: CLI command structure — standard UX with minimal runtime cost.
- **Go stdlib (os/exec, context, filepath, crypto/sha256, log/slog)**: process control, hashing, and logging without extra deps.

### Expected Features

The MVP must cover ecosystem detection, update execution with lockfile refresh, skip logic based on `.pupdate` state, opt‑out via `.pupignore`, and clear status output. Differentiators like cooldown windows and multi‑ecosystem grouping are valuable but should follow validation. Security‑driven updates and dependency replacement hints are explicitly deferred to v2+ due to complexity and maintenance burden.

**Must have (table stakes):**
- Ecosystem detection via manifest/lockfiles — users expect reliable auto‑selection.
- Update execution + lockfile consistency — core outcome of the tool.
- Skip unchanged updates via `.pupdate` state — protects `cd` latency.
- `.pupignore` / per‑repo opt‑out — safety and control.
- Status output + logs — trust and debuggability.

**Should have (competitive):**
- Update windows/cooldowns — reduce update noise without disabling automation.
- Multi‑ecosystem grouping — fewer runs in mixed repos.
- Additional ecosystems — expand coverage after MVP success.

**Defer (v2+):**
- Security/vulnerability‑driven updates — requires scanner integration.
- Deprecated dependency replacement hints — high maintenance and low MVP value.

### Architecture Approach

Architecture research recommends a staged pipeline (detect → plan → execute) with a manager registry for ecosystem‑specific handlers. The system should keep CLI thin, centralize orchestration in internal modules, and persist hashes and last‑run metadata in a `.pupdate` state file for fast skip paths.

**Major components:**
1. **Repo Probe** — detect manifests/lockfiles and compute hashes efficiently.
2. **Manager Registry** — map ecosystem to detect/plan/execute handlers.
3. **Planner** — decide run/skip based on state, ignore rules, and cooldowns.
4. **Executor** — run package managers with timeouts and capture output.
5. **State Store** — persist hashes/last run metadata in `.pupdate`.
6. **Hook Init + UI** — install bash/zsh hook and present concise status.

### Critical Pitfalls

1. **Lifecycle scripts on every `cd`** — default to frozen/CI modes and require explicit opt‑in for scripts.
2. **Lockfile mutation during auto‑installs** — hash lockfiles and use frozen/readonly flags to prevent diffs.
3. **Wrong project root in monorepos** — detect workspace roots and log root selection.
4. **Composer vendor drift** — treat missing/stale `vendor/` as “needs install.”
5. **PATH mismatch for runtime managers** — resolve binaries at invocation using current PATH.

## Implications for Roadmap

Based on research, suggested phase structure:

### Phase 1: Core Safety + MVP Pipeline
**Rationale:** Establish low‑latency, safe default behavior before expanding ecosystems.
**Delivers:** Detect → plan → execute pipeline, `.pupdate` state, safe install flags, and shell hook.
**Addresses:** Detection, update execution, skip logic, `.pupignore`, status output.
**Avoids:** Lifecycle scripts, lockfile churn, PATH mismatch.

### Phase 2: Ecosystem Expansion + Workspace Correctness
**Rationale:** Add ecosystem breadth only after correctness and safety are proven.
**Delivers:** Additional managers (npm/pnpm/yarn, Python, Go, Rust), workspace‑aware root selection, and composer vendor drift checks.
**Uses:** Manager registry, planner rules, state store schema versioning.
**Avoids:** Wrong-root installs, vendor drift issues.

### Phase 3: Policy Controls + Grouping
**Rationale:** Noise reduction features depend on stable detection and state tracking.
**Delivers:** Cooldown windows, update scheduling, multi‑ecosystem grouping.
**Implements:** Planner policy extensions + UI status summaries.

### Phase 4: Advanced Intelligence (Optional v2+)
**Rationale:** Security/vuln updates and replacement hints require external integrations and higher maintenance.
**Delivers:** Scanner integrations, security‑driven update triggers, replacement recommendations.

### Phase Ordering Rationale

- Dependencies: state + planner must exist before cooldowns/grouping.
- Architecture: manager registry enables low‑risk ecosystem expansion.
- Pitfalls: safety guardrails must be solved before broader adoption.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 2:** Ecosystem‑specific correctness (workspaces, vendor drift, tool flags).
- **Phase 4:** Security/vuln integrations and replacement mapping.

Phases with standard patterns (skip research‑phase):
- **Phase 1:** Go + Cobra CLI with staged pipeline and state file patterns.
- **Phase 3:** Cooldown/grouping logic once state tracking is solid.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Official sources for Go, Cobra, Release Please, GoReleaser, linting tools. |
| Features | MEDIUM | Based on Renovate/Dependabot expectations; some CLI‑specific assumptions. |
| Architecture | MEDIUM | Patterned after Renovate/Dependabot workflows; not a direct template. |
| Pitfalls | MEDIUM | Supported by official package manager docs; mitigation patterns are inferred. |

**Overall confidence:** MEDIUM

### Gaps to Address

- **Exact safe‑install flags per manager**: verify per‑ecosystem frozen/readonly flags in implementation.
- **Workspace root detection rules**: validate rules for each ecosystem (npm/yarn/pnpm workspaces, Bun, Composer).
- **State schema evolution**: define versioning/migration strategy before expanding ecosystems.

## Sources

### Primary (HIGH confidence)
- https://go.dev/dl/?mode=json — Go 1.26.1 release
- https://github.com/spf13/cobra/releases/latest — Cobra v1.10.2
- https://github.com/spf13/viper/releases/latest — Viper v1.21.0
- https://github.com/stretchr/testify/releases/latest — Testify v1.11.1
- https://github.com/goreleaser/goreleaser/releases/latest — GoReleaser v2.15.1
- https://github.com/googleapis/release-please-action/releases/latest — release-please-action v4.4.0
- https://github.com/golangci/golangci-lint/releases/latest — golangci-lint v2.11.4
- https://docs.npmjs.com/cli/v11/using-npm/scripts/ — npm lifecycle scripts
- https://v2.yarnpkg.com/advanced/lifecycle-scripts/ — Yarn lifecycle scripts
- https://bun.com/docs/pm/cli/install — Bun install behavior/frozen lockfile
- https://getcomposer.org/doc/01-basic-usage.md — Composer install/update semantics
- https://go.dev/ref/mod — Go module file mutation rules

### Secondary (MEDIUM confidence)
- https://docs.renovatebot.com/ — Renovate feature expectations
- https://docs.renovatebot.com/key-concepts/how-renovate-works/ — Renovate workflow patterns
- https://docs.github.com/en/code-security/dependabot/working-with-dependabot/dependabot-options-reference — Dependabot features
- https://docs.github.com/en/code-security/dependabot/working-with-dependabot/configuring-multi-ecosystem-updates — multi‑ecosystem grouping
- https://github.com/dependabot/cli — Dependabot CLI workflow reference
- https://docs.snyk.io/scan-with-snyk/pull-requests/snyk-pull-or-merge-requests — security update patterns

---
*Research completed: 2026-03-31*
*Ready for roadmap: yes*
