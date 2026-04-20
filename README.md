# pupdate

[![CI](https://github.com/aaronflorey/pupdate/actions/workflows/ci.yml/badge.svg)](https://github.com/aaronflorey/pupdate/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/aaronflorey/pupdate?sort=semver)](https://github.com/aaronflorey/pupdate/releases)
[![License](https://img.shields.io/github/license/aaronflorey/pupdate)](LICENSE)

`pupdate` is a fast Go CLI that keeps project dependencies up to date when you
enter a repository. It detects dependency ecosystems in the current directory and
depth-1 subdirectories, runs the matching package manager when work is needed,
and skips unnecessary installs when dependency files have not changed.

It is built for shell-hook usage on `cd`, so the common path stays low-latency,
safe by default, and easy to follow from concise stderr status output.

Detection is backed by `github.com/git-pkgs/manifests`, and install command
construction is backed by `github.com/git-pkgs/managers`, while `pupdate`
preserves its existing traversal limits, local state flow, and safe default
behavior.

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

- CI runs on pull requests and pushes to `main` across Linux, macOS, and Windows.
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

Escaped form for docs or templates that need literal shell interpolation text:

```bash
eval "\$(pupdate init --shell bash)"
eval "\$(pupdate init --shell zsh)"
eval "\$(pupdate init --shell fish)"
```

The generated hooks run `pupdate run --quiet` on directory changes, but skip
launching from `$HOME`. Quiet mode stays silent for no-op runs and only prints
status when an update actually executes.

## What It Does

- detects supported dependency ecosystems from lockfiles and manifests
- resolves manager binaries from your current `PATH`
- skips work when dependency inputs are unchanged since the last successful run
- stores local state in `.pupdate`
- respects `.pupignore` to disable automatic runs in a repository
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
| Go | `go.mod` | `go` | `go mod download` |
| Rust | `cargo.lock` | `cargo` | `cargo fetch --locked` |
| Git submodules | `.gitmodules` | `git` | `git submodule update --init --recursive` |

Manager binaries are resolved from the current process `PATH`, which keeps
`pupdate` compatible with tools like `nvm`, `asdf`, and `mise`.

## Commands

### `pupdate run`

Detects supported ecosystems in the current directory and emits human-readable
status lines on stderr. The command skips the user's home directory.

Flags:

- `--quiet`: suppress no-op output and child command output
- `--allow-scripts`: allow dependency manager lifecycle scripts where supported

Environment:

- `PUPDATE_SKIP_INSTALL=1`: disable installs while still running detection and status flow

### `pupdate init`

Prints the shell hook snippet for `bash`, `zsh`, or `fish`.

Flags:

- `--shell <bash|zsh|fish>`: choose the shell explicitly

## How It Works

1. `pupdate run` scans the current directory and depth-1 subdirectories for known
   lockfiles and manifests.
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
- `.pupignore` disables automatic runs for a repository.
- Hook-driven runs remain non-blocking and avoid launching from `$HOME`.
- Git submodule status failures are surfaced as stderr errors without crashing the command.
