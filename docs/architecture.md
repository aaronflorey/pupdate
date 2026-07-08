# Architecture

This document describes pupdate's internal structure, package boundaries, and
data flow. All paths refer to the repository root.

## Repository layout

```
cmd/pupdate/          # CLI entry point and all cobra commands
  root.go             # Root command, version, exit handling
  run.go              # `pupdate run` command definition
  run_execution.go    # Run orchestration, preflight, skip logic
  run_install.go      # Manager plan selection, install execution
  run_state.go        # State outcome application and persistence
  init.go             # `pupdate init` command
  init_shell.go       # Shell and mode resolution
  init_snippets.go    # Snippet generation dispatch
  init_snippet_bash.go
  init_snippet_zsh.go
  init_snippet_fish.go
  hook.go             # `pupdate hook` (hidden) — async background hook
  hook_process_unix.go
  hook_process_windows.go
  preflight.go        # Pre-run checks: home dir, root_directories, .pupignore
  config.go           # Config struct, path resolution, validation
  config_cmd.go       # `pupdate config` command
  status.go           # `pupdate status` command
  reset.go            # `pupdate reset` command
internal/
  detection/          # Ecosystem detection from lockfiles/manifests
    detector.go       # Directory scanning and signal matching
    matrix.go         # Signal-to-ecosystem mapping
    model.go          # Ecosystem types, DetectionResult
  freshness/          # Freshness evaluation (hash comparison)
    engine.go         # Core freshness logic, git submodule drift
    file_identity_darwin.go  # macOS file identity (inode, ctime)
    file_identity_linux.go   # Linux file identity (inode, ctime)
    file_identity_other.go   # No-op for unsupported platforms
  state/              # State file (.pupdate) model and storage
    model.go          # FileState struct, JSON encode/decode, schema
    store.go          # Atomic load/save with fsync
```

## Package boundaries

### `cmd/pupdate` (main package)

Contains all cobra commands and the orchestration logic. This package:

- Defines the CLI surface (`run`, `init`, `hook`, `config`, `status`, `reset`).
- Resolves user configuration.
- Runs preflight checks (home directory, root_directories, .pupignore).
- Coordinates detection, freshness evaluation, and install execution.
- Manages state persistence and background hook locking.

### `internal/detection`

Detects supported dependency ecosystems from lockfiles and manifests.

- **Entry point:** `DetectWithOptions(dir string, options Options) ([]DetectionResult, error)`
- **Scanning:** Scans the repository root (`.`), depth-1 subdirectories, and direct children of `packages/`. Optionally scans `workspace_globs` matches.
- **Signal matching:** Uses `github.com/git-pkgs/manifests` for ecosystem identification, with local fallbacks for Kasetto.
- **Gitignore support:** Reads `.gitignore` and skips matching directories during traversal.
- **Folder blacklist:** Skips directories whose exact name matches a `folder_blacklist` entry.
- **Output:** `[]DetectionResult`, each with `Ecosystem`, `Directory`, `Managers`, `MatchedFiles`, and `Warnings`.

### `internal/freshness`

Evaluates whether detected ecosystems need updates by comparing current lockfile
hashes against stored state.

- **Entry point:** `Evaluate(dir string, detections []detection.DetectionResult, current state.FileState) ([]EcosystemDecision, error)`
- **Hashing:** SHA-256 of matched lockfiles.
- **File-identity optimization:** On Linux and macOS, if a file's size, mtime, mode, inode (`file_id`), and ctime (`change_time_unix_nano`) are all unchanged, the stored hash is reused instead of rehashing. On other platforms, files are always rehashed.
- **Git submodules:** Runs `git submodule status --recursive` (2-second timeout) and checks for drift indicators (`-`, `+`, `U` prefixes).
- **PHP vendor check:** For PHP projects with legacy `vendor/.pupdate-checksum` state, checks if the `vendor/` directory exists; if missing, forces an update.
- **Output:** `[]EcosystemDecision`, each with `Decision` (`update` or `skip`), `Reason`, `Lockfiles`, and `LockfileMetadata`.

### `internal/state`

Manages the `.pupdate` state file.

- **Model:** `FileState` with `Version` (schema 1) and `Ecosystems` map.
- **State keys:** Ecosystem name for root (e.g. `go`), `<ecosystem>@<directory>` for subdirectories (e.g. `node@apps/web`).
- **Storage:** `Store` with `Load()` and `Save()`. Saves are atomic (temp file + rename + parent fsync on Unix).
- **Validation:** Invalid JSON is treated as empty with a warning. Schema version mismatches are treated as empty with a warning.

## Data flow

```
pupdate run
  │
  ├─ collectPreflight()
  │    ├─ resolveUserConfigPath()     → config path (PUPDATE_CONFIG or default)
  │    ├─ readUserConfig()            → raw config
  │    ├─ resolveUserConfig()         → resolved config (~ expansion, validation)
  │    ├─ collectPreflightSkipReason()
  │    │    ├─ isHomeDirectory()      → skip if $HOME
  │    │    ├─ isOutsideConfiguredRootDirectories()  → skip if outside roots
  │    │    └─ hasPupIgnore()         → skip if .pupignore present
  │    ├─ Store.Load()                → current state (or empty)
  │    ├─ detection.DetectWithOptions()  → []DetectionResult
  │    └─ freshness.Evaluate()        → []EcosystemDecision
  │
  ├─ resolveRunOptions()              → merge config defaults + flags
  │
  ├─ executeRunResults()
  │    └─ for each detection result:
  │         ├─ if DecisionSkip → print skip, record metadata refresh
  │         ├─ if DecisionUpdate:
  │         │    ├─ selectManagerPlan()  → manager + args
  │         │    ├─ lookPath(manager)    → check binary on PATH
  │         │    ├─ runInstall()        → execute command (30-min timeout)
  │         │    └─ postInstallLockfileState()  → re-hash after install
  │         └─ record outcome
  │
  └─ saveRunOutcomes()                → atomic state write (if changed)
```

## Background hook flow (async mode)

```
shell cd → hook snippet → pupdate hook --quiet --async
  │
  ├─ claimBackgroundHookLock()        → create .pupdate.hook.lock (O_EXCL)
  │    └─ if lock exists and not stale → skip (background run already active)
  │
  ├─ resolveExecutablePath()          → current binary path
  │
  ├─ startBackgroundProcess()          → fork: pupdate hook --child --lock-file <path>
  │
  └─ writeBackgroundHookLock()         → write PID to lock file

  (in background child:)
  pupdate hook --child --lock-file <path>
    ├─ defer removeBackgroundHookLock()
    └─ executeRun()                   → same as foreground run
```

## Design constraints

- **Low-latency `cd` hook:** The async background mode ensures the shell is never blocked. The hook only checks if `PWD` changed and is not `$HOME`.
- **PATH-based binary resolution:** Manager binaries are resolved via `exec.LookPath` from the current process `PATH`, ensuring compatibility with `nvm`, `asdf`, and `mise`.
- **Safe defaults:** Lifecycle scripts are disabled by default. Python managers are skipped by default. Frozen/locked install modes are used where supported.
- **Atomic state writes:** State is saved via temp file + rename + parent directory fsync to prevent corruption.
- **No daemon:** pupdate does not run as a background service. The async hook forks a short-lived process per `cd`.
