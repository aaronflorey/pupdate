# Phase 13: final milestone audit documentation drift follow-ups - Context

**Gathered:** 2026-04-30
**Status:** Completed
**Mode:** Autonomous follow-up from latest milestone audit findings

<domain>
## Phase Boundary

Close the remaining post-Phase-12 documentation drift without reopening runtime behavior, workflow YAML, or broader milestone scope.

</domain>

<decisions>
## Implementation Decisions

### phase structure
- Keep the remaining documentation drift split into exactly two plans so release-planning reference cleanup and README CI-claim correction stay independently executable and verifiable.

### release-planning references
- Every surviving release-planning and planning-state reference should point to `.github/workflows/release.yaml`.
- Those references should describe the current release model accurately: pushes to `main` and `master`, Release Please on that workflow, and conditional GoReleaser checkout of the created tag.

### ci platform claim
- README platform claims should match the actual `ci.yml` matrix exactly.
- Do not broaden the claim beyond the currently configured Linux and macOS runners.

### execution posture
- Keep this phase documentation-only.
- Prefer the smallest set of artifact edits that removes stale wording and filename drift.

</decisions>

<code_context>
## Existing Code Insights

- `.github/workflows/release.yaml` is now the only release workflow in the repository.
- That workflow triggers on pushes to `main` and `master`, runs Release Please, and conditionally runs GoReleaser after checking out `needs.release-please.outputs.tag_name` when `release_created == 'true'`.
- `.github/workflows/ci.yml` currently runs on `ubuntu-latest` and `macos-latest` only.
- `README.md` still claims CI runs across Linux, macOS, and Windows.
- Older planning artifacts in `.planning/phases/03-*` still mention `.github/workflows/release.yml` even though the surviving workflow file is `.github/workflows/release.yaml`.

</code_context>

<specifics>
## Specific Ideas

- Update stale release-planning filenames and flow descriptions in the narrowest set of release-related planning artifacts and state text.
- Correct the README CI platform sentence so it mirrors the actual CI matrix and trigger scope.

</specifics>

<deferred>
## Deferred Ideas

- Any workflow logic changes or CI matrix expansion.
- Broader README or milestone-doc rewrites beyond the identified drift.

</deferred>
