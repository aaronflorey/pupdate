# Pitfalls Research

**Domain:** Dependency update CLI (multi-ecosystem, on-directory-entry installs)
**Researched:** 2026-03-31
**Confidence:** MEDIUM

## Critical Pitfalls

### Pitfall 1: Triggering lifecycle scripts on every `cd`

**What goes wrong:**
Auto-running installs executes lifecycle scripts (preinstall/install/postinstall/prepare) that can compile native deps, mutate files, or run arbitrary code. This becomes a security risk and creates unpredictable slowdowns on directory entry.

**Why it happens:**
Package managers run lifecycle scripts by default during install flows (npm/yarn/pnpm), and some managers (bun) still run root scripts even if dependency scripts are restricted. A “simple install” is not a side‑effect-free operation.

**How to avoid:**
- Default to safe modes: use “frozen/ci/readonly” installs where available, or “install only if lockfile hash changed.”
- Provide an explicit allowlist/opt‑in for script execution per project.
- Log and surface when scripts would have run (so users can decide).

**Warning signs:**
- Users report unexpected builds or binary compilation on `cd`.
- Install logs show script hooks running (preinstall/postinstall/prepare).
- Security teams block the tool due to “untrusted code execution.”

**Phase to address:**
Phase 1 (Safety model + execution flags before adding more ecosystems).

---

### Pitfall 2: Mutating lockfiles during “automatic installs”

**What goes wrong:**
The CLI unintentionally updates lockfiles on directory entry, causing noisy diffs, CI failures, and developer distrust (“why did my lockfile change?”).

**Why it happens:**
Many managers will resolve or rewrite lockfiles unless a frozen/CI mode is used. Running “install” without guarding lockfile changes effectively becomes “update” in some contexts.

**How to avoid:**
- Prefer frozen/CI flags (e.g., bun `--frozen-lockfile`) and abort on mismatch.
- Hash lockfiles and only run installs when inputs changed.
- Expose a “dry‑run” or “status” mode to show what would change.

**Warning signs:**
- Lockfiles change after `pupdate run` even when dependency files are unchanged.
- PRs contain unexplained lockfile diffs after navigation.
- Teams disable auto‑runs due to noisy diffs.

**Phase to address:**
Phase 1 (Core execution guardrails + lockfile policy).

---

### Pitfall 3: Mis-handling lockfile vs. vendor dir drift (PHP/Composer)

**What goes wrong:**
The CLI trusts lockfile hashes alone, but `vendor/` may be out of sync (e.g., deleted or partially updated). Users believe dependencies are current when the actual installed tree is stale.

**Why it happens:**
Composer intentionally installs from `composer.lock` into `vendor/` and expects consistency between them. Hashing only the lockfile ignores vendor drift.

**How to avoid:**
- In composer projects, detect missing or stale `vendor/` and treat as “needs install.”
- Optionally include a lightweight vendor integrity check (presence of `vendor/composer/installed.json` or similar) without hashing the whole directory.

**Warning signs:**
- “Class not found” or autoload errors despite unchanged lockfile.
- `vendor/` missing or partially present after git operations.

**Phase to address:**
Phase 2 (Ecosystem-specific correctness rules).

---

### Pitfall 4: Running in the wrong project root (monorepos/workspaces)

**What goes wrong:**
The CLI runs installs from a subdirectory or the wrong workspace root, causing partial installs or unexpected dependency tree changes.

**Why it happens:**
Monorepos can have multiple manifests and nested package managers. Naive “nearest lockfile wins” logic misidentifies the correct root.

**How to avoid:**
- Detect workspace roots explicitly (e.g., package manager workspace config files).
- Prefer top-level root with lockfile, and provide overrides for sub‑package runs.
- Log which root was chosen and why.

**Warning signs:**
- Installs run for the wrong package when entering nested folders.
- Multiple managers detected in a single repo with no clear resolution.

**Phase to address:**
Phase 2 (Workspace detection + multi-root logic).

---

### Pitfall 5: Go module commands silently rewriting `go.mod`/`go.sum`

**What goes wrong:**
The CLI calls Go tooling that mutates `go.mod`/`go.sum` unexpectedly, leading to diffs on `cd` and confusing “why did go.mod change?” reports.

**Why it happens:**
Go commands that load packages can update module files when needed. In Go 1.16+ some commands refuse to update unless explicitly allowed, but other workflows still cause updates.

**How to avoid:**
- Use read‑only or no‑update flags where possible (`-mod=readonly` for build/test contexts).
- Only run Go commands when a module file changed or user explicitly requested update.
- Warn before any command that can modify `go.mod`/`go.sum`.

**Warning signs:**
- `go.mod`/`go.sum` changes after entering a repo with no dependency edits.
- Users report “go.mod churn” on navigation.

**Phase to address:**
Phase 3 (Go ecosystem integration hardening).

---

### Pitfall 6: Ignoring PATH-scoped runtime managers

**What goes wrong:**
The CLI uses system binaries instead of PATH‑scoped versions (nvm/asdf/mise), leading to dependency installs with mismatched tool versions.

**Why it happens:**
Dependency managers rely on correct PATH resolution in the *current* shell. If the CLI resolves binaries once or outside of the user’s environment, it can select the wrong toolchain.

**How to avoid:**
- Resolve binaries at runtime using the inherited PATH for each invocation.
- Avoid caching binary paths across shells unless explicitly configured.

**Warning signs:**
- Tool version mismatch errors (e.g., wrong Node/Bun/Composer versions).
- Works in shell, fails in pupdate.

**Phase to address:**
Phase 1 (Runtime detection + PATH resolution correctness).

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| “Hash only manifest files, ignore lockfiles” | Less I/O | Updates run when they shouldn’t; drift | Never (lockfile is the source of truth) |
| “Always run install on cd” | Simple behavior | Slow shells, script side effects | Only in MVP with opt‑out + warnings |
| “Single root detection rule” | Easy implementation | Breaks monorepos/workspaces | Only for single‑package MVP |

## Integration Gotchas

Common mistakes when connecting to external services.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| npm/yarn/pnpm/bun | Assuming installs are side‑effect‑free | Treat lifecycle scripts as potentially unsafe and user‑controlled |
| Composer | Using `update` vs `install` interchangeably | Use `install` for lockfile-consistent installs; reserve `update` for explicit user action |
| Go modules | Running `go mod tidy` on every change | Only run when explicitly requested or in maintenance workflows |

## Performance Traps

Patterns that work at small scale but fail as usage grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Full repo scans on every `cd` | Noticeable shell lag | Hash targeted files + cache | Immediately in large repos |
| Running full installs in monorepos | 10–60s delays on `cd` | Workspace‑aware selective installs | Any monorepo with 10+ packages |
| Reinstalling when lockfiles unchanged | Redundant work, IO churn | Store hash in `.pupdate` and skip | Immediately in daily navigation |

## Security Mistakes

Domain-specific security issues beyond general web security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| Auto-running lifecycle scripts from dependencies | Untrusted code execution | Default to ignore/deny scripts; require explicit trust/opt‑in |
| Installing from network on every `cd` | Supply chain exposure, token leaks | Cache + only resolve on lockfile changes |
| Silent updates | Users don’t notice tampering | Always log actions + show summary output |

## UX Pitfalls

Common user experience mistakes in this domain.

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| “Nothing happened” output | Users distrust tool | Always show status summary (skipped/updated/failed) |
| Blocking `cd` on slow installs | Shell feels broken | Run asynchronously or with timeouts + warning |
| No per‑repo opt‑out | Users uninstall tool | Support `.pupignore` or config disable |

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **Install logic:** Uses frozen/CI flags — verify lockfiles never change on `cd`.
- [ ] **Safety controls:** Script execution is opt‑in — verify scripts aren’t running by default.
- [ ] **Workspace support:** Root selection is correct — verify monorepo installs run at the intended root.
- [ ] **Composer:** Vendor drift handled — verify `vendor/` sync checks exist.

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Lifecycle scripts ran unexpectedly | MEDIUM | Disable auto‑run, re‑install with scripts off, audit changes |
| Lockfiles mutated | LOW | Revert lockfiles, re‑run with frozen/ci mode |
| Wrong root chosen | LOW | Add override config, rerun in correct root |
| Vendor drift | LOW | Delete `vendor/`, run `composer install` |

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Lifecycle scripts on `cd` | Phase 1 | Run with test project containing postinstall; confirm scripts blocked by default |
| Lockfile mutation | Phase 1 | Ensure installs are frozen/readonly and lockfiles are unchanged |
| Vendor drift (Composer) | Phase 2 | Simulate missing `vendor/`; confirm install runs | 
| Wrong project root | Phase 2 | Monorepo fixture with nested packages; verify root resolution |
| Go module file rewrites | Phase 3 | Run against Go repo; confirm `go.mod` unchanged on `cd` |
| PATH mismatch | Phase 1 | Use nvm/asdf/mise fixture; verify correct binary detection |

## Sources

- npm lifecycle scripts and install hooks (npm docs, updated 2026-03-06): https://docs.npmjs.com/cli/v11/using-npm/scripts/
- Yarn lifecycle scripts (Yarn Berry docs): https://v2.yarnpkg.com/advanced/lifecycle-scripts/
- Bun install behavior and lifecycle scripts, frozen lockfile (Bun docs): https://bun.com/docs/pm/cli/install
- Composer install/update and lockfile behavior (Composer docs): https://getcomposer.org/doc/01-basic-usage.md
- Go module system and `go` command updates to module files (Go docs): https://go.dev/ref/mod

---
*Pitfalls research for: dependency update CLI*
*Researched: 2026-03-31*
