# AGENTS.md

## Project

`pupdate` is a Go CLI whose shell hooks detect dependency manifests on directory changes, run safe package-manager commands, and cache successful results in local state. The hook path must remain low-latency and non-blocking.

## Toolchain And Commands

- Go 1.26 is pinned by both `go.mod` and `mise.toml`; run `mise install` after checkout and use `mise exec --` for Go commands.
- Build: `mise exec -- go build ./cmd/pupdate`
- Test all packages: `mise exec -- go test ./... -count=1`
- Test one package: `mise exec -- go test ./cmd/pupdate -count=1`
- Test one case: `mise exec -- go test ./cmd/pupdate -run '^TestExecuteCmdReturnsZeroWithoutPrintingOnSuccess$' -count=1`
- Vet: `mise exec -- go vet ./...`
- Format Go code: `mise exec -- gofmt -w ./cmd ./internal`
- Run the CI hot-path benchmarks: `mise exec -- go test ./internal/detection ./internal/freshness -run '^$' -bench 'Benchmark(DetectProjectTree|HashMatchedFiles)$' -count=1`
- CI runs vet, the uncached full suite, then hot-path benchmarks on Linux and macOS; mirror that order for hook, detection, freshness, state, or cross-platform changes.

## Code Map

- `cmd/pupdate/` is the `main` package and owns Cobra commands, config/preflight handling, install orchestration, shell snippets, async hook locking, and state coordination. The executable entrypoint is `cmd/pupdate/root.go`.
- `internal/detection/` scans the repo root, depth-1 directories, direct children of `packages/`, and configured `workspace_globs`; it applies `.gitignore` and exact-name `folder_blacklist` rules.
- `internal/freshness/` hashes matched inputs, reuses file metadata on Linux/macOS, and checks git-submodule drift. Changes here must preserve the performance guardrail in `performance_guardrail_test.go`.
- `internal/state/` owns schema-versioned JSON in `.pupdate`; saves are atomic. Invalid or incompatible state is intentionally treated as empty with a warning.
- `docs/architecture.md` documents the execution and async-hook flows; `docs/development.md` and `.github/workflows/ci.yml` are the command sources of truth.

## Behavioral Constraints

- Resolve package-manager executables from the current process `PATH`; hardcoded paths break `nvm`, `asdf`, and `mise` users.
- Keep hook execution silent on no-op `--quiet` runs, concise on stderr when work occurs, and async by default. Preserve bash, zsh, and fish snippet behavior.
- Keep existing safe install defaults: frozen/locked modes, lifecycle scripts disabled where supported, and Python installs gated by `--allow-scripts`.
- Do not broaden traversal casually; scan depth, `workspace_globs`, `.gitignore`, `folder_blacklist`, and the repo-wide `.pupignore` serve distinct behavior.
- `.pupdate` and `.pupdate.hook.lock` are generated per-repository machine state. Never commit or use them as fixtures outside temporary test directories.
- Use `t.TempDir()` for filesystem tests and shared command helpers from `cmd/pupdate/test_env_test.go` where applicable.
- Update `README.md` and relevant docs when flags, shell setup, output, supported managers, or config behavior changes.

## Repository Gotchas

- `internal/detection/matrix.go` combines `git-pkgs/manifests` identification with local Kasetto fallbacks; adding an ecosystem usually also requires install-plan and README changes.
- Install commands are constructed through embedded `git-pkgs/managers` definitions in `cmd/pupdate/run_install.go`; retain non-interactive and safety flags when changing plans.
- Do not edit the generated block between `BEGIN genignore` and `END genignore` in `.gitignore` manually.
- Release Please owns semver release PRs/tags and `CHANGELOG.md`; GoReleaser publishes Linux/macOS amd64/arm64 archives and the Homebrew cask. Use conventional commit prefixes so release classification works; workflow sources are `.github/workflows/release.yaml`, `release-please-config.json`, and `.goreleaser.yaml`.
