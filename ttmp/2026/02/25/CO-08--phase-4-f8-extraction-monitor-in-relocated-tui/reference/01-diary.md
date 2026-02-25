---
Title: Diary
Ticket: CO-08
Status: active
Topics:
    - cozodb
    - go
    - goja
    - tui
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/app/model.go
      Note: F8 routing and tab integration
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/app/model_parity_test.go
      Note: F8 hotkey parity assertion
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/screens/extraction/model.go
      Note: CO-08 Workstream A extraction screen scaffold
    - Path: ttmp/2026/02/25/CO-08--phase-4-f8-extraction-monitor-in-relocated-tui/design/01-implementation-plan-phase-4-f8-extraction-monitor.md
      Note: Phase 4 plan referenced by diary
    - Path: ttmp/2026/02/25/CO-08--phase-4-f8-extraction-monitor-in-relocated-tui/tasks.md
      Note: Phase 4 task queue referenced by diary
ExternalSources: []
Summary: Implementation diary for CO-08
LastUpdated: 2026-02-25T12:35:00-05:00
WhatFor: Track phase 4 extraction monitor implementation
WhenToUse: Use when reviewing CO-08 execution progress
---



# Diary

## Goal

Track CO-08 implementation decisions and execution history once Phase 4 begins.

## Step 1: Diary Initialization and Queueing

CO-08 work has not started yet because active implementation is currently in CO-07 phases 1-3. This step initializes the diary so phase transitions remain traceable.

### Prompt Context

**User prompt (verbatim):** "alright, implement it now, task by task, committing as you go, keeping a diary in each ticket."

**Assistant interpretation:** Maintain a live diary in every ticket while implementing phases in sequence.

**Inferred user intent:** Ensure no ticket lacks implementation traceability.

**Commit (code):** N/A

### What I did
- Created CO-08 diary document.
- Recorded that execution is queued behind CO-07 completion.

### Why
- User requested diary coverage for each ticket.

### What worked
- Ticket now has a persistent diary file ready for Phase 4 entries.

### What didn't work
- N/A

### What I learned
- Queue-state entries reduce ambiguity about inactive tickets.

### What was tricky to build
- Keeping the diary useful without inventing work before phase start.

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- Add Step 2 when first CO-08 code commit lands.

### Code review instructions
- Confirm this file exists and is linked from CO-08 index docs.

### Technical details
- Phase dependency: CO-08 follows CO-07 runtime foundation.

## Step 2: Implement F8 Extraction Screen Scaffold and Router Wiring (Workstream A)

This step starts CO-08 implementation by adding the extraction screen skeleton and routing hooks in the TUI app shell. The goal is to establish the navigation and lifecycle contract first so later workstreams can focus on plugin discovery/run/import behaviors without touching core routing again.

The scaffold intentionally returns placeholder status text for planned key interactions (`p`, `n`, `r`, `i`, `e`) while preserving the existing F1-F7 behavior.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Begin executing CO-08 tasks in order and commit each milestone.

**Inferred user intent:** Incrementally deliver the phase with clear, reviewable units.

**Commit (code):** `ebe1940` — "co-08: scaffold F8 extraction screen and route wiring"

### What I did
- Added new extraction screen package file:
  - `internal/tui/screens/extraction/model.go`
- Implemented extraction model constructor and placeholder update/view behavior.
- Updated app router in `internal/tui/app/model.go`:
  - added `screenExtraction` enum value
  - added `extraction` model field and initialization in `New`
  - wired `f8` key to switch/init extraction screen
  - added extraction resize propagation in window-size handler
  - added extraction update dispatch in screen switch
  - added extraction content branch in `View`
  - added status tab `[F8]Extract`
- Updated app parity test to include `F8` hotkey route check.
- Ran validation:
  - `go test ./... -count=1` from `cozo-extraction-tui`

### Why
- Workstream A is the structural prerequisite for all F8 feature work in CO-08.

### What worked
- F8 navigation compiles and routes correctly.
- Existing tests continue to pass.
- New app parity assertion verifies F8 key mapping.

### What didn't work
- N/A in this step.

### What I learned
- Keeping scaffold behavior explicitly placeholder-oriented avoids accidental UX promises before plugin/runtime flow is implemented.

### What was tricky to build
- Ensuring tab/status and router updates stayed fully synchronized (enum, update, view, resize, and tab list) to avoid partially wired screen states.

### What warrants a second pair of eyes
- Visual style compatibility of extraction placeholder screen with existing TUI screens can be tuned once Workstream B/C UI pieces land.

### What should be done in the future
- Implement Workstream B plugin discovery/overlay so placeholder key paths become real state transitions.

### Code review instructions
- Start with:
  - `/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/screens/extraction/model.go`
  - `/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/app/model.go`
- Validate with:
  - `cd /home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui`
  - `go test ./... -count=1`

### Technical details
- Added tab label in status bar: `[F8]Extract`.
- Extraction screen currently uses status messages to acknowledge planned key actions without side effects.
