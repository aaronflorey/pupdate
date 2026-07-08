# Deployment

This document covers build artifacts, CI, and the release pipeline.

## Build artifacts

GoReleaser builds pupdate for the following targets (defined in `.goreleaser.yaml`):

| OS | Architecture |
|----|--------------|
| Linux | amd64 |
| Linux | arm64 |
| macOS | amd64 |
| macOS | arm64 |

Build settings:

- `CGO_ENABLED=0` — static binary, no C dependencies.
- Entry point: `./cmd/pupdate`.
- ldflags inject version, commit, date, and `builtBy=goreleaser`.
- Archives are `tar.gz` format.

## CI pipeline

CI is defined in `.github/workflows/ci.yml` and runs on pull requests and pushes
to `main`.

### Jobs

| Job | Matrix | Steps |
|-----|--------|-------|
| `test` | `ubuntu-latest`, `macos-latest` | `go vet ./...`, `go test ./... -count=1`, hot-path benchmarks. |

CI uses `actions/setup-go@v6` with `go-version-file: go.mod` for Go version
management.

## Release pipeline

The release pipeline is defined in `.github/workflows/release.yaml` and uses
two stages: release-please and GoReleaser.

### Stage 1: release-please

- **Trigger:** Push to `main` or `master`.
- **Config:** `release-please-config.json` (release type: `go`, package name: `pupdate`, changelog: `CHANGELOG.md`).
- **Manifest:** `.release-please-manifest.json` (current version: `0.7.0`).
- **Behavior:** Maintains a release PR that accumulates conventional commits. When merged, it creates a tagged release and updates `CHANGELOG.md`.
- **Output:** `release_created` and `tag_name` outputs for the GoReleaser job.

### Stage 2: GoReleaser

- **Trigger:** release-please creates a release (`release_created == 'true'`).
- **Checkout:** Checks out the created tag.
- **Action:** `goreleaser/goreleaser-action@v7` with `args: release --clean`.
- **Secrets:** `GITHUB_TOKEN` (for GitHub release assets) and `HOMEBREW_TAP_GITHUB_TOKEN` (for Homebrew tap updates).

### Homebrew tap

GoReleaser publishes a cask to `aaronflorey/homebrew-tap`:

- **Cask name:** `pupdate`
- **Repository:** `aaronflorey/homebrew-tap` (branch: `main`)
- **Homepage:** `https://github.com/aaronflorey/pupdate`
- **Description:** `Auto-update project dependencies on directory entry`

Users install via:

```bash
brew install --cask aaronflorey/tap/pupdate
```

## Release versioning

- **Type:** Semantic versioning with `v` prefix (e.g. `v0.7.0`).
- **Config:** `include-v-in-tag: true`, `include-component-in-tag: false`.
- **Changelog:** `CHANGELOG.md` is maintained by release-please.
- **Conventional commits:** Required for release-please to classify changes (`feat:`, `fix:`, `chore:`, etc.). `docs:` and `test:` commits are excluded from the changelog.

## Dependabot

`.github/dependabot.yml` configures weekly dependency updates for:

- GitHub Actions (`github-actions` ecosystem).
- Go modules (`gomod` ecosystem).

Both have `open-pull-requests-limit: 5` and run weekly.
