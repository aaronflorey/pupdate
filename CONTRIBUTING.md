# Contributing

## Local Setup

- Install [mise](https://mise.jdx.dev/) and run `mise install` to install the Go version pinned by `mise.toml` and `go.mod`.
- Clone the repo and work from a feature branch.

## Common Commands

```bash
mise exec -- gofmt -w ./cmd ./internal
mise exec -- go mod tidy
mise exec -- go vet ./...
mise exec -- go test ./... -count=1
mise exec -- go test ./internal/detection ./internal/freshness -run '^$' -bench 'Benchmark(DetectProjectTree|HashMatchedFiles)$' -count=1
```

Run a focused test while refactoring:

```bash
mise exec -- go test ./cmd/pupdate -count=1
```

Before opening a pull request, mirror CI order: confirm `gofmt -l ./cmd ./internal` and `go mod tidy -diff` produce no output, then run vet, the uncached full suite, and the hot-path benchmarks.

## Scope

- Keep changes small and focused.
- Preserve the low-latency `cd` hook path.
- Resolve package manager binaries from the current `PATH`.
- Avoid changing command behavior unless the change is intentional and covered by tests or docs.

## CLI Changes

- Keep stdout machine-readable when a command already uses structured output.
- Keep stderr concise so hook-driven runs stay readable.
- Prefer non-interactive flags and safe defaults.

## Pull Requests

- Update tests when behavior changes.
- Update `README.md` when flags, output, or setup steps change.
- Call out any cross-platform or shell-specific behavior in the PR description.
- Use conventional commits (`feat:`, `fix:`, `chore:`) so release automation can classify changes.

## Community

By participating, you agree to follow [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md).
