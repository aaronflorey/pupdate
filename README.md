# pupdate

`pupdate` is a fast Go CLI that detects supported dependency ecosystems in the current
directory and runs safe, manager-specific install commands when needed.

## Quick Start

```bash
pupdate run
```

By default, `run` prints machine-readable JSON on stdout and concise status lines on stderr.

## Shell Hook Setup (bash/zsh)

Install the hook into your current shell session:

```bash
eval "$(pupdate init --shell bash)"
```

```bash
eval "$(pupdate init --shell zsh)"
```

Escaped form (for docs/templates that require literal shell interpolation text):

```bash
eval "\$(pupdate init --shell bash)"
eval "\$(pupdate init --shell zsh)"
```

The generated hooks run `pupdate run --quiet` on directory changes so prompt behavior stays
low-noise and non-blocking.

## Run Status Output

During hook-driven or manual runs, status lines are emitted on stderr with stable prefixes:

- `pupdate: skip repo (.pupignore)`
- `pupdate: skip <ecosystem> (<reason>)`
- `pupdate: run <manager> <args>`
- `pupdate: error <manager> install failed: <err>`

These lines are intended to be easy to scan and grep from shell history/logging.

## Safety Defaults

For current MVP ecosystems, install plans default to safe flags:

- bun: `install --frozen-lockfile --ignore-scripts`
- composer: `install --no-interaction --prefer-dist --no-scripts`

Phase-2 ecosystem coverage adds:

- npm: `ci --ignore-scripts`
- pnpm: `install --frozen-lockfile --ignore-scripts`
- yarn: `install --frozen-lockfile --ignore-scripts`
- uv: `sync --frozen`
- poetry: `install --no-interaction --sync`
- pip: `install -r requirements.txt --disable-pip-version-check --no-input`
- go: `mod download`
- cargo: `fetch --locked`
- git submodule: `git submodule update --init --recursive`

For git submodule repos, pupdate checks `git submodule status --recursive` and can trigger
updates when submodule state drifts even if `.gitmodules` content is unchanged.

Manager binaries are resolved from your current `PATH` at runtime.
