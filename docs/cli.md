# CLI Reference

pupdate is a Cobra-based CLI. The root command is `pupdate`; command definitions
live in `internal/cli/`, and `cmd/pupdate/main.go` is the thin executable entrypoint.

## `pupdate`

Root command. Prints version with `--version` / `-v`.

```bash
pupdate --version
```

## `pupdate run`

Detects supported ecosystems in the current directory and runs dependency
updates when needed. Emits concise human-readable status lines on stderr.

```bash
pupdate run [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--quiet` | bool | `false` | Suppress no-op output and child command output. |
| `--allow-scripts` | bool | `false` | Allow dependency manager lifecycle scripts where supported, and opt into Python installs that can execute install/build code. |
| `--dry-run` | bool | `false` | Show what would run without executing installs or saving state. |

### Environment variables

| Variable | Effect |
|----------|--------|
| `PUPDATE_SKIP_INSTALL=1` | Run detection only. No installs or `.pupdate` writes occur. Accepted values: `1`, `true`, `yes` (case-insensitive). |
| `PUPDATE_CONFIG=/path/to/config.yaml` | Override the config file path for this invocation. |

### Output format

Status lines are printed to **stderr** with the `pupdate:` prefix:

| Line | Meaning |
|------|---------|
| `pupdate: skip repo ($HOME)` | Skipped because the working directory is `$HOME`. |
| `pupdate: skip repo (outside configured root_directories)` | Skipped because the working directory is not a direct child of a configured root. |
| `pupdate: skip repo (.pupignore)` | Skipped because `.pupignore` is present. |
| `pupdate: skip repo (another run is active)` | A hook worker skipped because another repository operation holds the advisory lock. |
| `pupdate: installs disabled via PUPDATE_SKIP_INSTALL` | Installs are disabled via the environment variable. |
| `pupdate: --dry-run (no installs, no state changes)` | Dry-run mode is active. |
| `pupdate: skip <target> (<reason>)` | A specific ecosystem was detected but its install was skipped. |
| `pupdate: run <manager> <args> (in <subdir>)` | An install command is starting. `(in <subdir>)` is omitted for the root directory. |
| `pupdate: done <manager> (in <subdir>)` | An install command completed successfully. |
| `pupdate: dry-run <manager> <args> (in <subdir>)` | Dry-run preview of the install command. |
| `pupdate: error: dependency installs failed: ...` | One or more managers failed after independent targets finished and successful state was saved. |
| `pupdate: error git submodule status failed: <err>` | Git submodule status check failed (surfaced as error, does not crash). |

The `<target>` in skip lines is the ecosystem name for the root directory (e.g. `go`) or `<ecosystem>:<directory>` for subdirectories (e.g. `node:apps/web`).

### `--quiet` behavior

Successful quiet hook workers are fully silent, including successful updates.
Failures still produce one final error. Manual `run --quiet` continues to show
its routine `run` and `done` lines while suppressing skip output.

### Config defaults

The `quiet` and `allow_scripts` config keys set defaults for `run` without
needing flags. Explicit command-line flags always win. See
[Configuration](configuration.md#quiet) and
[Configuration](configuration.md#allow_scripts).

### Examples

```bash
# Basic run
pupdate run

# Quiet run (for shell hooks)
pupdate run --quiet

# Dry run to preview commands
pupdate run --dry-run

# Allow lifecycle scripts and Python installs
pupdate run --allow-scripts

# Disable installs, just see detection
PUPDATE_SKIP_INSTALL=1 pupdate run
```

## `pupdate init`

Prints the shell hook snippet for `bash`, `zsh`, or `fish`.

```bash
pupdate init [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--shell` | string | `""` (auto-detect from `$SHELL`) | Shell to configure: `bash`, `zsh`, or `fish`. Falls back to `bash` if `$SHELL` is empty or not one of those supported shells. |
| `--mode` | string | `async` | Hook execution mode: `async` (background, default) or `foreground` (synchronous). |

### Usage

```bash
# Auto-detect shell, async mode (default)
eval "$(pupdate init)"

# Explicit shell
eval "$(pupdate init --shell bash)"
eval "$(pupdate init --shell zsh)"
eval "$(pupdate init --shell fish)"

# Foreground (synchronous) mode
eval "$(pupdate init --shell bash --mode foreground)"
```

### What the hook does

The generated hook runs the resolved `pupdate` executable path with
`hook --quiet` on directory changes. In async mode (default), it adds `--async`
so the update runs in the background without blocking the shell.

The hook skips launching from `$HOME`. Successful quiet hook workers emit no
status or manager output; failures emit one final error.

### Persistent startup-file management

```bash
# Preview, then apply
pupdate init install --shell bash
pupdate init install --shell bash --mode foreground --apply

# Preview removal, then apply
pupdate init uninstall --shell bash
pupdate init uninstall --shell bash --apply
```

Both subcommands accept `--target <path>` and preview without writing unless
`--apply` is present. Install also accepts `--mode async|foreground`. Default
targets are `~/.bashrc`, `$ZDOTDIR/.zshrc` or `~/.zshrc`, and
`$XDG_CONFIG_HOME/fish/config.fish` or `~/.config/fish/config.fish`.

Exactly one marked pupdate block is managed. Duplicate, nested, or unmatched
markers are errors. Existing symlinks are resolved; applied writes are atomic,
preserve permissions, and create `<target>.pupdate.bak` once without ever
overwriting or automatically restoring it.

### Shell-specific behavior

- **bash:** Uses `PROMPT_COMMAND` to trigger the hook on each prompt.
- **zsh:** Uses `add-zsh-hook` with `chpwd` and `precmd` hooks.
- **fish:** Uses `--on-variable PWD` to trigger on directory change.

## `pupdate hook`

**Hidden command.** Runs the internal shell-hook flow. This is invoked by the
generated shell snippets, not directly by users.

```bash
pupdate hook [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--quiet` | bool | `false` | Suppress no-op output and child command output. |
| `--async` | bool | `false` | Launch the hook work in the background. |
| `--child` | bool | `false` | Run as the detached background hook worker. |
| `--lock-file` | string | `""` | Accepted only for compatibility with older child invocations; contents are ignored. |

The async parent only launches a child. The child acquires the same nonblocking
repository advisory lock as manual `run` and `reset`. Manual contention exits
1; hook contention is a successful skip and is silent with `--quiet`.

## `pupdate config`

Prints the resolved user config path and active config values. If the config
file is missing, it reports the missing path and the implicit default values
without creating `config.yaml`.

```bash
pupdate config
```

### Output fields

| Field | Description |
|-------|-------------|
| `path` | Resolved config file path. |
| `exists` | Whether the config file exists. |
| `root_directories` | Configured `root_directories` values (raw). |
| `root_directories_resolved` | Resolved `root_directories` values after `~` expansion and path normalization. |
| `workspace_globs` | Configured `workspace_globs` values (raw). |
| `workspace_globs_resolved` | Resolved `workspace_globs` values. |
| `folder_blacklist` | Configured `folder_blacklist` values (raw). |
| `folder_blacklist_resolved` | Resolved `folder_blacklist` values. |
| `quiet` | Configured `quiet` value, or `(not set)`. |
| `allow_scripts` | Configured `allow_scripts` value, or `(not set)`. |
| `project_path`, `project_exists` | Root-local `.pupdate.yaml` path and existence. |
| `project_workspace_globs`, `project_folder_blacklist` | Raw project traversal additions. |
| `project_*_resolved` | Normalized project traversal additions. |
| `effective_workspace_globs`, `effective_folder_blacklist` | Stable user-first merged traversal values. |

## `pupdate reset`

Deletes the `.pupdate` state file in the current directory. The next
`pupdate run` will treat all dependencies as stale and re-evaluate from
scratch.

```bash
pupdate reset
```

### Output

- If the file existed: `removed <path>`
- If the file did not exist: `no .pupdate file found; nothing to reset`

## `pupdate status`

Prints a read-only diagnostic snapshot for the current directory so you can
see whether `pupdate run` would skip, wait, or be blocked before it actually
tries to execute anything.

```bash
pupdate status
```

### Output includes

- Overall `run_status` and `run_reason` for the current directory.
- User and project config paths, existence, raw/resolved traversal values, and effective merged values.
- Configured and effective `quiet` / `allow_scripts` run defaults.
- Background hook lock path and whether a detached hook run is currently active.
- `.pupdate` path, existence, and any state decode warnings.
- One section per detected target with matched files, freshness result, `last_success_at`, install readiness, and resolved manager path.
- `run_status: disabled` and per-target `install_status: disabled` when `PUPDATE_SKIP_INSTALL` is active.

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | Command completed successfully (including skips and no-ops). |
| `1` | Command failed, a manual operation contended on the lock, or one or more managers failed. The error is printed once as `pupdate: error: <err>`. |

SIGINT or SIGTERM cancels the active manager, stops later targets, saves any
earlier successful outcomes, releases the repository lock, and exits 1 with
`pupdate: error: operation canceled`. On Linux and macOS, cancellation
terminates the manager's process group so descendants cannot retain the lock.

Repository-derived strings containing control characters, DEL, or terminal
escapes are rendered using quoted ASCII escapes in all command output.
