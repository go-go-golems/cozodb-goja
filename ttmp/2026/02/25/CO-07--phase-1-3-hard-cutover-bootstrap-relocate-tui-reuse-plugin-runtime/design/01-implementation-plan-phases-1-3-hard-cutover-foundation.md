---
Title: 'Implementation Plan: Phases 1-3 Hard Cutover Foundation'
Ticket: CO-07
Status: active
Topics:
    - cozodb
    - go
    - goja
    - tui
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: 2026-02-18--cozodb-extraction/cozo-relationship-js-runner/main.go
      Note: |-
        Existing goja+geppetto host wiring to extract into reusable app package
        runtime wiring source to adapt
    - Path: 2026-02-18--cozodb-extraction/cozo-relationship-js-runner/plugin_loader.go
      Note: |-
        Plugin contract and input canonicalization code to adapt
        plugin loader logic to adapt
    - Path: 2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/lib/relationship_extractor_factory.js
      Note: Existing JS extraction factory to reuse in relocated module scripts
    - Path: cozodb-goja/cmd/cozo-tui/main.go
      Note: Existing TUI entrypoint to relocate
    - Path: cozodb-goja/go.mod
      Note: Reusable library module retained as dependency target
    - Path: cozodb-goja/internal/tui/app/model.go
      Note: |-
        Existing app router and F1-F7 screen wiring
        source app router to relocate
    - Path: cozodb-goja/internal/tui/screens
      Note: Existing screen implementations for relocation parity
    - Path: cozodb-goja/internal/tui/seeddata/seed.go
      Note: Existing schema/seed logic to move into relocated module
    - Path: go.work
      Note: |-
        Workspace module composition that must include the new extraction-side TUI module
        workspace module list for relocation
ExternalSources: []
Summary: 'Detailed implementation plan for phases 1-3: extraction-side module bootstrap, F1-F7 TUI relocation, and plugin-runtime foundation with hard-cutover posture'
LastUpdated: 2026-02-25T11:58:00-05:00
WhatFor: Execution spec for building the relocation foundation before F8/F9 feature delivery
WhenToUse: Use when implementing CO-07 code changes and validating relocation parity and runtime reuse
---


# CO-07 Implementation Plan (Phases 1-3)

## 1. Objective

Deliver the foundational migration in one coherent phase bundle:

1. create extraction-side Go app module,
2. relocate existing F1-F7 TUI into that module,
3. stand up reusable plugin runtime (`internal/geppettohost` + `internal/plugins`) adapted from existing extraction runner assets.

This ticket intentionally stops before new UI features (F8/F9). It creates the platform those features require.

Hard-cutover policy for this ticket:

1. no compatibility wrapper layer,
2. no dual-run mode,
3. no new feature work in old TUI path once relocated parity passes.

---

## 2. In-Scope vs Out-of-Scope

## In scope

1. New module bootstrap under `2026-02-18--cozodb-extraction`.
2. Relocation of `cmd/cozo-tui` and `internal/tui/*` from `cozodb-goja` into extraction-side module.
3. Import rewrites from old internal paths to new module paths + `cozodb-goja/pkg/cozoapi`.
4. Plugin runtime foundation packages:
   - `internal/geppettohost`
   - `internal/plugins`
5. JS plugin script relocation into extraction-side scripts path.
6. Baseline tests and parity smoke commands for F1-F7.

## Out of scope

1. F8 screen behavior.
2. F9 screen behavior.
3. Final legacy path deletion in `cozodb-goja` (handled in CO-10).

---

## 3. Target Module Structure After CO-07

```text
2026-02-18--cozodb-extraction/cozo-extraction-tui/
  go.mod
  cmd/cozo-tui/main.go
  internal/
    tui/
      app/model.go
      screens/
        dashboard/model.go
        people/model.go
        relationships/model.go
        evolution/model.go
        network/model.go
        timeline/model.go
        query/model.go
      seeddata/seed.go
    geppettohost/
      host.go
      runtime.go
      embedder.go
      options.go
      errors.go
    plugins/
      loader.go
      types.go
      runner.go
      loader_test.go
  scripts/
    plugins/
      relation_extractor_template.js
      relation_extractor_reflective.js
      lib/*
```

Key boundary:

1. `cozo-extraction-tui` depends on `github.com/go-go-golems/cozodb-goja/pkg/cozoapi`.
2. `cozodb-goja` has no dependency on extraction-side module.

---

## 4. Phase 1 Detailed Plan: Module Bootstrap

## 4.1 Module creation and workspace wiring

Steps:

1. Create `2026-02-18--cozodb-extraction/cozo-extraction-tui`.
2. Initialize `go.mod` with stable temporary module path.
3. Add required dependencies:
   - `github.com/go-go-golems/cozodb-goja`
   - bubbletea stack (`bubbletea`, `bubbles`, `lipgloss`)
   - goja/geppetto packages needed for runtime packages
4. Add new module to root `go.work`.

Validation:

1. `go list ./...` succeeds.
2. `go test ./...` runs (initially minimal tests).

## 4.2 Build reproducibility guardrails

1. Add `Makefile` targets or script commands for:
   - `test`
   - `run-tui`
   - `lint` (if linting enabled in workspace)
2. Add one preflight script under ticket scripts with command sequence and expected output.

---

## 5. Phase 2 Detailed Plan: Relocate F1-F7 TUI

## 5.1 Copy and package layout

1. Copy `cozodb-goja/internal/tui/*` into `cozo-extraction-tui/internal/tui/*`.
2. Copy `cozodb-goja/cmd/cozo-tui/main.go` into `cozo-extraction-tui/cmd/cozo-tui/main.go`.
3. Ensure the moved command imports relocated `internal/tui/app` and `cozodb-goja/pkg/cozoapi` APIs.

## 5.2 Import and symbol rewrite

Rewrite these groups:

1. Old:
   - `github.com/go-go-golems/cozodb-goja/internal/tui/...`
2. New:
   - `<new-module>/internal/tui/...`
3. Keep library imports unchanged for bindings:
   - `github.com/go-go-golems/cozodb-goja/pkg/cozoapi`
   - `github.com/go-go-golems/cozodb-goja/pkg/cozoapi/cozocgo`

## 5.3 F1-F7 parity checklist

Run and verify each screen:

1. `F1` dashboard counts and recent rows.
2. `F2` people list + selection preview.
3. `F3` relationships list + drilldown.
4. `F4` evolution snapshots view.
5. `F5` network graph rendering.
6. `F6` timeline rendering.
7. `F7` query console execution.

Acceptance rule:

1. No behavior regressions beyond cosmetic differences.
2. No panics on screen transitions.

## 5.4 Seed and schema relocation checks

1. Confirm moved `seeddata.SeedIfEmpty` behavior matches old path for `mem` engine.
2. Confirm persistent engine behavior remains opt-in for seed flag.
3. Record command outputs in CO-07 changelog/diary if diary added later.

---

## 6. Phase 3 Detailed Plan: Plugin Runtime Foundation

## 6.1 `internal/plugins` package

Minimum types:

1. `Descriptor` (id, name, kind, apiVersion, filePath).
2. `RunInput` (transcript, prompt, profile, timeoutMs, engineOptions).
3. `ExtractionResult` (persons, relationships, behaviors, events).

Loader responsibilities:

1. discover plugin scripts from configured directory,
2. validate descriptor contract,
3. normalize run input,
4. decode run output with strict validation.

Test responsibilities:

1. reject invalid descriptor shape,
2. reject wrong api version,
3. reject empty transcript,
4. accept valid template plugin.

## 6.2 `internal/geppettohost` package

Host responsibilities:

1. instantiate goja runtime,
2. register `cozodb` and `geppetto` modules,
3. expose hooks for embedding and plugin execution,
4. map runtime panics/errors into typed Go errors.

Host options object must include:

1. script root,
2. engine defaults/profile,
3. timeout defaults,
4. optional event sink/persister settings,
5. environment sourcing policy.

## 6.3 JS script relocation and consistency

1. move/copy plugin scripts from `cozo-relationship-js-runner/scripts/*` into new module scripts path,
2. preserve relative imports in `scripts/lib/*`,
3. ensure `scriptRoot` config points to this new scripts location,
4. keep template and reflective variants both available.

## 6.4 Runtime smoke command

Add a dev command in `cozo-extraction-tui` that can execute a plugin script against a sample transcript without entering full TUI mode. This is for debugging runtime issues quickly. Implemented command path: `cmd/cozo-plugin-run/main.go`.

---

## 7. File-by-File Change List

## Module bootstrap files

1. `2026-02-18--cozodb-extraction/cozo-extraction-tui/go.mod` (new)
2. `go.work` (update `use` list)

## Relocation files

1. `.../cmd/cozo-tui/main.go` (new relocated)
2. `.../internal/tui/app/model.go` (new relocated)
3. `.../internal/tui/screens/*/model.go` (new relocated)
4. `.../internal/tui/seeddata/seed.go` (new relocated)

## Runtime foundation files

1. `.../internal/geppettohost/*.go` (new)
2. `.../internal/plugins/*.go` (new)
3. `.../internal/plugins/*_test.go` (new)
4. `.../cmd/cozo-plugin-run/main.go` (new)
5. `.../scripts/relation_extractor_*.js` (new/relocated)
6. `.../scripts/lib/*.js` (new/relocated)
7. `.../scripts/fixtures/*` (new)

---

## 8. Testing Plan for CO-07

## Unit tests

1. plugin descriptor parser and validator,
2. canonical run input validation,
3. output decoding edge cases,
4. host runtime creation and module registration sanity.

## Integration tests

1. relocated TUI boots in `mem` mode,
2. moved seed data initializes relation set,
3. plugin run with fixture transcript returns valid extraction payload.

## Manual smoke checks

1. all F1-F7 keys route correctly,
2. query screen executes basic script,
3. plugin runtime smoke command executes template plugin.

---

## 9. Failure Modes and Mitigations

1. Import rewrite misses nested screen package:
   - mitigation: compile after each directory move, not after full move only.
2. Plugin scripts fail due to incorrect `scriptRoot`:
   - mitigation: add startup log for resolved script root and plugin discovery path.
3. Hidden dependency on old internal paths remains:
   - mitigation: run `rg` check for `cozodb-goja/internal/tui` references in new module and CI gates.
4. go.work mismatch in CI/local:
   - mitigation: include module-local `go test` and explicit dependency checks.

---

## 10. Definition of Done (CO-07)

CO-07 is done when all of the following are true:

1. New extraction-side module exists and is in `go.work`.
2. Relocated TUI runs with F1-F7 parity.
3. Plugin runtime foundation packages compile and pass initial tests.
4. Plugin scripts are discoverable from relocated scripts path.
5. No new development depends on old TUI location for phases 4-5 work.

This ticket is the gate for starting CO-08 and CO-09.
