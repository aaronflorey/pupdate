# Troubleshooting

This document covers common symptoms, their likely causes, and fixes. All
status messages documented here come from the source code in `cmd/pupdate/`.

## Diagnostic commands

| Command | Purpose |
|---------|---------|
| `pupdate status` | Full read-only diagnostic snapshot for the current directory. |
| `pupdate config` | Show resolved config path and active values. |
| `pupdate run --dry-run` | Show what would run without executing installs or saving state. |
| `pupdate run` (without `--quiet`) | See all detection and freshness decisions on stderr. |

## Symptoms and fixes

### pupdate does nothing when I `cd` into a project

**Symptom:** No output at all when entering a directory.

**Likely causes:**

1. **You are in `$HOME`.** pupdate skips the home directory by design.
   - Message (without `--quiet`): `pupdate: skip repo ($HOME)`
   - Source: `cmd/pupdate/preflight.go` (`collectPreflightSkipReason`).

2. **`root_directories` is configured and you are outside them.**
   - Message: `pupdate: skip repo (outside configured root_directories)`
   - Fix: Either `cd` into a direct child of a configured root, or remove/adjust `root_directories` in your config. See [Configuration](configuration.md#root_directories).

3. **`.pupignore` is present.**
   - Message: `pupdate: skip repo (.pupignore)`
   - Fix: Remove the `.pupignore` file if you want pupdate to run in this repo.

4. **The hook is in async mode and the run is a quiet no-op.** Async hooks use `--quiet` by default, so no-op runs produce no output. This is expected behavior.
   - Fix: Run `pupdate run` manually (without `--quiet`) to see full status, or check `pupdate status`.

5. **A background hook is already running.**
   - Message: `pupdate: skip repo (background run already active)`
   - Fix: Wait for the existing run to finish, or remove `.pupdate.hook.lock` if it is stale.

### `pupdate: skip <target> (<reason>)`

This means a specific ecosystem was detected but its install was skipped. Common reasons:

| Reason | Cause | Fix |
|--------|-------|-----|
| `dependency lockfiles unchanged since last successful run` | Lockfiles have not changed since the last successful install. | This is expected — no action needed. Run `pupdate reset` to force a re-evaluation. |
| `dependency lockfiles changed since last successful run` | Lockfiles changed but the decision was still skip. | Should not normally appear with this reason; check `pupdate status`. |
| `missing prior lockfile hash` | No prior state exists for this ecosystem. | This is expected on first run — the install should execute. |
| `multiple Node lockfiles detected; skipping install` | More than one Node lockfile in the same directory. | Remove the extra lockfile(s) so only one package manager is active. |
| `multiple Python lockfiles detected; skipping install` | More than one Python lockfile in the same directory. | Remove the extra lockfile(s). |
| `python manager <name> can execute install/build code; rerun with --allow-scripts to allow` | Python managers are skipped by default for safety. | Pass `--allow-scripts` if you trust the project. |
| `Kasetto project has no local kasetto.yaml or kasetto.yml; skipping install` | Only `kasetto.lock` was found without a config file. | Add a `kasetto.yaml` or `kasetto.yml` to the project. |
| `<manager> not found on PATH` | The package manager binary is not installed or not on `PATH`. | Install the manager or ensure it is on `PATH` (e.g., activate `nvm`, `asdf`, `mise`). |

### `pupdate: error <manager> install failed: <err>`

**Cause:** The package manager command exited with a non-zero status.

**Diagnosis:**

1. Run the install command manually to see the full error output:
   ```bash
   pupdate run --dry-run
   ```
   This prints the exact command pupdate would run. Then run it yourself in the project directory.

2. Check that the manager binary is the correct version and is on `PATH`:
   ```bash
   which <manager>
   <manager> --version
   ```

**Common fixes:**

- **Node `npm ci` fails:** Ensure `package-lock.json` is in sync with `package.json`. Run `npm install` manually to regenerate the lockfile.
- **Composer fails:** Ensure `composer.lock` is valid. Run `composer install` manually.
- **Go `go mod download` fails:** Ensure `go.mod` is valid and the module cache is accessible. Run `go mod tidy` manually.
- **Cargo `cargo fetch --locked` fails:** Ensure `Cargo.lock` is in sync. Run `cargo fetch` manually.

### `pupdate: error git submodule status failed: <err>`

**Cause:** `git submodule status --recursive` failed (with a 2-second timeout).

**Diagnosis:**

- Ensure you are in a git repository.
- Ensure `git` is on `PATH`.
- Run `git submodule status --recursive` manually to see the error.
- If the command times out (large submodule trees), the error will mention the 2-second timeout.

**Fix:** Resolve the git submodule issue manually. pupdate surfaces this as a stderr error without crashing the command.

### State file warnings

#### `state file is invalid; treating as empty`

**Cause:** The `.pupdate` file contains invalid JSON.

**Fix:** Run `pupdate reset` to delete the state file. The next run will re-evaluate from scratch.

#### `state schema version mismatch: got X expected Y; treating as empty state`

**Cause:** The `.pupdate` file was written by a different (incompatible) version of pupdate.

**Fix:** Run `pupdate reset` to delete the state file. The next run will re-evaluate from scratch with the current schema.

### Config errors

#### `failed to parse <path>: <err>`

**Cause:** The config file contains invalid YAML.

**Fix:** Validate your YAML syntax. Common issues include incorrect indentation, unquoted strings with special characters, or tabs instead of spaces.

#### `failed to resolve root_directories[N]: <err>`

**Cause:** A `root_directories` entry could not be resolved to a valid path.

**Fix:** Check the path exists and is accessible. Use `~` for home directory expansion.

#### `workspace glob must be relative` / `workspace glob must not match the repository root` / `workspace glob must stay within the repository root`

**Cause:** A `workspace_globs` entry is absolute, matches the root, or escapes the root.

**Fix:** Use relative glob patterns like `apps/*` or `services/*`.

#### `folder blacklist entry must be an exact directory name, not a glob` / `folder blacklist entry must be an exact directory name, not a path`

**Cause:** A `folder_blacklist` entry contains glob characters (`*`, `?`, `[`, `]`) or path separators (`/`, `\`).

**Fix:** Use exact directory names only, e.g. `vendor` or `node_modules`.

### pupdate keeps rehashing lockfiles on every run

**Cause:** The file-identity optimization (which skips rehashing unchanged lockfiles) is only implemented on Linux and macOS. On other platforms, `file_id` and `change_time_unix_nano` are empty, so `canReuseStoredLockfileHash` always returns `false`.

**Fix:** This is expected behavior on unsupported platforms. Use Linux or macOS for the optimized path.

### Background hook lock is stuck

**Symptom:** `pupdate: skip repo (background run already active)` persists.

**Cause:** The `.pupdate.hook.lock` file was not cleaned up (e.g., the background process was killed).

**Fix:** The lock is automatically considered stale after 10 minutes, or when the PID is no longer running. To force cleanup immediately:

```bash
rm .pupdate.hook.lock
```

## Escalation

If none of the above resolves your issue:

1. Run `pupdate status` and save the output.
2. Run `pupdate run --dry-run` and save the output.
3. Run `pupdate config` and save the output.
4. Open a GitHub issue with the outputs and your platform/shell details.

For security-sensitive issues, use [GitHub Private Vulnerability Reporting](https://github.com/aaronflorey/pupdate/security) — do not open a public issue. See [`SECURITY.md`](../SECURITY.md).
