# Phase 15 Research - Performance, Diagnostics, and Config/Hook Follow-Ups

**Date:** 2026-04-30
**Phase:** 15-performance-diagnostics-and-config-hook-follow-ups
**Question:** How should the approved post-Phase-14 maintenance improvements be split so future execution stays narrow while covering performance, observability, state hygiene, config breadth, and opt-in hook behavior?

## Confirmed Gaps

1. **Freshness hot-path work still has a performance/correctness follow-up**
   - Phase 11 plan 03 removed unsafe metadata-only hash reuse.
   - The approved follow-up is to reuse stored lockfile metadata only when it is safe to short-circuit full freshness hashing.

2. **There is no dedicated diagnostic surface for operator troubleshooting**
   - Users currently rely on run output, state files, and code behavior to understand why updates did or did not happen.
   - The approved follow-up is a `pupdate status` or `pupdate doctor` command.

3. **Stored state can drift when tracked targets disappear**
   - `.pupdate` needs a maintenance pass to prune entries for removed directories, ecosystems, or lockfiles so state stays truthful over time.

4. **Performance claims need explicit measurement and regression protection**
   - The project promises low-latency shell-hook behavior but does not yet have a dedicated maintenance phase for benchmarks and CI guardrails around that budget.

5. **User configuration is still intentionally narrow**
   - `root_directories` exists today, but the approved follow-up is to broaden the config surface in a controlled way.

6. **Hook execution still lacks an opt-in async/background mode**
   - Foreground shell-hook behavior is the current default.
   - The approved follow-up is to add a clearly optional background execution mode.

## Recommended Phase Breakdown

### Plan 15-01 - Safe lockfile-metadata reuse

- Define the safe boundary for reusing stored lockfile metadata without reintroducing the correctness issue fixed in Phase 11.
- Keep the work focused on hot-path freshness decisions and state data needed to prove the shortcut is safe.

### Plan 15-02 - Diagnostic command

- Add a dedicated CLI surface for surfacing current detection, freshness, state, config, and environment issues.
- Decide during execution whether `status` or `doctor` better matches the intended operator workflow.

### Plan 15-03 - Stale state pruning

- Clean `.pupdate` entries when tracked directories or ecosystems disappear.
- Keep pruning semantics explicit so users do not lose still-relevant state unexpectedly.

### Plan 15-04 - Benchmarks and latency guardrails

- Add focused benchmarks and CI checks that protect the hot path from regressions.
- Keep guardrails practical so they catch drift without creating noisy or flaky failures.

### Plan 15-05 - Config expansion

- Define and implement the next small set of user-facing config options beyond `root_directories`.
- Keep schema/documentation/test changes aligned so config growth remains understandable.

### Plan 15-06 - Opt-in async/background hook mode

- Add a clearly optional shell-hook execution mode that can decouple navigation latency from update execution.
- Preserve foreground behavior as the default and maintain visible status/error expectations.

## Research Outcome

Phase 15 should be tracked as six narrow maintenance plans because the approved work splits cleanly into one concern each for safe freshness optimization, diagnostics, stale-state hygiene, measurable latency protection, config expansion, and opt-in async hook execution.
