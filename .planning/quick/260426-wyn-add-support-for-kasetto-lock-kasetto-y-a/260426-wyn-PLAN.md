# Quick Plan: Add support for Kasetto

## Goal

Teach `pupdate` to detect `kasetto.yaml`, `kasetto.yml`, and `kasetto.lock`, then run `kst sync` when an update is needed.

## Planned Changes

1. Add a local Kasetto ecosystem and detection fallback for the supported signal files.
2. Add a direct manager plan for `kst sync` because upstream manager definitions do not include Kasetto.
3. Extend focused tests for detection, manager planning, and run output.
4. Update docs to list Kasetto as a supported ecosystem.
