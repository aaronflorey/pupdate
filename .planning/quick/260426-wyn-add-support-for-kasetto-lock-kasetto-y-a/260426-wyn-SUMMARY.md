# Quick Summary: Add support for Kasetto

Status: Done

This quick task adds local support for Kasetto detection and execution without waiting for upstream manifest or manager metadata support.

Changes completed:
- Added a dedicated `kasetto` ecosystem with detection for `kasetto.lock`, `kasetto.yaml`, and `kasetto.yml`.
- Added a direct install plan for `kst sync` when Kasetto is detected.
- Updated detection and run-path tests to cover the new ecosystem.
- Documented Kasetto in the README supported ecosystems table.

Verification:
- `tmp="$(mktemp -d)" && HOME="$tmp" XDG_CONFIG_HOME="$tmp" go test ./internal/detection ./cmd/pupdate`
