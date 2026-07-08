# Configuration

pupdate reads configuration from a YAML file and environment variables. This
document covers every config source, default, precedence rule, and validation
behavior.

## Configuration sources and precedence

| Priority | Source | Description |
|----------|--------|-------------|
| 1 (highest) | `PUPDATE_CONFIG` env var | Override the config file path for a single invocation. |
| 2 | `~/.config/pupdate/config.yaml` | Default user config file. |
| 3 (lowest) | Implicit defaults | Used when no config file exists or a key is absent. |

### Config path resolution

The config path is resolved in `cmd/pupdate/config.go` (`resolveUserConfigPath`):

1. If `PUPDATE_CONFIG` is set and non-empty, its value is resolved to an absolute path (relative to the current working directory) and used.
2. Otherwise, the default config directory is resolved:
   - **macOS:** `$XDG_CONFIG_HOME` if set, otherwise `~/.config`.
   - **Linux/other:** `os.UserConfigDir()` (typically `$XDG_CONFIG_HOME` or `~/.config`).
3. The config file path is `<config-dir>/pupdate/config.yaml`.

When the config file is missing, `pupdate run` and `pupdate config` use implicit defaults without creating the config directory or writing `config.yaml`.

## Config file format

The config file is YAML. The struct is defined in `cmd/pupdate/config.go` (`userConfig`):

```yaml
# Restrict runs to top-level project directories under these roots.
# ~ expands to the user's home directory.
# Example: ~/code
root_directories:
  - ~/code

# Repo-relative directory globs to scan in addition to the default shallow scan.
# Example: apps/* or services/*
workspace_globs:
  - apps/*
  - services/*

# Exact directory names to skip during traversal (not globs, not paths).
# Example: vendor, node_modules
folder_blacklist:
  - vendor
  - node_modules

# Set the default quiet mode for `run` without needing --quiet in shell aliases.
quiet: true

# Set the default lifecycle-script policy for `run`.
# Explicit command flags still win over this setting.
allow_scripts: false
```

## Config keys

### `root_directories`

- **Type:** list of strings (paths)
- **Default:** `[]` (empty — no restriction)
- **`~` expansion:** Yes, expands to the user's home directory.
- **Behavior:** When set, `pupdate run` only executes when the current working directory is a **direct child** of one of the configured roots. If the working directory is not a direct child of any root, the run is skipped with: `pupdate: skip repo (outside configured root_directories)`.
- **Validation:** Each entry is resolved to an absolute path. Invalid paths produce an error.
- **Consumed in:** `cmd/pupdate/preflight.go` (`collectPreflightSkipReason` → `isOutsideConfiguredRootDirectories`), `cmd/pupdate/run_execution.go`.

### `workspace_globs`

- **Type:** list of strings (glob patterns)
- **Default:** `[]` (empty — no additional workspace scanning)
- **Behavior:** Each glob is matched relative to the repository root. Matching directories are added to the scan. pupdate does **not** continue scanning deeper nested directories under each match unless another configured glob matches them.
- **Validation rules** (`normalizeWorkspaceGlob` in `config.go`):
  - Must be a valid `filepath.Match` glob pattern.
  - Must be relative (absolute paths are rejected).
  - Must not match the repository root (`.` is rejected).
  - Must not escape the repository root (`..` is rejected).
- **Consumed in:** `internal/detection/detector.go` (`scanWorkspaceGlob`).

Example for a monorepo:

```yaml
workspace_globs:
  - apps/*
  - services/*
```

### `folder_blacklist`

- **Type:** list of strings (exact directory names)
- **Default:** `[]` (empty — no blacklisting)
- **Behavior:** Any directory whose exact name matches an entry is skipped during traversal, both in the default shallow scan and while expanding `workspace_globs`.
- **Validation rules** (`normalizeFolderBlacklistEntry` in `config.go`):
  - Must be an exact directory name — **not** a glob (no `*`, `?`, `[`, `]`).
  - Must be a single name — **not** a path (no `/` or `\`).
  - Must not be `.` or `..`.
- **Consumed in:** `internal/detection/detector.go` (`shouldSkipDirectory`).

Example:

```yaml
workspace_globs:
  - apps/*
  - foo/*/*
folder_blacklist:
  - blah
  - vendor
```

### `quiet`

- **Type:** boolean (pointer — `nil` means "not set")
- **Default:** `nil` (not set — falls back to the `--quiet` flag default of `false`)
- **Behavior:** Sets the default quiet mode for `run`. When `true`, no-op output and child command output are suppressed. An explicit `--quiet` or `--quiet=false` flag on the command line always wins.
- **Consumed in:** `cmd/pupdate/run_execution.go` (`resolveRunOptions`).

### `allow_scripts`

- **Type:** boolean (pointer — `nil` means "not set")
- **Default:** `nil` (not set — falls back to the `--allow-scripts` flag default of `false`)
- **Behavior:** Sets the default lifecycle-script policy for `run`. When `true`, dependency manager lifecycle scripts are allowed where supported, and Python installs are opted in. An explicit `--allow-scripts` flag on the command line always wins.
- **Consumed in:** `cmd/pupdate/run_execution.go` (`resolveRunOptions`).

## Environment variables

### `PUPDATE_CONFIG`

- **Purpose:** Override the config file path for a single invocation.
- **Resolution:** Relative paths resolve from the current working directory to an absolute path.
- **Consumed in:** `cmd/pupdate/config.go` (`resolveUserConfigPath`).

```bash
PUPDATE_CONFIG=/path/to/config.yaml pupdate run
```

### `PUPDATE_SKIP_INSTALL`

- **Purpose:** Disable installs while still running detection and status flow.
- **Accepted values:** `1`, `true`, `yes` (case-insensitive). Any other value is treated as disabled.
- **Behavior:** When active, pupdate detects ecosystems and prints status but does not execute any install commands or save state.
- **Output:** `pupdate: installs disabled via PUPDATE_SKIP_INSTALL`
- **Consumed in:** `cmd/pupdate/run_execution.go` (`isInstallDisabled`).

```bash
PUPDATE_SKIP_INSTALL=1 pupdate run
```

## `.pupignore`

- **Purpose:** Short-circuit the entire pupdate run for a repository before detection, freshness checks, and installs.
- **Location:** A file named `.pupignore` in the repository root.
- **Behavior:** When present (as a file, not a directory), the run is skipped immediately with: `pupdate: skip repo (.pupignore)`.
- **Consumed in:** `cmd/pupdate/run_execution.go` (`hasPupIgnore`), `cmd/pupdate/preflight.go`.

## State file: `.pupdate`

- **Location:** `.pupdate` in the current working directory (the repository root).
- **Format:** JSON, schema version 1.
- **Schema** (defined in `internal/state/model.go`):

```json
{
  "version": 1,
  "ecosystems": {
    "go": {
      "last_success_at": "2026-06-23T05:53:42Z",
      "lockfiles": {
        "go.mod": "<sha256-hex>"
      },
      "lockfile_metadata": {
        "go.mod": {
          "size": 823,
          "mod_time_unix_nano": 1781848327321647000,
          "mode": "-rw-rw-r--",
          "file_id": "16777235:3537193704",
          "change_time_unix_nano": 1781848332820494089
        }
      }
    }
  }
}
```

- **State keys:** Per ecosystem and directory. The root directory uses the ecosystem name (e.g. `go`). Subdirectories use `<ecosystem>@<directory>` (e.g. `node@apps/web`).
- **`file_id` and `change_time_unix_nano`:** Only populated on Linux and macOS. These enable a shortcut that skips rehashing unchanged lockfiles. On other platforms, these fields are empty and lockfiles are rehashed on every run.
- **Validation:** If the state file is invalid JSON, it is treated as empty with a warning: `state file is invalid; treating as empty`. If the schema version does not match, it is treated as empty with a warning about the version mismatch.
- **Atomic writes:** State is saved via a temp file + atomic rename, with a parent directory `fsync` on Unix (`internal/state/store.go`).

## Background hook lock: `.pupdate.hook.lock`

- **Location:** `.pupdate.hook.lock` in the current working directory.
- **Format:** Plain text. First line is the claimed-at Unix timestamp. Optional second line is `pid=<pid>`.
- **Purpose:** Prevents overlapping async background hook runs in the same repository.
- **Stale after:** 10 minutes (`backgroundHookStaleAfter` in `cmd/pupdate/hook.go`).
- **Stale detection:** If the lock file's PID is no longer running, or the lock is older than 10 minutes, or the file is empty, it is considered stale and removed.

## Validation and failure modes

| Scenario | Behavior |
|----------|----------|
| Config file missing | Uses implicit defaults; no error. |
| Config file has invalid YAML | Error: `failed to parse <path>: <err>`. |
| `root_directories` entry has invalid path | Error: `failed to resolve root_directories[N]: <err>`. |
| `workspace_globs` entry is absolute | Error: `workspace glob must be relative`. |
| `workspace_globs` entry matches root | Error: `workspace glob must not match the repository root`. |
| `workspace_globs` entry escapes root | Error: `workspace glob must stay within the repository root`. |
| `workspace_globs` entry is invalid glob | Error: `invalid glob pattern: <err>`. |
| `folder_blacklist` entry contains glob chars | Error: `folder blacklist entry must be an exact directory name, not a glob`. |
| `folder_blacklist` entry is a path | Error: `folder blacklist entry must be an exact directory name, not a path`. |
| State file invalid JSON | Treated as empty; warning: `state file is invalid; treating as empty`. |
| State schema version mismatch | Treated as empty; warning about version mismatch. |

## Checking your config

```bash
pupdate config
```

This prints the resolved config path, whether the file exists, and all resolved values:

```
path: /home/user/.config/pupdate/config.yaml
exists: true
root_directories: ~/code
root_directories_resolved: /home/user/code
workspace_globs: apps/*, services/*
workspace_globs_resolved: apps/*, services/*
folder_blacklist: vendor
folder_blacklist_resolved: vendor
quiet: (not set)
allow_scripts: (not set)
```
