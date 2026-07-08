# pupdate Documentation

`pupdate` is a fast Go CLI that keeps project dependencies up to date when you
enter a repository. It detects dependency ecosystems in the current directory,
runs the matching package manager when work is needed, and skips unnecessary
installs when dependency files have not changed.

It is built for shell-hook usage on `cd`, so the common path stays low-latency,
safe by default, and easy to follow from concise stderr status output.

## What this documentation covers

- **Getting started** — install, set up the shell hook, and verify it works.
- **Configuration** — user config file, environment variables, defaults, and precedence.
- **CLI reference** — every command and flag with exact behavior.
- **Architecture** — internal packages, detection, freshness, and state model.
- **Troubleshooting** — common symptoms, causes, and fixes.
- **Development** — local setup, testing, linting, and contribution workflow.
- **Deployment / releases** — CI, release-please, and GoReleaser pipeline.

## Quick navigation

| Document | Purpose |
|----------|---------|
| [Getting Started](getting-started.md) | Install pupdate and run it for the first time. |
| [Configuration](configuration.md) | Config file, environment variables, defaults, and precedence. |
| [CLI Reference](cli.md) | Commands, flags, arguments, and output formats. |
| [Architecture](architecture.md) | Internal packages, detection, freshness engine, and state model. |
| [Troubleshooting](troubleshooting.md) | Symptoms, causes, diagnostic commands, and fixes. |
| [Development](development.md) | Repository layout, local workflow, testing, and conventions. |
| [Deployment](deployment.md) | Build artifacts, CI, and release pipeline. |

## Recommended reading order

1. **New users:** [Getting Started](getting-started.md) → [Configuration](configuration.md) → [CLI Reference](cli.md)
2. **Contributors:** [Development](development.md) → [Architecture](architecture.md) → [Deployment](deployment.md)
3. **Operators:** [Troubleshooting](troubleshooting.md) → [Configuration](configuration.md)

## Supported ecosystems

| Ecosystem | Detected By | Manager | Default Behavior |
|-----------|-------------|---------|------------------|
| Node | `bun.lock` | `bun` | `bun install --frozen-lockfile --ignore-scripts` |
| Node | `package-lock.json` | `npm` | `npm ci --ignore-scripts` |
| Node | `pnpm-lock.yaml` | `pnpm` | `pnpm install --frozen-lockfile --ignore-scripts` |
| Node | `yarn.lock` | `yarn` | `yarn install --frozen-lockfile --ignore-scripts` |
| PHP | `composer.lock` | `composer` | `composer install --no-interaction --prefer-dist --no-scripts` |
| Python | `uv.lock` | `uv` | skipped by default; `--allow-scripts` opts into `uv sync --frozen` |
| Python | `poetry.lock` | `poetry` | skipped by default; `--allow-scripts` opts into `poetry install --no-interaction --sync` |
| Python | `requirements.txt` | `pip` | skipped by default; `--allow-scripts` opts into `pip install -r requirements.txt --disable-pip-version-check --no-input` |
| Kasetto | `kasetto.lock`, `kasetto.yaml`, `kasetto.yml` | `kst` | `kst sync --project --config <local-config>` |
| Go | `go.mod` | `go` | `go mod download` |
| Rust | `cargo.lock` | `cargo` | `cargo fetch --locked` |
| Git submodules | `.gitmodules` | `git` | `git submodule update --init --recursive` |

Manager binaries are resolved from the current process `PATH`, which keeps
`pupdate` compatible with tools like `nvm`, `asdf`, and `mise`.

## Known limitations

- **Supported platforms:** Linux and macOS are the supported release platforms. Other operating systems are best-effort only and are not shipped as release binaries.
- **File-identity optimization:** The shortcut that skips rehashing unchanged lockfiles is only implemented on Linux and macOS (`internal/freshness/file_identity_linux.go`, `internal/freshness/file_identity_darwin.go`). On other platforms, freshness checks may rehash unchanged lockfiles more often.
- **Python safety:** Python managers (`uv`, `poetry`, `pip`) are skipped by default because they can execute install/build code. You must explicitly pass `--allow-scripts` to opt in.
- **Node multiple lockfiles:** If more than one Node lockfile is detected in the same directory, the install is skipped with a warning.
- **Kasetto lock-only:** Kasetto installs only run when a local `kasetto.yaml` or `kasetto.yml` is detected. Lock-only detections are skipped.
