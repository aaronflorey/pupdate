<!-- GSD:project-start source:PROJECT.md -->
## Project

**pupdate**

pupdate is a fast Go CLI that detects package ecosystems in the current directory and runs the correct dependency update commands automatically when entering a project. It supports JavaScript/TypeScript, PHP, Go, Rust, and Python projects by detecting lockfiles/manifests and using the corresponding package manager binary from the user's configured `PATH`. It tracks update history in a local `.pupdate` state file and skips unnecessary work unless dependency files changed.

**Core Value:** Keep project dependencies up to date automatically on directory entry without noticeably slowing down shell navigation.

### Constraints

- **Performance**: Hook invocation on `cd` must be low-latency — avoid expensive scans/processes when no relevant files changed.
- **Runtime detection**: Manager binaries must be discovered via current process `PATH` — required for environment managers like `nvm`, `asdf`, and `mise`.
- **Shell compatibility**: Initial v1 should support at least `bash` and `zsh` hook setup reliably via `init`.
- **Safety**: Auto-update behavior must be transparent and non-blocking; users need visible status and errors.
- **Release automation**: GitHub Actions pipeline must support semver/tagged releases via Release Please and GoReleaser.
<!-- GSD:project-end -->

<!-- GSD:stack-start source:research/STACK.md -->
## Technology Stack

## Recommendation
- **Language**: Go (stable toolchain, current LTS branch)
- **CLI framework**: `spf13/cobra`
- **Config/state serialization**: `gopkg.in/yaml.v3` (or JSON only using stdlib)
- **Fast file checks**: stdlib `os.Stat` + `filepath`
- **Process execution**: stdlib `os/exec` with context timeouts
- **Shell integration**: generated shell snippets for `bash` and `zsh`
- **Release automation**: Release Please + GoReleaser + GitHub Actions
## Why this stack
- Go compiles to a single fast binary with low startup latency.
- `cobra` gives predictable command ergonomics (`init`, `run`, `status`) and shell completion support.
- stdlib process + filesystem APIs are enough for high-performance detection/update paths.
- GitHub-native release tooling keeps changelog, tagging, and binaries aligned.
## What not to use
- Full daemon/background service for v1: unnecessary complexity and platform edge cases.
- Heavy file watchers for every directory: higher overhead than cheap mtime checks on `cd`.
- Shell-specific hardcoded binary paths: breaks user-managed runtime environments.
## Confidence
- **High**: Go + stdlib + Cobra for cross-platform CLI
- **High**: Release Please + GoReleaser for OSS-style releases
- **Medium**: YAML vs JSON for `.pupdate` (JSON is simpler; YAML is more human-readable)
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

- Assume the default git branch is `main` unless the user explicitly says otherwise.
- Always push to `main` unless the user explicitly says otherwise.
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

Architecture not yet mapped. Follow existing patterns found in the codebase.
<!-- GSD:architecture-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd:quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd:debug` for investigation and bug fixing
- `/gsd:execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->



<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd:profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
