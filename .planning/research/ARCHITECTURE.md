# Architecture Research

**Domain:** Dependency update CLI (multi-ecosystem)
**Researched:** 2026-03-31
**Confidence:** MEDIUM

## Standard Architecture

### System Overview

```
┌────────────────────────────────────────────────────────────────────┐
│                         CLI / Shell Hook Layer                      │
├────────────────────────────────────────────────────────────────────┤
│  ┌──────────┐  ┌───────────────┐  ┌────────────┐  ┌──────────────┐  │
│  │  CLI     │  │ Config Loader │  │ Logger/UI │  │ Hook Init    │  │
│  └────┬─────┘  └──────┬────────┘  └────┬───────┘  └──────┬───────┘  │
│       │               │                │               │          │
├───────┴───────────────┴────────────────┴───────────────┴──────────┤
│                      Core Orchestration Layer                      │
├────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────┐ │
│  │ Repo Probe  │  │ Manager Reg. │  │ Update Plan │  │ Executor │ │
│  └────┬────────┘  └────┬─────────┘  └────┬─────────┘  └────┬─────┘ │
│       │                │                 │                │       │
├───────┴────────────────┴─────────────────┴────────────────┴───────┤
│                        Data / State Layer                           │
├────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────────────────┐  │
│  │ .pupdate    │  │ Cache/Hashes │  │ OS/Filesystem Abstraction │  │
│  └─────────────┘  └──────────────┘  └───────────────────────────┘  │
└────────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| CLI/Commands | Parse args, dispatch subcommands (`run`, `init`, `status`) | Cobra commands with shared config bootstrap |
| Config Loader | Merge config from flags/env/files, `.pupignore` | Config module + ignore resolver |
| Repo Probe | Detect ecosystems via manifest/lockfiles, compute hashes | Fast filesystem scan + stat/hash cache |
| Manager Registry | Map ecosystem → handler (detect, plan, execute) | Registry of manager structs w/ capability interface |
| Update Planner | Decide which managers should run, skip if unchanged | Change detection + policy rules (ignore, cooldown) |
| Executor | Run package manager commands with timeouts, capture output | `os/exec` with context/timeouts |
| State Store | Persist hashes and last-run results in `.pupdate` | JSON/YAML local file with schema version |
| Logger/UI | Print status, errors, non-blocking output | Structured logger + short status lines |
| Hook Init | Install shell hook script | Emit shell snippet for bash/zsh |

## Recommended Project Structure

```
cmd/
└── pupdate/              # Cobra entrypoint
    └── main.go
internal/
├── cli/                  # Command wiring and flag/config binding
├── config/               # Config loading, env/flag/file merge
├── detect/               # Repo probe + lockfile/manifest detection
├── managers/             # Manager registry + per-ecosystem handlers
│   ├── composer/
│   └── bun/
├── planner/              # Update planning + skip logic
├── exec/                 # Command runner + timeouts
├── state/                # .pupdate persistence + schema versioning
├── hooks/                # bash/zsh init snippet generation
└── ui/                   # logging, status formatting
```

### Structure Rationale

- **managers/** isolates ecosystem-specific logic and keeps the core pipeline stable as more ecosystems are added.
- **detect/** and **planner/** separate “what exists” from “what should run,” enabling fast skip paths.

## Architectural Patterns

### Pattern 1: Manager Registry (Pluggable Ecosystem Handlers)

**What:** Each ecosystem implements a standard interface: detect → plan → execute.
**When to use:** Multi-ecosystem CLIs that need consistent behavior with per-ecosystem differences.
**Trade-offs:** Slight boilerplate; large benefit for testability and expansion.

**Example:**
```go
type Manager interface {
    Detect(ctx Context, repo Repo) (bool, error)
    Plan(ctx Context, repo Repo, state State) (Plan, error)
    Execute(ctx Context, plan Plan) (Result, error)
}
```

### Pattern 2: Staged Pipeline (Detect → Plan → Execute)

**What:** Run dependency updates in deterministic stages, skipping execution when inputs are unchanged.
**When to use:** CLIs invoked frequently (e.g., on `cd`) where latency matters.
**Trade-offs:** Requires reliable state/hashing; reduces unnecessary work.

**Example:**
```go
if !probe.Changed() { return Skip("unchanged") }
plan := planner.Build(probe, state)
return executor.Run(plan)
```

### Pattern 3: Isolation Boundary for Untrusted Updates (Optional)

**What:** Run updates in a sandbox/proxy to avoid exposing secrets to untrusted dependency tooling.
**When to use:** Security-sensitive environments or tools that execute install scripts.
**Trade-offs:** Added complexity and runtime overhead; may be out of scope for v1.

## Data Flow

### Request Flow

```
Shell hook / user command
    ↓
CLI → Config Loader → Repo Probe → Manager Registry
    ↓                     ↓
State Store ← Planner ← Change Detection
    ↓
Executor → Package Manager CLI → Filesystem
    ↓
Logger/UI + State Store update
```

### Key Data Flows

1. **Detection & Skip Flow:** Repo probe computes lockfile hash → compare with `.pupdate` → planner decides skip or run.
2. **Execution Flow:** Planner emits command plan → executor runs ecosystem command → writes updated state + output summary.

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| Small repos | Single pass scan of manifest/lockfiles; no parallelism needed |
| Large monorepos | Narrow scanning to known paths; avoid deep tree walks; cache stat results |
| Many ecosystems | Consider lazy manager initialization and parallel plan building only |

### Scaling Priorities

1. **First bottleneck:** filesystem scanning on `cd` → fix with targeted file checks and cached hashes.
2. **Second bottleneck:** slow package manager invocations → fix with timeouts + skip policies.

## Anti-Patterns

### Anti-Pattern 1: Full repository scan on every `cd`

**What people do:** Walk entire tree to discover files.
**Why it's wrong:** High latency and unnecessary IO.
**Do this instead:** Check for known manifest/lockfile names at repo root and known subpaths first.

### Anti-Pattern 2: Hardcoded package manager paths

**What people do:** Assume `/usr/bin/npm` etc.
**Why it's wrong:** Breaks users with `nvm`, `asdf`, `mise`.
**Do this instead:** Resolve binaries from the current process `PATH`.

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| Package registries | Indirect via package manager CLI | Avoid direct registry calls for v1 |
| VCS (optional) | Read-only metadata (branch, commit) | Only if needed for status output |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| CLI ↔ Core Orchestrator | function calls | Keep CLI thin; logic in internal modules |
| Planner ↔ Executor | plan object | Enables dry-run/status subcommands |

## Build Order Implications

1. **Repo Probe + State Store** → required for skip logic and low-latency `run`.
2. **Manager Registry + one manager (composer/bun)** → proves detection + execution pipeline.
3. **Planner + Executor** → standardized plan object + command runner.
4. **CLI + Hook Init** → user-facing integration and install flow.
5. **Additional managers + optional isolation** → expand ecosystems, add security boundaries as needed.

## Sources

- Renovate “How Renovate works” (workflow + modules): https://docs.renovatebot.com/key-concepts/how-renovate-works/ (MEDIUM)
- Dependabot CLI README (update job flow, proxy/updater isolation): https://github.com/dependabot/cli (MEDIUM)

---
*Architecture research for: dependency update CLI*
*Researched: 2026-03-31*
