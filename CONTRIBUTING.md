# Contributing

## Local Setup

- Install Go 1.26 or the version declared in `go.mod`.
- Clone the repo and work from a feature branch.

## Common Commands

```bash
go test ./...
gofmt -w ./cmd ./internal
```

Run a focused test while refactoring:

```bash
go test ./cmd/pupdate -count=1
```

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
