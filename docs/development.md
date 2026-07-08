# Development

This document covers local setup, repository layout, testing, linting, and
contribution workflow.

## Prerequisites

- **Go 1.26** or later. The version is declared in `go.mod` and `mise.toml`.
- If you use [mise](https://mise.jdx.dev/), the Go version is managed automatically:
  ```bash
  mise install
  ```

## Local setup

```bash
git clone https://github.com/aaronflorey/pupdate.git
cd pupdate
go build ./cmd/pupdate
```

The binary is output to the repository root as `pupdate`.

## Common commands

### Build

```bash
go build ./cmd/pupdate
```

### Run tests

```bash
go test ./...
```

Run a focused test package:

```bash
go test ./cmd/pupdate -count=1
```

### Vet

```bash
go vet ./...
```

### Format

```bash
gofmt -w ./cmd ./internal
```

### Benchmarks

The CI runs hot-path benchmarks for detection and freshness:

```bash
go test ./internal/detection ./internal/freshness -run '^$' -bench 'Benchmark(DetectProjectTree|HashMatchedFiles)$' -count=1
```

## Repository layout

```
cmd/pupdate/          # CLI commands and orchestration
internal/
  detection/          # Ecosystem detection
  freshness/          # Freshness evaluation
  state/              # State file model and storage
docs/                 # This documentation
.github/workflows/    # CI and release pipelines
.goreleaser.yaml      # GoReleaser build/release config
release-please-config.json
.release-please-manifest.json
```

See [Architecture](architecture.md) for internal package details.

## Coding conventions

From `CONTRIBUTING.md`:

- Keep changes small and focused.
- Preserve the low-latency `cd` hook path.
- Resolve package manager binaries from the current `PATH`.
- Avoid changing command behavior unless the change is intentional and covered by tests or docs.
- Keep stdout machine-readable when a command already uses structured output.
- Keep stderr concise so hook-driven runs stay readable.
- Prefer non-interactive flags and safe defaults.
- Use conventional commits (`feat:`, `fix:`, `chore:`) so release automation can classify changes.

## Testing strategy

Tests are colocated with source files (`*_test.go`). The test suite covers:

- **Command behavior:** `cmd/pupdate/*_test.go` — CLI commands, flags, output.
- **Detection:** `internal/detection/detector_test.go` — ecosystem detection, scanning, blacklisting.
- **Freshness:** `internal/freshness/engine_test.go` — hash comparison, git submodule drift, file identity.
- **State:** `internal/state/model_test.go`, `internal/state/store_test.go` — encode/decode, atomic save.
- **Benchmarks:** `internal/detection/detector_benchmark_test.go`, `internal/freshness/engine_benchmark_test.go`, `internal/freshness/performance_guardrail_test.go`.

### Adding tests

- Place test files next to the source file they test.
- Use table-driven tests for command/flag combinations.
- Use `t.TempDir()` for filesystem isolation.
- The `cmd/pupdate/test_env_test.go` file provides shared test helpers.

## Pull requests

From `CONTRIBUTING.md`:

- Update tests when behavior changes.
- Update `README.md` when flags, output, or setup steps change.
- Call out any cross-platform or shell-specific behavior in the PR description.
- Use conventional commits so release automation can classify changes.

## Key dependencies

| Dependency | Purpose |
|------------|---------|
| `github.com/spf13/cobra` | CLI command framework. |
| `gopkg.in/yaml.v3` | Config file parsing. |
| `github.com/git-pkgs/manifests` | Ecosystem identification from lockfile/manifest names. |
| `github.com/git-pkgs/managers` | Install command construction for package managers. |
| `github.com/git-pkgs/gitignore` | `.gitignore` pattern matching during directory traversal. |

## mise configuration

`mise.toml` declares the Go version:

```toml
[tools]
go = "1.26"
```

If you use mise, run `mise install` to install the correct Go version automatically.
