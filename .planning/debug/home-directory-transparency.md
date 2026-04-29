---
status: complete
trigger: "/gsd:debug pupdate runs in home directory and prints git/submodule + multiple Node lockfiles messages instead of staying transparent"
---

# Debug Session

## Current Focus

### hypothesis
Home-directory skipping was relying too narrowly on the process HOME environment, so a mismatched or missing HOME could let pupdate scan the actual home directory.

### next_action
Done.

## Evidence

- timestamp: 2026-04-22T00:00:00Z
  note: User reports pupdate runs in home directory and prints git/submodule plus multiple Node lockfile messages instead of remaining transparent.
- timestamp: 2026-04-22T00:10:00Z
  note: executeRun already short-circuits in the home directory, but the check used os.UserHomeDir(), which depends on HOME and can miss the actual account home when HOME is wrong.
- timestamp: 2026-04-22T00:12:00Z
  note: Added a fallback to os/user.Current().HomeDir and covered it with a regression test; targeted Go tests passed.

## Resolution

### root_cause
Home-directory detection depended only on os.UserHomeDir()/HOME, so when HOME did not resolve to the real account home, pupdate could treat the real home directory like a normal project root and emit noisy ecosystem output.

### fix
Home-directory detection now checks both HOME and the current OS user home directory, deduplicates candidates, and skips the directory as soon as either matches; a regression test locks in the fallback behavior.
