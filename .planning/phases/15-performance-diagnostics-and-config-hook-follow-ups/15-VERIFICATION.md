---
phase: 15-performance-diagnostics-and-config-hook-follow-ups
verified: 2026-04-30T03:59:41Z
status: passed
score: 6/6 must-haves verified
overrides_applied: 0
human_verification:
  - test: "Exercise generated bash/zsh/fish hooks in foreground and async modes during real directory changes"
    expected: "Default init snippet stays foreground, async mode returns the prompt immediately, and hook-triggered status/error output remains understandable in the interactive shell"
    why_human: "Interactive shell timing and detached stderr visibility cannot be fully proven from static inspection or unit tests"
  - test: "Trigger overlapping async hook runs in a real repo"
    expected: "Second background launch is skipped without corrupting .pupdate state, and status output matches the lock lifecycle while the child process is active or stale"
    why_human: "Real process scheduling and shell-trigger overlap behavior are environment-dependent"
---

# Phase 15: performance, diagnostics, and config/hook follow-ups Verification Report

**Phase Goal:** Land the next approved maintenance improvements by tightening safe hot-path freshness reuse, improving operator diagnostics, pruning stale state, broadening config surface area, and adding optional asynchronous hook execution without reopening milestone-scale scope.
**Verified:** 2026-04-30T03:59:41Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
| --- | --- | --- | --- |
| 1 | Safe freshness reuse only skips rehashing when unchanged-file identity is provably strong | ✓ VERIFIED | `internal/freshness/engine.go:262-289` reuses stored hashes only when size, mtime, mode, file ID, and change time all match; `internal/freshness/engine_test.go:125-298` covers reuse, old-state fallback rehashing, metadata drift, and same-metadata content rewrite detection. |
| 2 | Users have a read-only diagnostic command that explains run/skip/block reasons from real runtime state | ✓ VERIFIED | `cmd/pupdate/status.go:48-146` adds `status`, loads real config/state, runs detection and freshness evaluation, and renders per-target readiness; `cmd/pupdate/root.go:24-27` wires the command; `cmd/pupdate/status_test.go:18-268` covers ready, repo-skip, invalid-config, active hook lock, and missing-manager cases. |
| 3 | `.pupdate` state is pruned when previously tracked targets disappear, without mutating still-active target state | ✓ VERIFIED | `cmd/pupdate/run_state.go:20-107` builds the next state from active detections and saves cleanup-only changes; `cmd/pupdate/run_execution.go:78-79` passes current detections into persistence; `cmd/pupdate/run_state_test.go:134-168` and `cmd/pupdate/run_test.go:913-960` verify stale subdirectory pruning and active-target retention. |
| 4 | Hot-path performance is measured and guarded automatically | ✓ VERIFIED | `internal/detection/detector_benchmark_test.go:9-32` benchmarks detection, `internal/freshness/engine_benchmark_test.go:11-60` benchmarks reuse vs rehash, `internal/freshness/performance_guardrail_test.go:11-62` enforces a 20x relative guardrail, and `.github/workflows/ci.yml:37-38` runs the targeted benchmarks in CI. |
| 5 | User config now controls more than `root_directories`, and the new config defaults flow through run and status behavior | ✓ VERIFIED | `cmd/pupdate/config.go:16-20` adds `quiet` and `allow_scripts`; `cmd/pupdate/run_execution.go:101-115` resolves config defaults with flag override precedence; `cmd/pupdate/status.go:94-146,302-324` reports configured/effective values and uses shared run-option resolution; `cmd/pupdate/config_cmd.go:39-46` surfaces them in `config`; tests in `cmd/pupdate/config_test.go`, `config_cmd_test.go`, and `status_test.go` cover defaults, overrides, parse errors, and surfaced output. |
| 6 | Async hook execution is opt-in, backgrounded through a hidden hook command, and default foreground hook behavior remains intact | ✓ VERIFIED | `cmd/pupdate/init.go:13-45`, `init_shell.go:12-55`, and `init_snippets.go:5-20` keep foreground as the default and append `--async` only for explicit async mode; `cmd/pupdate/hook.go:41-280` adds the hidden hook command, detached child launch, repo-local lock, stale-lock handling, and status reporting; `cmd/pupdate/init_test.go:9-212` and `cmd/pupdate/hook_test.go:16-265` cover default foreground snippets, async snippets, overlap skip, detached launch, lock cleanup, and running-PID lock handling. |

**Score:** 6/6 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
| --- | --- | --- | --- |
| `internal/freshness/engine.go` | Safe hot-path lockfile hash reuse logic | ✓ VERIFIED | Substantive freshness engine; reuse gate is narrow and backed by strong metadata checks. |
| `internal/state/model.go` | Persisted metadata fields needed for safe reuse | ✓ VERIFIED | Adds `file_id` and `change_time_unix_nano` to lockfile metadata. |
| `cmd/pupdate/status.go` | Read-only diagnostics surface | ✓ VERIFIED | Collects config, state, detection, freshness, and hook-lock status; no save/install calls. |
| `cmd/pupdate/run_state.go` | Pruning and cleanup-only state persistence | ✓ VERIFIED | Keeps only active keys and persists metadata refresh/prune-only changes. |
| `internal/detection/detector_benchmark_test.go` | Detection benchmark | ✓ VERIFIED | Benchmarks project-tree detection over representative fixtures. |
| `internal/freshness/performance_guardrail_test.go` | Automatic latency regression guardrail | ✓ VERIFIED | Fails if stored-hash reuse is no longer meaningfully faster than rehashing. |
| `cmd/pupdate/config.go` | Expanded config schema | ✓ VERIFIED | Adds `quiet` and `allow_scripts` without introducing a parallel config model. |
| `cmd/pupdate/hook.go` | Hidden async hook implementation and overlap lock | ✓ VERIFIED | Implements async child launch, lock claim/remove flow, and status visibility. |
| `README.md` | Updated operator-facing docs for status/config/async hook usage | ✓ VERIFIED | Documents `status`, config keys, and `init --mode async`. |

### Key Link Verification

| From | To | Via | Status | Details |
| --- | --- | --- | --- | --- |
| `cmd/pupdate/root.go` | `cmd/pupdate/status.go` | `cmd.AddCommand(newStatusCmd())` | ✓ WIRED | `root.go:24-27` registers `status`, `hook`, `init`, and `config`. |
| `cmd/pupdate/status.go` | detection/freshness/state/config runtime layers | `collectStatusSnapshot()` | ✓ WIRED | `status.go:72-146` loads config/state, runs `detectFn(".")`, and calls `evaluateFreshnessFn(".", results, currentState)`. |
| `cmd/pupdate/run_execution.go` | `cmd/pupdate/run_state.go` | `saveRunOutcomes(...)` | ✓ WIRED | `run_execution.go:78-79` forwards live detections/outcomes into state persistence so pruning and metadata refresh can occur. |
| `cmd/pupdate/init*.go` | `cmd/pupdate/hook.go` | generated `pupdate hook --quiet[ --async]` snippets | ✓ WIRED | `init_snippets.go:5-20` generates hook invocations; `hook.go:65-79` dispatches foreground vs async execution. |
| `.github/workflows/ci.yml` | benchmark/guardrail tests | `go test ... -bench ...` | ✓ WIRED | `ci.yml:37-38` runs the targeted detection and freshness benchmarks automatically in CI. |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| --- | --- | --- | --- | --- |
| `cmd/pupdate/status.go` | `snapshot.Targets` / `snapshot.RunStatus` | `detectFn(".")` + `evaluateFreshnessFn(".", results, currentState)` + loaded config/state | Yes | ✓ FLOWING |
| `cmd/pupdate/run_state.go` | `next.Ecosystems` | current `.pupdate` state + live `activeResults` + run outcomes | Yes | ✓ FLOWING |
| `cmd/pupdate/hook.go` | hook lock status/path | real filesystem lock file + PID liveness probe | Yes | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| --- | --- | --- | --- |
| Targeted maintenance regressions pass | `go test ./cmd/pupdate ./internal/freshness ./internal/detection ./internal/state -count=1` | `ok` for all four packages | ✓ PASS |
| Hook/status regression coverage passes | `go test ./cmd/pupdate -run 'TestInit.*|TestExecuteHook.*|TestLaunchBackgroundHook.*|TestStatus.*|TestRunPrunesUndetectedSubdirectoryStateWithoutChangingActiveTargets' -count=1` | `ok github.com/aaronflorey/pupdate/cmd/pupdate` | ✓ PASS |
| Hot-path benchmarks still show large reuse win | `go test ./internal/detection ./internal/freshness -run '^$' -bench 'Benchmark(DetectProjectTree|HashMatchedFiles)$' -count=1` | Detection `80288 ns/op`; freshness reuse `1617 ns/op` vs rehash `2092912 ns/op` | ✓ PASS |
| CLI snippet/config surfaces behave as documented | `go run ./cmd/pupdate init --shell bash`; `go run ./cmd/pupdate init --shell bash --mode async`; `XDG_CONFIG_HOME="$(mktemp -d)" HOME="/tmp/pupdate-home" go run ./cmd/pupdate config` | Default snippet prints `pupdate hook --quiet`; async snippet prints `pupdate hook --quiet --async`; config prints unset defaults without creating config | ✓ PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| --- | --- | --- | --- | --- |
| N/A | Phase 15 | No requirement IDs declared; phase is maintenance hardening | ✓ SATISFIED | Roadmap entry and all six maintenance truths verified directly against code. |

### Anti-Patterns Found

No blocker or warning anti-patterns found in the inspected phase artifacts. The only grep hit was a test-only error string in `cmd/pupdate/config_test.go`.

### Human Verification Required

### 1. Interactive shell mode parity

**Test:** Install generated bash, zsh, and fish snippets for both `pupdate init --shell <shell>` and `pupdate init --shell <shell> --mode async`, then `cd` between repos that do and do not need updates.
**Expected:** Foreground mode keeps the pre-existing visible hook behavior; async mode returns the prompt immediately while still surfacing understandable run/error output.
**Why human:** Interactive hook timing and stderr presentation depend on the real shell environment.

### 2. Overlap and stale-lock behavior under real detached runs

**Test:** Trigger two rapid directory-entry events in the same repo with async mode enabled, then inspect `pupdate status` during and after the detached run.
**Expected:** The second launch is skipped while the first lock is active; `pupdate status` reports `active` during the run and `idle` or `stale` appropriately after completion/failure.
**Why human:** Detached child process timing and lock lifecycle are OS/shell scheduling behaviors, not purely static code properties.

### Gaps Summary

No code-level gaps were found. All six maintenance subplans are implemented and wired into the CLI/runtime layers. Remaining verification is limited to real interactive-shell confirmation for async hook UX and overlap behavior.

### Human Validation

- 2026-04-30: User validated the interactive shell checks for foreground vs async hook behavior and overlap/lock lifecycle. Result: passed.

---

_Verified: 2026-04-30T03:59:41Z_
_Verifier: the agent (gsd-verifier)_
