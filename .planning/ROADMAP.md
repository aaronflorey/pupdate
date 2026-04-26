# Roadmap: pupdate

## Overview

Maintain a fast, low-latency CLI that detects supported dependency ecosystems and runs the correct safe update command automatically on directory entry.

## Phases

- [ ] **Phase 1: Core Detection and Execution** - Detect supported ecosystems and run their package manager commands safely.
- [ ] **Phase 2: Freshness and State Tracking** - Skip unnecessary work using local dependency state.
- [ ] **Phase 3: Shell Integration and Config** - Provide reliable shell hooks and user configuration.
- [ ] **Phase 4: Release and Ecosystem Expansion** - Improve releases and add support for more package ecosystems.

## Phase Details

### Phase 1: Core Detection and Execution
**Goal**: Detect dependency signals and execute the correct manager command.
**Depends on**: Nothing (first phase)
**Success Criteria** (what must be TRUE):
1. `pupdate run` detects supported dependency ecosystems in the current directory and allowed subdirectories.
2. `pupdate run` executes the right manager command for each supported ecosystem.
3. Default install behavior remains safe and easy to understand.
**Plans**: ongoing

Plans:
- [ ] 01-01: Maintain detection matrix and manager selection behavior.

### Phase 2: Freshness and State Tracking
**Goal**: Avoid repeated installs when dependency inputs are unchanged.
**Depends on**: Phase 1
**Success Criteria** (what must be TRUE):
1. Dependency input state is stored locally.
2. No-op runs skip unnecessary installs.
**Plans**: ongoing

Plans:
- [ ] 02-01: Maintain `.pupdate` state and freshness decisions.

### Phase 3: Shell Integration and Config
**Goal**: Make hook-driven usage reliable in interactive shells.
**Depends on**: Phase 2
**Success Criteria** (what must be TRUE):
1. Shell init output works for supported shells.
2. Config defaults are created automatically when needed.
**Plans**: ongoing

Plans:
- [ ] 03-01: Maintain shell hook and config flows.

### Phase 4: Release and Ecosystem Expansion
**Goal**: Expand supported ecosystems without regressing speed or safety.
**Depends on**: Phase 3
**Success Criteria** (what must be TRUE):
1. New package ecosystems can be added with targeted detection and execution updates.
2. Release automation remains intact.
**Plans**: ongoing

Plans:
- [ ] 04-01: Add and maintain additional ecosystem support.

## Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Core Detection and Execution | ongoing | In progress | - |
| 2. Freshness and State Tracking | ongoing | In progress | - |
| 3. Shell Integration and Config | ongoing | In progress | - |
| 4. Release and Ecosystem Expansion | ongoing | In progress | - |
