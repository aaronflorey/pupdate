# Architecture

This document describes pupdate's internal structure, package boundaries, and
data flow. All paths refer to the repository root.

## Repository layout

```
cmd/pupdate/
  main.go             # Build metadata, signal context, CLI delegation
internal/
  cli/                # Cobra commands and orchestration
  config/             # Strict user/project config loading and merging
  detection/          # Ecosystem detection from lockfiles/manifests
  freshness/          # Freshness evaluation (hash comparison)
  install/            # Safe plans and context-aware manager execution
  repolock/           # Cross-command advisory coordination
  shellhook/          # Snippets, managed files, detached launch
  state/              # State file (.pupdate) model and storage
  userenv/            # Home/config/path resolution
```

## Package boundaries

### `cmd/pupdate` and `internal/cli`

The executable creates a SIGINT/SIGTERM-aware context and delegates to
`internal/cli.ExecuteContext`. The CLI package:

- Defines the CLI surface (`run`, `init`, `hook`, `config`, `status`, `reset`).
- Loads strict user and traversal-only project configuration through `internal/config`.
- Runs preflight checks (home directory, root_directories, .pupignore).
- Coordinates detection, freshness evaluation, and install execution.
- Coordinates state persistence, shell-hook services, and repository locking.
- Captures filesystem, process, lookup, and service dependencies per command so
  orchestration has no mutable process-wide runtime seams.

Supporting packages keep manager policy, platform process handling, lock
semantics, and managed shell-file writes separate from Cobra handlers.

### `internal/detection`

Detects supported dependency ecosystems from lockfiles and manifests.

- **Entry point:** `DetectWithOptions(dir string, options Options) ([]DetectionResult, error)`
- **Scanning:** Scans the repository root (`.`), depth-1 subdirectories, and direct children of `packages/`. Optionally scans `workspace_globs` matches.
- **Signal matching:** Builds a registry from embedded manager definitions, with local fallbacks for Kasetto and Git submodules.
- **Gitignore support:** Reads `.gitignore` and skips matching directories during traversal.
- **Folder blacklist:** Skips directories whose exact name matches a `folder_blacklist` entry.
- **Output:** `[]DetectionResult`, each with `Ecosystem`, `Directory`, `Managers`, `MatchedFiles`, and `Warnings`.

### `internal/freshness`

Evaluates whether detected ecosystems need updates by comparing current lockfile
hashes against stored state.

- **Entry points:** `EvaluateContext(...)` for decisions and `SnapshotMatchedFiles(...)` for post-install hashes without another Git probe.
- **Hashing:** SHA-256 of matched lockfiles.
- **File-identity optimization:** On Linux and macOS, if a file's size, mtime, mode, inode (`file_id`), and ctime (`change_time_unix_nano`) are all unchanged, the stored hash is reused instead of rehashing. On other platforms, files are always rehashed.
- **Git submodules:** Runs `git submodule status --recursive` independently from each detected Git project directory (2-second timeout) and checks for drift indicators (`-`, `+`, `U` prefixes).
- **PHP vendor check:** For PHP projects with legacy `vendor/.pupdate-checksum` state, checks if the `vendor/` directory exists; if missing, forces an update.
- **Output:** `[]EcosystemDecision`, each with `Decision` (`update` or `skip`), `Reason`, `Lockfiles`, and `LockfileMetadata`.

### `internal/state`

Manages the `.pupdate` state file.

- **Model:** `FileState` with `Version` (schema 1) and `Ecosystems` map.
- **State keys:** Ecosystem name for root (e.g. `go`), `<ecosystem>@<directory>` for subdirectories (e.g. `node@apps/web`).
- **Storage:** `Store.Load()` reports state, warnings, and whether the file existed; `Save()` is atomic (temp file + rename + parent fsync on Unix).
- **Validation:** Invalid JSON is treated as empty with a warning. Schema version mismatches are treated as empty with a warning.

## Data flow

```
pupdate run
  │
  ├─ repolock.Acquire()              → nonblocking advisory lock
  │
  ├─ collectPreflight()
  │    ├─ config.Loader.Load()        → user + project config in one read each
  │    ├─ resolve/merge config        → user-first traversal union
  │    ├─ collectPreflightSkipReason()
  │    │    ├─ isHomeDirectory()      → skip if $HOME
  │    │    ├─ isOutsideConfiguredRootDirectories()  → skip if outside roots
  │    │    └─ hasPupIgnore()         → skip if .pupignore present
  │    ├─ Store.Load()                → current state (or empty)
  │    ├─ detection.DetectWithOptions()  → []DetectionResult
  │    └─ freshness.EvaluateContext() → []EcosystemDecision
  │
  ├─ resolveRunOptions()              → merge config defaults + flags
  │
  ├─ executeRunResults()
  │    └─ for each detection result:
  │         ├─ effectiveActionFor()   → unchanged, disabled, policy, PATH, ready
  │         ├─ if ready:
  │         │    ├─ install.Runner.Run() → execute command (30-min timeout)
  │         │    └─ SnapshotMatchedFiles() → re-hash without another Git probe
  │         └─ record outcome
  │
  ├─ saveRunOutcomes()                → save successes before aggregate failure
  └─ release repository lock
```

## Background hook flow (async mode)

```
shell cd → hook snippet → pupdate hook --quiet --async
  │
  ├─ shellhook.ResolvedExecutablePath() → current binary path
  │
  └─ shellhook.StartBackground()      → fork: pupdate hook --child
      ├─ Setsid (Unix)                → new session, detached from terminal
      └─ stdin/stdout/stderr = /dev/null

  (in background child:)
  pupdate hook --child
    ├─ repolock.Acquire()             → same lock as run/reset
    ├─ executeRun()                   → full operation while locked
    └─ release lock; coordination file remains
```

## Design constraints

- **Low-latency `cd` hook:** The async background mode ensures the shell is never blocked. The hook only checks if `PWD` changed and is not `$HOME`.
- **Terminal detachment:** The background child runs in a new session (`setsid` on Unix) with stdin, stdout, and stderr connected to `/dev/null`. This prevents the child from holding the shell's terminal file descriptors open, which would block the shell until the child exits.
- **PATH-based binary resolution:** Manager binaries are resolved via `exec.LookPath` from the current process `PATH`, ensuring compatibility with `nvm`, `asdf`, and `mise`.
- **Safe defaults:** Composer plugins and scripts are disabled, other lifecycle scripts are disabled where supported, Python managers require opt-in, Yarn Modern is blocked, and frozen/locked modes are used where supported.
- **Process trees:** On Linux and macOS, every manager runs in its own process group; timeout or cancellation kills descendants before the repository lock is released.
- **Persistent coordination:** Manual runs, hook workers, and resets use the same advisory lock. File age and legacy PID contents do not indicate activity.
- **Atomic state writes:** State is saved via temp file + rename + parent directory fsync to prevent corruption.
- **No daemon:** pupdate does not run as a background service. The async hook forks a short-lived process per `cd`.
