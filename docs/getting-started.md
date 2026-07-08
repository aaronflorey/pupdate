# Getting Started

## Prerequisites

- **Go 1.26** or later (for building from source). The version is declared in `go.mod` and `mise.toml`.
- A supported package manager for each ecosystem you want pupdate to manage (`bun`, `npm`, `pnpm`, `yarn`, `composer`, `uv`, `poetry`, `pip`, `kst`, `go`, `cargo`, `git`). These must be discoverable on your `PATH`.
- A supported shell: **bash**, **zsh**, or **fish**.
- Supported operating systems: **Linux** and **macOS**. Other platforms are best-effort.

## Installation

### Homebrew (recommended for macOS/Linux)

```bash
brew install --cask aaronflorey/tap/pupdate
```

### Go toolchain

```bash
go install github.com/aaronflorey/pupdate/cmd/pupdate@latest
```

### bin

Requires [`bin`](https://github.com/aaronflorey/bin) to be installed and available on `PATH`.

```bash
bin install github.com/aaronflorey/pupdate
```

## Verify the installation

```bash
pupdate --version
```

You should see a version string. If you built from source without GoReleaser ldflags, it will show `dev`.

## First run

Run pupdate manually in any project directory that has a lockfile or manifest:

```bash
cd ~/code/my-project
pupdate run
```

You will see status lines on stderr. For example, in a Go project:

```
pupdate: run go mod download
pupdate: done go
```

On the next run with no changes, it will skip:

```
pupdate: skip go (dependency lockfiles unchanged since last successful run)
```

## Set up the shell hook

The shell hook runs `pupdate` automatically when you `cd` into a project directory.

### bash

```bash
eval "$(pupdate init --shell bash)"
```

Add this to your `~/.bashrc` or `~/.bash_profile` to make it permanent.

### zsh

```bash
eval "$(pupdate init --shell zsh)"
```

Add this to your `~/.zshrc` to make it permanent.

### fish

```bash
eval "$(pupdate init --shell fish)"
```

Add this to your `~/.config/fish/config.fish` to make it permanent.

### Foreground mode (synchronous)

By default, the hook runs asynchronously in the background so it never blocks your shell. If you prefer synchronous execution:

```bash
eval "$(pupdate init --shell bash --mode foreground)"
```

## Verify the hook works

1. Open a new shell (or re-source your config).
2. `cd` into a project with a lockfile.
3. You should see `pupdate:` status lines on stderr, or nothing if the run is a quiet no-op.

To check the current state of detection, freshness, and config:

```bash
pupdate status
```

## Common next steps

- **Configure root directories** to restrict pupdate to specific parent folders. See [Configuration](configuration.md).
- **Add workspace globs** for monorepos. See [Configuration](configuration.md#workspace_globs).
- **Add `.pupignore`** to a repository to skip it entirely. See [Configuration](configuration.md#pupignore).
- **Check what pupdate sees** with `pupdate status` and `pupdate config`.
- **Reset state** if something seems stuck: `pupdate reset`.

## State files

pupdate creates two local state files in your project directory:

| File | Purpose |
|------|---------|
| `.pupdate` | JSON state file tracking lockfile hashes and last-success timestamps per ecosystem. |
| `.pupdate.hook.lock` | Lock file for async background hook runs; prevents overlapping runs in the same repo. |

Add both to your project's `.gitignore`:

```
.pupdate
.pupdate.hook.lock
```
