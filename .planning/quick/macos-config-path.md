---
status: complete
trigger: "/gsd:quick make pupdate config path use ~/.config/pupdate/config.yaml on macOS too; keep tests green across OS"
---

# Quick Task

## Current Focus

### hypothesis
Switching config path resolution to a small OS-aware helper can force macOS to use `~/.config` without changing non-macOS behavior.

### next_action
Done.

## Evidence

- timestamp: 2026-04-22T00:00:00Z
  note: `resolveUserConfigPath` previously delegated fully to `os.UserConfigDir`, which points to `~/Library/Application Support` on macOS.
- timestamp: 2026-04-22T00:10:00Z
  note: Added darwin-specific config-dir resolution that prefers `XDG_CONFIG_HOME` and otherwise uses `$HOME/.config`.
- timestamp: 2026-04-22T00:15:00Z
  note: Added unit coverage for darwin path resolution and adjusted existing config-dir error tests to pin non-darwin behavior.
- timestamp: 2026-04-22T00:20:00Z
  note: `rtk go test ./...` passed (`91 passed`).

## Resolution

### root_cause
The resolver relied on `os.UserConfigDir` for every OS, so macOS defaulted to Apple's platform config location instead of `~/.config`.

### fix
Introduced `resolveUserConfigDir` to make darwin resolve to `XDG_CONFIG_HOME` or `$HOME/.config`, kept non-darwin on `os.UserConfigDir`, and covered the behavior with cross-OS-safe tests.
