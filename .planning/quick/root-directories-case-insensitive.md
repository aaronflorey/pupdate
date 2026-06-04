---
status: complete
trigger: "/gsd:quick make the config root_directories case-insensitive"
---

# Quick Task

## Current Focus

### hypothesis
Normalizing configured roots and the current working directory to a shared case before relative-path checks will make `root_directories` matching case-insensitive without changing the top-level-only rule.

### next_action
Done.

## Evidence

- timestamp: 2026-04-26T00:00:00Z
  note: `isTopLevelDirectoryWithinRoot` currently compares `filepath.Rel` on resolved paths directly, so mixed-case `root_directories` entries do not match on case-sensitive platforms.
- timestamp: 2026-04-26T00:10:00Z
  note: Switched root-directory comparison to use a lowercased normalized path on both sides before `filepath.Rel`, preserving the existing top-level-only check.
- timestamp: 2026-04-26T00:15:00Z
  note: Added unit and run-command regression tests for mixed-case configured roots and verified them with `XDG_CONFIG_HOME=$(mktemp -d) go test ./cmd/pupdate -run 'TestIsTopLevelDirectoryWithinRoot|TestRunAllowsConfiguredRootDirectoriesCaseInsensitiveMatch'`.

## Resolution

### root_cause
Configured root directories were resolved but still compared with exact path casing, so a mixed-case config entry failed to match the current working directory on case-sensitive platforms.

### fix
Normalized both the configured root and candidate directory to a shared lowercase form before the relative-path check, and added regression coverage for the direct matcher plus the `run` command path.
