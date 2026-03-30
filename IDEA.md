
Task: Create a golang tool that installs all packages in a local repository

#### Stack
- Golang
- spf12/cobra
- release-please
- goreleaser
- github actions

#### Commands

`run`:  it should scan for package managers in the local repository and run `install` for that package manager
`init`: sets up the users shell to run `run` when you cd into a repository

`run` should be extremely lightweight so that the user doesn't notice any lag when you `cd` - think how `nvm` works

#### State
store a `.pupdate` file that keeps state of the last run. it should store the hash of the lock file for that package manager (if one exists), and only run if that hash changes or no hash exists.

if the folder has a `.pupignore` then don't run in the current folder.

consider if we should also store the hash of the local vendor file to see if it's also drifted. ie: if composer.lock hasn't changed but ./vendor has, then we should run composer install anyway.

#### Package Managers
we'll start with a minimum set of package managers. composer and bun should be in the first MVP.

- PHP: composer
- Node: bun/pnpm/npm/yarn
- Python: uv, poetry, pip
- Go: go mod
- Rust: cargo
- Git: .gitmodules (make sure the local submodules are the correct version)

 # LLM Code Generation Guidelines (Anti-Slop Baseline)

  ## Core Principles
  - Prefer clarity over cleverness. Code must be easy to read at a glance.
  - Prioritise maintainability over short-term convenience or micro-optimisation.
  - Follow existing project conventions and patterns before introducing new ones.

  ## Structure & Organisation
  - Group related logic into appropriate namespaces/modules.
  - Keep functions and classes small, focused, and single-purpose.
  - Avoid deep nesting; favour early returns and simple control flow.

  ## Reusability & Abstraction
  - Eliminate duplication. Extract shared logic into reusable components.
  - Use interfaces/abstractions where multiple implementations are likely.
  - Ensure code is loosely coupled and testable.

  ## Typing & Safety
  - Enforce strict typing; avoid implicit or weak types.
  - Validate inputs and handle errors explicitly; do not allow silent failures.

  ## Naming & Readability
  - Use clear, descriptive naming.
  - Avoid abbreviations unless they are widely understood.
  - Write code that is understandable without needing comments.

  ## Comments
  - Only add comments where intent is not obvious at a glance.
  - Do not restate what the code already clearly expresses.

  ## Dependencies & Libraries
  - Prefer existing, well-maintained libraries over custom implementations.
  - Do not reimplement standard functionality (e.g. HTTP, validation, parsing, caching).
  - Search for packages before writing new logic.
  - If a package solves ≥80% of the problem, use it and extend if necessary.

  ## Package Selection Criteria
  - Must have recent activity (release or commit within the last 6 months).
  - Must have meaningful adoption (generally 100–500+ stars, context dependent).
  - Must have clear documentation.
  - Prefer widely used, ecosystem-standard libraries.
  - Avoid abandoned, low-quality, or single-maintainer risk projects (unless necessary).

  ## Package Selection Process
  - Identify 2–3 candidate libraries.
  - Select the most appropriate based on maintenance, adoption, and fit.
  - Briefly justify the choice and why alternatives were not chosen.

  ## Constraints
  - Do not introduce unnecessary dependencies.
  - Do not over-engineer solutions.
  - Keep implementations simple, predictable, and maintainable.
