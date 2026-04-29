# Phase 02 Research — Implement Other Package Managers

**Date:** 2026-03-31
**Phase:** 02-implement-other-package-managers-from-idea-md
**Question:** What do we need to know to plan this phase well?

## Scope Extracted from IDEA.md

Add ecosystem coverage beyond MVP (bun/composer) while preserving current constraints:

- Keep `run` low-latency on directory entry.
- Resolve binaries from runtime `PATH`.
- Keep safe/default install behavior (non-interactive, frozen/locked, scripts minimized where supported).
- Continue hash-based skip behavior using `.pupdate` state.

Target requirement IDs for this phase:

- `ECO-02`: npm, pnpm, yarn
- `ECO-03`: uv, poetry, pip
- `ECO-04`: go mod, cargo
- `ECO-05`: git submodules

## Existing Codebase Patterns to Reuse

1. **Signal-based detection** (`internal/detection/matrix.go`, `internal/detection/detector.go`)
   - Lockfile/manifest signals are deterministic and tested.
2. **Safety-first manager planning** (`cmd/pupdate/run.go::selectManagerPlan`)
   - Exact args are asserted in tests.
3. **Freshness decisions from lockfile hashes** (`internal/freshness/engine.go`)
   - Decision reasons are explicit and user-visible.
4. **Success-only state persistence** (`cmd/pupdate/run.go::applySuccessfulOutcomes`)
   - Failed runs do not mutate `.pupdate`.

## Recommended Technical Approach

### 1) Expand detection matrix first (contract-first)

Add signals and manager mapping for new ecosystems before execution wiring.

Recommended lockfile/manifest signals:

- Node: `bun.lock`, `pnpm-lock.yaml`, `package-lock.json`, `yarn.lock`
- Python: `uv.lock`, `poetry.lock`, `requirements.txt`
- Go: `go.mod`
- Rust: `Cargo.lock`
- Git: `.gitmodules`

### 2) Add manager plans with concrete safe defaults

Proposed install/update commands:

- `bun install --frozen-lockfile --ignore-scripts`
- `pnpm install --frozen-lockfile --ignore-scripts`
- `npm ci --ignore-scripts`
- `yarn install --frozen-lockfile --ignore-scripts`
- `composer install --no-interaction --prefer-dist --no-scripts`
- `uv sync --frozen`
- `poetry install --no-interaction --sync`
- `pip install -r requirements.txt --disable-pip-version-check --no-input`
- `go mod download`
- `cargo fetch --locked`
- `git submodule update --init --recursive`

### 3) Handle submodule drift explicitly (ECO-05)

`.gitmodules` hash alone is insufficient to detect detached/stale local submodule checkout state.

Recommended runtime check for git ecosystems:

- Run `git submodule status --recursive`
- If any line starts with `-`, `+`, or `U`, treat as stale and force update even when `.gitmodules` hash is unchanged.

### 4) Preserve runtime and safety constraints

- Continue `exec.LookPath` checks for every manager binary.
- Keep stderr status contract:
  - `pupdate: skip ...`
  - `pupdate: run ...`
  - `pupdate: error ...`
- Keep skip-first behavior for unchanged lockfiles unless git submodule drift is detected.

## Risks & Mitigations

1. **Multiple lockfiles in same ecosystem**
   - Risk: choosing wrong manager.
   - Mitigation: deterministic manager precedence + explicit warnings and skip on ambiguity.

2. **Tool availability mismatch (`PATH`)**
   - Risk: command failures during hooks.
   - Mitigation: retain per-manager `LookPath` checks and print explicit skip reason.

3. **pip mutating environment unexpectedly**
   - Risk: noisy/slow updates.
   - Mitigation: lock to `requirements.txt` path and non-interactive flags; keep update guarded by freshness and explicit status output.

4. **Submodule checks increasing latency**
   - Risk: hook slowdown.
   - Mitigation: run submodule status only when `.gitmodules` is detected.

## Test Strategy (research output)

- Expand detection tests to assert new ecosystem signals and manager mapping.
- Expand run tests to assert exact manager args for each new manager.
- Add git submodule drift tests to prove update occurs when submodule state is stale.
- Keep smoke verification command `go test ./... -count=1`.

## Validation Architecture

Use fast, continuous `go test` checks during execution with task-scoped commands and full-suite confirmation after each plan wave.

- Task-level: targeted package tests (e.g., `go test ./internal/detection -count=1`)
- Wave-level: `go test ./... -count=1`
- Block on failing tests before moving to next wave.
