---
status: complete
trigger: "/gsd:quick add a pupdate config command that shows the resolved user config path and active values, including root_directory resolution"
---

# Quick Task

## Current Focus

### hypothesis
The existing config loader already resolves `root_directory`, so a small refactor can expose both the raw configured value and the active resolved value for a new `pupdate config` command.

### next_action
Done.

## Evidence

- timestamp: 2026-04-22T00:00:00Z
  note: `cmd/pupdate/config.go` already resolves the user config file from `os.UserConfigDir()` and expands `root_directory` via `expandConfiguredDirectory`.
- timestamp: 2026-04-22T00:05:00Z
  note: `cmd/pupdate/root.go` currently only wires `run` and `init`, so adding `config` is localized to the CLI package.
- timestamp: 2026-04-22T00:12:00Z
  note: Added `pupdate config`, refactored config loading so raw and resolved values can both be reported, and updated README command docs.
- timestamp: 2026-04-22T00:14:00Z
  note: `rtk go test ./...` passed after adding config command coverage for present and missing config files.

## Resolution

### root_cause
There was no dedicated CLI surface for inspecting the resolved user config path or the effective `root_directory`, so users had to infer resolution behavior indirectly from `run` command skips.

### fix
Added `pupdate config` to print the resolved config file path, whether the file exists, the raw configured `root_directory`, and the resolved active value after expansion and normalization.
