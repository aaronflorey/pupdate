# pupdate

[![CI](https://github.com/aaronflorey/pupdate/actions/workflows/ci.yml/badge.svg)](https://github.com/aaronflorey/pupdate/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/aaronflorey/pupdate?sort=semver)](https://github.com/aaronflorey/pupdate/releases)
[![License](https://img.shields.io/github/license/aaronflorey/pupdate)](LICENSE)

`pupdate` is a fast Go CLI that keeps project dependencies up to date when you
enter a repository. It detects dependency ecosystems in the current directory,
other depth-1 subdirectories, and direct children of `packages/`, runs the
matching package manager when work is needed, and skips unnecessary installs
when dependency files have not changed.

It is built for shell-hook usage on `cd`, so the common path stays low-latency,
safe by default, and easy to follow from concise stderr status output.

Detection is backed by `github.com/git-pkgs/manifests` plus small local
fallbacks for unsupported ecosystems, and install command construction is backed
by `github.com/git-pkgs/managers` with the same approach. `pupdate` preserves
its existing traversal limits, local state flow, and safe default behavior.

## Install

```bash
bin install github.com/aaronflorey/pupdate
```

Requires [`bin`](https://github.com/aaronflorey/bin) to be installed and available on `PATH`.

## Project Docs

- Contribution guide: [`CONTRIBUTING.md`](CONTRIBUTING.md)
- Code of conduct: [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md)
- Security policy: [`SECURITY.md`](SECURITY.md)

## Release and CI

- CI runs on pull requests and pushes to `main` across Linux and macOS.
- `release-please` maintains changelog and semver release PRs.
- GoReleaser publishes tagged release binaries and updates Homebrew tap artifacts.
- Release notes are published on the [GitHub Releases page](https://github.com/aaronflorey/pupdate/releases).

## Quick Start

Run manually in a project:

```bash
pupdate run
```

Install the shell hook for your current session:

```bash
eval "$(pupdate init --shell bash)"
```

```bash
eval "$(pupdate init --shell zsh)"
```

```bash
eval "$(pupdate init --shell fish)"
```

Use explicit foreground mode when you want synchronous hook execution:

```bash
eval "$(pupdate init --shell bash --mode foreground)"
```

Escaped form for docs or templates that need literal shell interpolation text:

```bash
eval "\$(pupdate init --shell bash)"
eval "\$(pupdate init --shell zsh)"
eval "\$(pupdate init --shell fish)"
```

The generated hooks run `pupdate hook --quiet` on directory changes, but skip
launching from `$HOME`. Quiet mode stays silent for no-op runs and only prints
status when an update actually executes.

By default, `pupdate init` generates hooks that run `pupdate hook --quiet --async`
in the background. Overlapping background runs in the same repo are skipped
while a recent `.pupdate.hook.lock` exists.

## What It Does

- detects supported dependency ecosystems from lockfiles and manifests
- resolves manager binaries from your current `PATH`
- skips work when dependency inputs are unchanged since the last successful run
- stores local state in `.pupdate`
- explains current detection, freshness, config, and PATH readiness with `pupdate status`
- respects `.pupignore` to skip the full run (detection, freshness checks, and installs)
- uses safe defaults and requires explicit opt-in for lifecycle scripts

## Supported Ecosystems

| Ecosystem | Detected By | Manager | Default Command |
|-----------|-------------|---------|-----------------|
| Node | `bun.lock` | `bun` | `bun install --frozen-lockfile --ignore-scripts` |
| Node | `package-lock.json` | `npm` | `npm ci --ignore-scripts` |
| Node | `pnpm-lock.yaml` | `pnpm` | `pnpm install --frozen-lockfile --ignore-scripts` |
| Node | `yarn.lock` | `yarn` | `yarn install --frozen-lockfile --ignore-scripts` |
| PHP | `composer.lock` | `composer` | `composer install --no-interaction --prefer-dist --no-scripts` |
| Python | `uv.lock` | `uv` | `uv sync --frozen` |
| Python | `poetry.lock` | `poetry` | `poetry install --no-interaction --sync` |
| Python | `requirements.txt` | `pip` | `pip install -r requirements.txt --disable-pip-version-check --no-input` |
| Kasetto | `kasetto.lock`, `kasetto.yaml`, `kasetto.yml` | `kst` | `kst sync --project --config <local-config>` |
| Go | `go.mod` | `go` | `go mod download` |
| Rust | `cargo.lock` | `cargo` | `cargo fetch --locked` |
| Git submodules | `.gitmodules` | `git` | `git submodule update --init --recursive` |

Manager binaries are resolved from the current process `PATH`, which keeps
`pupdate` compatible with tools like `nvm`, `asdf`, and `mise`.

Kasetto installs only run when a local `kasetto.yaml` or `kasetto.yml` is
detected. Lock-only Kasetto detections are skipped.

## Commands

### `pupdate run`

Detects supported ecosystems in the current directory and emits human-readable
status lines on stderr. The command skips the user's home directory and can be
restricted to top-level project directories inside configured roots.

Flags:

- `--quiet`: suppress no-op output and child command output
- `--allow-scripts`: allow dependency manager lifecycle scripts where supported

Environment:

- `PUPDATE_SKIP_INSTALL=1`: disable installs while still running detection and status flow

User config:

- `~/.config/pupdate/config.yaml`
- `root_directories:` with one or more paths (for example `~/code`): only run when the current working directory is a direct child of one of those roots; `~` expands to the user's home directory
- `quiet: true|false`: set the default `run` quiet mode without needing `--quiet` in shell aliases or wrappers
- `allow_scripts: true|false`: set the default lifecycle-script policy for `run`; explicit command flags still win
- when missing, `pupdate run` and `pupdate config` use implicit defaults (`root_directories: []`) without creating the config directory or writing `config.yaml`

### `pupdate init`

Prints the shell hook snippet for `bash`, `zsh`, or `fish`.

Flags:

- `--shell <bash|zsh|fish>`: choose the shell explicitly
- `--mode <foreground|async>`: choose synchronous foreground execution, or use the default async background hook that detaches update execution from the shell transition

### `pupdate config`

Prints the resolved user config path and active config values. If the config file
is missing, it reports the missing path and the implicit default values without
creating `config.yaml`.

Output includes:

- whether the config file currently exists
- the configured `root_directories` values from the file
- the resolved `root_directories` values after `~` expansion and path normalization
- the configured `quiet` and `allow_scripts` values when set

### `pupdate status`

Prints a read-only diagnostic snapshot for the current directory so you can see
whether `pupdate run` would skip, wait, or be blocked before it actually tries
to execute anything.

Output includes:

- overall `run_status` and `run_reason` for the current directory
- config path, existence, and resolved `root_directories`
- configured and effective `quiet` / `allow_scripts` run defaults
- background hook lock path and whether a detached hook run is currently active
- `.pupdate` path, existence, and any state decode warnings
- one section per detected target with matched files, freshness result, install readiness, and resolved manager path

## How It Works

1. `pupdate run` scans the current directory, other depth-1 subdirectories, and
   direct children of `packages/` for known lockfiles and manifests.
2. It detects the matching ecosystem and package manager.
3. It compares current file hashes against the local `.pupdate` state file using
   namespaced keys per ecosystem and directory.
4. If dependency inputs are unchanged since the last successful run, it skips work.
5. If inputs changed, or if no successful state exists yet, it runs the manager's
   safe default command.
6. On success, it updates `.pupdate` with the latest lockfile hashes and timestamp.

For git submodules, `pupdate` also checks `git submodule status --recursive` so it
can react to submodule drift even when `.gitmodules` itself has not changed.

## Status Output

Manual runs use stable, concise stderr prefixes:

- `pupdate: skip repo ($HOME)`
- `pupdate: skip repo (outside configured root_directories)`
- `pupdate: skip repo (.pupignore)`
- `pupdate: installs disabled via PUPDATE_SKIP_INSTALL`
- `pupdate: skip <target> (<reason>)`
- `pupdate: run <manager> <args> (in <subdir>)`
- `pupdate: done <manager> (in <subdir>)`
- `pupdate: error <manager> install failed: <err>`

`--quiet` suppresses skip and no-op lines so hook-driven runs stay silent unless
an update actually runs. Successful updates still print `run` and `done` lines,
and failed updates print an `error` line.

These messages are meant to stay readable in interactive shells and easy to grep
from shell history or logs.

## Safety Model

- Safe and frozen install modes are used by default where supported.
- Lifecycle scripts are disabled by default where supported.
- `--allow-scripts` is required to opt into lifecycle scripts.
- `.pupignore` short-circuits the full run for a repository before detection and state checks.
- `root_directories` can restrict runs to top-level project directories under specific roots.
- Hook-driven runs default to non-blocking async execution and avoid launching from `$HOME`.
- Async hook mode skips overlapping background runs per repo using `.pupdate.hook.lock`.
- Git submodule status failures are surfaced as stderr errors without crashing the command.
