---
phase: quick-260412-uxf-add-support-for-fish-shell-to-the-env-se
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - cmd/pupdate/init.go
  - cmd/pupdate/init_test.go
  - README.md
autonomous: true
requirements:
  - UXF-FISH-01
must_haves:
  truths:
    - "User can run `pupdate init --shell fish` and get a usable fish hook snippet."
    - "When `SHELL` resolves to fish and `--shell` is omitted, `pupdate init` emits fish snippet instead of bash fallback."
    - "Docs show fish setup alongside bash/zsh usage."
  artifacts:
    - path: "cmd/pupdate/init.go"
      provides: "fish snippet generation and shell resolution support"
      contains: "fish"
    - path: "cmd/pupdate/init_test.go"
      provides: "contract tests for fish explicit/default shell behavior"
      contains: "fish"
    - path: "README.md"
      provides: "operator docs for fish init usage"
      contains: "init --shell fish"
  key_links:
    - from: "cmd/pupdate/init.go"
      to: "cmd/pupdate/init_test.go"
      via: "root command init contract tests"
      pattern: "init --shell fish"
    - from: "README.md"
      to: "cmd/pupdate/init.go"
      via: "documented command maps to implemented shell option"
      pattern: "--shell <bash|zsh|fish>"
---

<objective>
Add fish shell support to `pupdate init` so users can install hook snippets in fish with the same quiet/non-blocking behavior as bash/zsh.

Purpose: Close shell UX gap for fish users without changing run-path performance characteristics.
Output: fish snippet support in CLI + tests + docs.
</objective>

<execution_context>
@$HOME/.config/opencode/get-shit-done/workflows/execute-plan.md
@$HOME/.config/opencode/get-shit-done/templates/summary.md
</execution_context>

<context>
@AGENTS.md
@cmd/pupdate/init.go
@cmd/pupdate/init_test.go
@README.md
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Extend init command contracts to cover fish shell output</name>
  <files>cmd/pupdate/init_test.go</files>
  <behavior>
    - Test 1: `pupdate init --shell fish` succeeds and output contains fish hook trigger plus `pupdate run --quiet`.
    - Test 2: unsupported shell error message lists fish in supported shells.
    - Test 3: when `SHELL=fish` and `--shell` omitted, emitted snippet is fish variant.
  </behavior>
  <action>Add/adjust tests first to define fish behavior contract while preserving existing stderr-visibility assertions (no `2>/dev/null`). Keep test scope focused to init command behavior only.</action>
  <verify>
    <automated>go test ./cmd/pupdate -run TestInit -count=1</automated>
  </verify>
  <done>Tests fail initially against current implementation and clearly describe required fish behavior.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Implement fish snippet generation and shell resolution</name>
  <files>cmd/pupdate/init.go</files>
  <behavior>
    - Explicit `--shell fish` returns fish snippet.
    - `resolveShell` accepts fish for explicit and env-derived shell detection.
    - Default fallback remains bash for unknown/empty shells.
  </behavior>
  <action>Add a fish init snippet that triggers `pupdate run --quiet` on directory change using fish-native event handling, then wire `newInitCmd` switch and `resolveShell` accepted values to include fish. Keep snippet lightweight and non-blocking (no extra subprocesses or filesystem scans).</action>
  <verify>
    <automated>go test ./cmd/pupdate -run TestInit -count=1</automated>
  </verify>
  <done>`init` supports fish explicitly and by SHELL detection, and init tests pass.</done>
</task>

<task type="auto">
  <name>Task 3: Document fish shell setup in README command docs</name>
  <files>README.md</files>
  <action>Update Quick Start and `pupdate init` sections to include fish example commands and expand shell flag docs from `bash|zsh` to `bash|zsh|fish`. Preserve existing messaging about quiet mode and stderr visibility.</action>
  <verify>
    <automated>go test ./cmd/pupdate -run TestInit -count=1</automated>
  </verify>
  <done>README includes fish copy/paste setup and matches implemented flag contract.</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| user shell config → generated snippet | Untrusted shell startup context executes printed hook code. |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-quick-01 | T | `cmd/pupdate/init.go` snippet output | mitigate | Keep fish snippet static (no interpolation of user input), matching existing bash/zsh static-template model. |
| T-quick-02 | D | hook execution on directory change | mitigate | Preserve lightweight hook pattern (`pupdate run --quiet` only) with no additional scans/processes in snippet logic. |
| T-quick-03 | R | unsupported shell handling | mitigate | Keep actionable error text listing supported shells so operators can configure correctly and audit behavior. |
</threat_model>

<verification>
- `go test ./cmd/pupdate -run TestInit -count=1`
- Optional smoke check:
  - `go run ./cmd/pupdate init --shell fish`
  - verify output contains fish event hook and `pupdate run --quiet`
</verification>

<success_criteria>
- `pupdate init --shell fish` returns a fish-native hook snippet.
- `pupdate init` auto-selects fish when `SHELL` is fish.
- Unsupported shell messaging now lists `bash, zsh, fish`.
- README documents fish setup and flag contract consistently.
</success_criteria>

<output>
After completion, create `.planning/quick/260412-uxf-add-support-for-fish-shell-to-the-env-se/260412-uxf-SUMMARY.md`
</output>
