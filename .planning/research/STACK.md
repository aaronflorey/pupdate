# Stack Research

**Domain:** Fast local dependency-update CLI (repo-entry hook)
**Researched:** 2026-03-31
**Confidence:** HIGH

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.26.1 | Single-binary, low-latency CLI runtime | Fast startup and low memory overhead are critical for a `cd` hook; Go produces static binaries with predictable performance. **Confidence: HIGH** (official Go downloads list) |
| spf13/cobra | v1.10.2 | CLI command structure, flags, help, completions | De‑facto standard for Go CLIs; supports subcommands, shell completion, and clean UX without heavy runtime cost. **Confidence: HIGH** (official release) |
| Go stdlib (os/exec, context, filepath, crypto/sha256, log/slog) | Go 1.26.1 | Process execution, hashing, filesystem checks, structured logging | Keeps dependencies minimal and predictable; `crypto/sha256` is fast enough for lockfile hashing while avoiding external libs; `slog` standardizes logs without overhead. **Confidence: HIGH** (official Go docs) |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| spf13/viper | v1.21.0 | Config file + env + flag layering | Only if you add user‑level config beyond `.pupdate` (e.g., global defaults, per‑user overrides). Viper 1.21.0 requires Go ≥1.23. **Confidence: HIGH** (official release + go.mod) |
| stretchr/testify | v1.11.1 | Assertions/mocks for tests | Use to keep unit tests readable and compact; avoids re‑implementing helpers. **Confidence: HIGH** (official release) |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| GoReleaser | v2.15.1 | Cross‑platform builds + release packaging | Standard for Go CLI release pipelines; integrates with GitHub Actions. **Confidence: HIGH** (official release) |
| release-please-action | v4.4.0 | Automated semver + changelog PRs | Matches the project’s release‑automation constraint. **Confidence: HIGH** (official release) |
| golangci-lint | v2.11.4 | Fast multi‑linter | Enforces style + correctness without manual config sprawl. **Confidence: HIGH** (official release) |

## Installation

```bash
# Core
go get github.com/spf13/cobra@v1.10.2

# Supporting (only if needed)
go get github.com/spf13/viper@v1.21.0
go get github.com/stretchr/testify@v1.11.1

# Dev tools
go install github.com/goreleaser/goreleaser@v2.15.1
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.11.4
```

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| Go + Cobra | Rust + clap | If you already have a Rust toolchain and want maximal single‑binary size/perf control. |
| Cobra | urfave/cli v2 | If you prefer a smaller API surface and fewer dependencies (still Go). |
| Viper (optional) | stdlib + custom config parsing | If you want minimal deps and only simple flags + JSON config. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| Node‑based CLI frameworks (e.g., oclif) | Cold‑start latency and runtime overhead are noticeable on `cd` hooks. | Go binary with Cobra |
| Background daemon or file‑watcher (fsnotify) for every repo | Adds CPU/memory overhead and cross‑platform edge cases; violates low‑latency constraint. | Lockfile hash + cheap mtime checks on `cd` |
| Hard‑coding package‑manager paths | Breaks environment managers (nvm/asdf/mise) and violates PATH‑based detection requirement. | Resolve manager binaries from current PATH |

## Stack Patterns by Variant

**If you want the smallest dependency footprint:**
- Use Cobra + stdlib only
- Keep `.pupdate` as JSON and avoid Viper
- Because startup latency and binary size matter most for `cd`

**If you need per‑user configuration:**
- Add Viper for config/env layering
- Because it standardizes override precedence without custom plumbing

## Version Compatibility

| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| Go 1.26.1 | Cobra v1.10.2 | Cobra’s go.mod sets go 1.15, so 1.26.1 is fully compatible. |
| Go 1.26.1 | Viper v1.21.0 | Viper requires Go ≥1.23.0. |

## Sources

- https://go.dev/dl/?mode=json — Go 1.26.1 current stable (HIGH)
- https://github.com/spf13/cobra/releases/latest — Cobra v1.10.2 (HIGH)
- https://raw.githubusercontent.com/spf13/cobra/v1.10.2/go.mod — Cobra minimum Go version (HIGH)
- https://github.com/spf13/viper/releases/latest — Viper v1.21.0 (HIGH)
- https://raw.githubusercontent.com/spf13/viper/v1.21.0/go.mod — Viper minimum Go version (HIGH)
- https://github.com/stretchr/testify/releases/latest — Testify v1.11.1 (HIGH)
- https://github.com/goreleaser/goreleaser/releases/latest — GoReleaser v2.15.1 (HIGH)
- https://github.com/googleapis/release-please-action/releases/latest — release-please-action v4.4.0 (HIGH)
- https://github.com/golangci/golangci-lint/releases/latest — golangci-lint v2.11.4 (HIGH)

---
*Stack research for: dependency update CLI*
*Researched: 2026-03-31*
