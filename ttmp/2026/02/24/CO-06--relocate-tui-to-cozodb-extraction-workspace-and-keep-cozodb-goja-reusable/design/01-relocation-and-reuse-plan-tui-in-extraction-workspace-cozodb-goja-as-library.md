---
Title: 'Relocation and Reuse Plan: TUI in Extraction Workspace, Cozodb-Goja as Library'
Ticket: CO-06
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
    - Path: 2026-02-18--cozodb-extraction/cozo-relationship-explorer/server/cozo_service.py
      Note: Prior schema/index/insertion implementation used as historical reference only
    - Path: 2026-02-18--cozodb-extraction/cozo-relationship-explorer/server/routers.ts
      Note: |-
        Existing TS->Python subprocess bridge to retire in the Go-native target architecture
        legacy subprocess bridge to retire
    - Path: 2026-02-18--cozodb-extraction/cozo-relationship-js-runner/main.go
      Note: |-
        Existing geppetto+goja host wiring and run orchestration to reuse/adapt
        geppetto host wiring reuse candidate
    - Path: 2026-02-18--cozodb-extraction/cozo-relationship-js-runner/plugin_loader.go
      Note: |-
        Existing plugin descriptor contract enforcement to reuse/adapt
        plugin contract validation reuse candidate
    - Path: 2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/lib/relationship_extractor_factory.js
      Note: |-
        Existing extraction session construction patterns for plugin runtime
        plugin extraction factory reuse candidate
    - Path: cozodb-goja/cmd/cozo-tui/main.go
      Note: |-
        Existing TUI app entrypoint that must move to extraction workspace
        current TUI entrypoint slated for relocation
    - Path: cozodb-goja/go.mod
      Note: |-
        Current reusable module path and dependency surface
        library module path and dependency surface
    - Path: cozodb-goja/internal/tui/app/model.go
      Note: |-
        Root Bubble Tea router and screen wiring currently in internal path
        internal visibility and screen routing constraints
    - Path: cozodb-goja/internal/tui/seeddata/seed.go
      Note: |-
        Current Cozo schema bootstrap and sample data seeding flow
        schema/seed baseline to move and extend
    - Path: cozodb-goja/pkg/cozoapi/module/cozodb.go
      Note: |-
        Reusable JS module API retained in bindings package
        reusable JS API boundary to retain
    - Path: go.work
      Note: |-
        Workspace-level module composition and local replace behavior
        workspace module topology and planned extraction module insertion
ExternalSources:
    - https://docs.cozodb.org/en/latest/vector.html
Summary: Detailed research and implementation plan to move TUI into the extraction workspace while preserving cozodb-goja as reusable bindings package and maximizing code reuse
LastUpdated: 2026-02-25T12:26:00-05:00
WhatFor: Execution plan for repository reorganization, dependency boundaries, and phased implementation completion of the TUI extraction/vector roadmap
WhenToUse: Use before starting the relocation and during each implementation phase to keep architecture and migration sequencing consistent
---


# CO-06 Relocation and Reuse Plan

## 1. Executive Summary

The requested direction is technically correct and strategically useful:

1. **Move the application** (TUI + extraction workflows + plugin scripts + embedding/vector-search orchestration) into `2026-02-18--cozodb-extraction`.
2. **Keep `cozodb-goja` as a reusable library** (Cozo API abstraction, JS module, adapters, backend wiring) that other tools/apps can import.
3. **Reuse existing code aggressively**, but only where contracts and runtime boundaries are already aligned.

The major architectural benefit is a clean split between:

1. **Domain app repo area**: extraction-focused product behaviors and UI.
2. **Reusable infrastructure module**: stable bindings package (`github.com/go-go-golems/cozodb-goja`).

The main technical constraint is current package visibility: the TUI code in `cozodb-goja` lives under `internal/tui/*`, which cannot be imported from a different Go module. This means relocation requires either:

1. copy-and-adapt the TUI package into the extraction workspace, or
2. refactor `cozodb-goja` to expose a non-`internal` public TUI package (less desirable for a bindings library), or
3. temporarily keep build in one module until refactor is complete.

Recommended execution path (lowest risk, highest reuse):

1. Create a new extraction-side Go app module (for example `2026-02-18--cozodb-extraction/cozo-extraction-tui`).
2. Copy TUI code from `cozodb-goja/internal/tui` into that new module and switch imports to `github.com/go-go-golems/cozodb-goja/pkg/cozoapi`.
3. Reuse/adapt plugin runtime pieces from `cozo-relationship-js-runner`.
4. Keep all Cozo bindings + JS module evolution in `cozodb-goja`.
5. Wire both modules via top-level `go.work` during active development.

---

## 2. Problem Statement and Success Criteria

### 2.1 Problem

Current state spreads implementation responsibility in a way that is becoming awkward:

1. `cozodb-goja` contains both reusable bindings and application-specific TUI code.
2. `2026-02-18--cozodb-extraction` contains extraction runner assets and historical extraction app material but not the active TUI.
3. CO-05 work (F8 extraction monitor and F9 vector search) semantically belongs with extraction workflows, plugin packs, and embedding assets.

### 2.2 Desired end-state

1. Extraction product work lives in extraction workspace.
2. `cozodb-goja` is reusable and not tied to one app.
3. TUI completion (F8/F9) can directly reuse plugin and extraction assets already in extraction workspace.

### 2.3 Success criteria

1. TUI builds and runs from extraction workspace path.
2. TUI imports `cozodb-goja` as module dependency for Cozo/JS bindings.
3. Existing F1-F7 functionality remains intact after move.
4. F8/F9 completion can proceed using local extraction plugin assets with minimal duplication.
5. `cozodb-goja` remains independently testable and usable by at least one non-TUI program.

---

## 3. Current-State Research Findings

## 3.1 Module and workspace boundaries

### Evidence

1. Top-level `go.work` currently uses:
   - `./cozodb-goja`
   - `./geppetto`
   - `./go-go-goja`
   - `./goja`
2. `cozodb-goja/go.mod` module path is `github.com/go-go-golems/cozodb-goja`.
3. `2026-02-18--cozodb-extraction` currently has **no** root `go.mod` and **no** extraction-local `go.work`.

### Implication

A new extraction-side TUI module must be added explicitly (new `go.mod`) and included in top-level `go.work` for local development.

## 3.2 TUI code location and coupling

### Evidence

1. TUI entrypoint: `cozodb-goja/cmd/cozo-tui/main.go`.
2. TUI package tree: `cozodb-goja/internal/tui/*`.
3. Size: about **3,607 LOC** across app+7 screen models+seeddata.
4. Imports strongly reference module-internal paths:
   - `github.com/go-go-golems/cozodb-goja/internal/tui/screens/...`
   - `github.com/go-go-golems/cozodb-goja/pkg/cozoapi`

### Constraint

Because code is under `internal/`, a module in `2026-02-18--cozodb-extraction` cannot import it.

### Implication

The move is not a pure path change; it needs either copy+adapt or a visibility refactor.

## 3.3 Reusable bindings surface in `cozodb-goja`

### Evidence

1. `pkg/cozoapi` is already a clean library surface (about **2,583 LOC**).
2. Includes backends (`cozocgo`, `fakebackend`), DB/policy/atomic/template/relation helpers, and JS module wrapper in `pkg/cozoapi/module/cozodb.go`.
3. Existing tests exist for bindings and JS module integration (`*_test.go` under `pkg/cozoapi`).

### Implication

Bindings package is already shaped for reuse; we should preserve and strengthen this boundary.

## 3.4 Extraction runner assets available for reuse

### Evidence

`2026-02-18--cozodb-extraction/cozo-relationship-js-runner` has about **2,933 LOC** with reusable components:

1. `plugin_loader.go` (~212 LOC): descriptor validation, contract enforcement, input canonicalization, output decoding.
2. JS plugin libraries:
   - `relationship_extractor_factory.js` (~222 LOC)
   - `relationship_parsing.js` (~211 LOC)
   - `relationship_constants.js` (~101 LOC)
3. Host wiring in `main.go` (~730 LOC), including geppetto registration and runtime setup.

### Caveat

No `_test.go` coverage exists in this runner package currently.

### Implication

This code is highly reusable, but we should extract/refactor with tests when moving into TUI-side runtime services.

## 3.5 Legacy TS/Python explorer state

### Evidence

1. `cozo-relationship-explorer/server` (~2,935 LOC) depends on TS->Python subprocess bridges (`execFile`, `run_python310.sh`).
2. Python service uses pseudo-random embeddings and schema/string-field compromises.

### Implication

Treat this area as reference-only for schema/query ideas; do not use as runtime base for the new relocated Go TUI path.

---

## 4. Reuse Matrix (What to Keep, Adapt, or Ignore)

## 4.1 Direct reuse candidates (high confidence)

1. `cozodb-goja/pkg/cozoapi/*` for DB operations and policies.
2. `cozodb-goja/pkg/cozoapi/module/cozodb.go` for JS module availability inside plugin runtime.
3. Runner plugin contract semantics from `plugin_loader.go`.
4. JS extraction plugin libs in `cozo-relationship-js-runner/scripts/lib/*`.
5. Cozo HNSW query/index templates from extraction docs and existing CO-05 document artifacts.

Expected direct-reuse share from extraction-side Go runner assets: **50-70%**.

## 4.2 Reuse with targeted refactor

1. `cozo-relationship-js-runner/main.go` (split command/glazed plumbing from runtime host logic).
2. `run_recorder.go` and `eval_report.go` (optional phase, telemetry only).
3. TUI code migration from `internal/tui` to extraction module path (imports and package structure updates).

## 4.3 Reference-only (not recommended as runtime code)

1. `cozo-relationship-explorer/server/*.py` and `routers.ts` subprocess path.
2. Frontend TSX app in `cozo-relationship-explorer/client/src`.
3. Manus/cayley side projects unless you explicitly need graph-query functionality.

---

## 5. Architecture Options and Tradeoffs

## Option A: Copy TUI into extraction module, keep bindings in `cozodb-goja` (Recommended)

### Shape

1. Create extraction-side module: `2026-02-18--cozodb-extraction/cozo-extraction-tui`.
2. Copy `cozodb-goja/internal/tui/*` into extraction module (for example `internal/tui/*` or `pkg/tui/*` inside the new module).
3. Update imports to use `github.com/go-go-golems/cozodb-goja/pkg/cozoapi`.
4. Add plugin runtime packages adapted from `cozo-relationship-js-runner`.

### Pros

1. Fastest path that respects `internal` visibility constraints.
2. Keeps `cozodb-goja` clean as reusable bindings package.
3. Lets extraction/TUI evolve quickly in one workspace.

### Cons

1. Temporary code duplication until old TUI path is removed.
2. Requires immediate hard-cut removal in `cozodb-goja` once relocation parity passes.

## Option B: Refactor TUI out of `internal` in `cozodb-goja`, import from extraction module

### Shape

1. Move TUI code to public package path in `cozodb-goja` (for example `pkg/tuiapp`).
2. Build extraction-side wrapper app importing `cozodb-goja/pkg/tuiapp`.

### Pros

1. Avoids temporary copy duplication.
2. One TUI codebase.

### Cons

1. Makes TUI part of library module’s public API, which is contrary to your “bindings package” intent.
2. Creates tighter coupling between library and one app style.

## Option C: Keep TUI in `cozodb-goja`, only move plugins/assets

### Pros

1. Low immediate migration effort.

### Cons

1. Does not satisfy your structural objective (“entity extraction + plugins + embeddings in one place”).
2. Leaves long-term ownership boundary blurry.

## Recommendation

Choose **Option A** now with a hard cutover:

1. Move and finish TUI in extraction workspace.
2. Remove old TUI paths from `cozodb-goja` immediately after relocated parity validation.

---

## 6. Target Architecture (Post-Relocation)

```text
workspace root: /home/manuel/workspaces/2026-02-24/cozodb-goja-init

1) Reusable library module
   cozodb-goja/
     pkg/cozoapi/...            # stays reusable
     pkg/cozoapi/module/...     # JS require("cozodb") module
     cmd/XXX                    # optional sample REPL

2) Extraction-side app module (new)
   2026-02-18--cozodb-extraction/cozo-extraction-tui/
     cmd/cozo-tui/main.go
     internal/tui/...           # moved app/screens/seeddata
     internal/plugins/...       # adapted loader/runner
     internal/geppettohost/...  # runtime registration glue
     scripts/plugins/...        # extraction plugins + templates

3) Existing extraction runner (optional transition)
   2026-02-18--cozodb-extraction/cozo-relationship-js-runner/
     # either becomes test harness or gets gradually merged into cozo-extraction-tui runtime packages
```

Boundary rule:

1. `cozodb-goja` exports DB/JS bindings.
2. extraction module owns app UX + extraction domain orchestration.

## 6.1 `internal/geppettohost` details

The relocated app must include `internal/geppettohost` as a thin integration layer.

Responsibilities:

1. create goja runtime + `require` registry,
2. register `require(\"cozodb\")` via `cozodb-goja/pkg/cozoapi/module`,
3. register Geppetto JS integration via `gp.Register(...)`,
4. expose a small host API for extraction services:
   - `LoadPluginDescriptors(...)`
   - `RunPlugin(...)`
   - `EmbedText(...)`
5. normalize runtime and plugin errors into typed Go errors usable by TUI screens.

This package must not reimplement Geppetto internals. It only wires existing Geppetto modules into the relocated app runtime.

---

## 7. Detailed Implementation Plan

## Phase 0: Preparation and Guardrails

1. Create CO-06 baseline docs/tasks/changelog/diary (this ticket).
2. Freeze current CO-05 baseline behavior with smoke command record.
3. Ensure top-level `go.work` includes all local modules needed for development.

Deliverables:

1. Updated go.work entries.
2. baseline run commands + expected outputs recorded in diary.

## Phase 1: New extraction-side module bootstrap

### Steps

1. Create directory `2026-02-18--cozodb-extraction/cozo-extraction-tui`.
2. Initialize `go.mod` with module path (temporary private path acceptable during local development).
3. Add direct dependency on `github.com/go-go-golems/cozodb-goja`.
4. Add module to root `go.work`.

### Commands (template)

```bash
cd 2026-02-18--cozodb-extraction
mkdir -p cozo-extraction-tui/cmd/cozo-tui
cd cozo-extraction-tui
go mod init github.com/manuel/cozo-extraction-tui

go get github.com/go-go-golems/cozodb-goja@latest

cd /home/manuel/workspaces/2026-02-24/cozodb-goja-init
go work use ./2026-02-18--cozodb-extraction/cozo-extraction-tui
```

### Acceptance

1. `go list ./...` succeeds in new module.
2. import of `github.com/go-go-golems/cozodb-goja/pkg/cozoapi` compiles.

## Phase 2: Move TUI code and adapt imports

### Steps

1. Copy current TUI tree:
   - `cozodb-goja/internal/tui/app`
   - `cozodb-goja/internal/tui/screens/*`
   - `cozodb-goja/internal/tui/seeddata`
2. Place into new module under `internal/tui/...`.
3. Rewrite imports from old module path to:
   - local new-module paths for app/screens
   - `github.com/go-go-golems/cozodb-goja/pkg/cozoapi` for bindings
4. Copy/adapt `cmd/cozo-tui/main.go` into new module cmd path.

### Acceptance

1. `go run ./cmd/cozo-tui` starts with F1-F7 behavior parity.
2. mem engine + seed behavior works.

## Phase 3: Establish plugin runtime packages in extraction module

### Steps

1. Extract/adapt from `cozo-relationship-js-runner`:
   - `plugin_loader.go` into `internal/plugins/loader.go`
   - runtime glue from `main.go` into `internal/geppettohost/host.go`
2. Keep command-specific glazed/Cobra parsing out of core runtime packages.
3. Add unit tests for loader validation and run input canonicalization.

### Acceptance

1. Plugin descriptor validation tests green.
2. Minimal plugin execution returns decoded extraction payload.

## Phase 4: Wire F8 extraction monitor

### Steps

1. Implement F8 screen package in relocated TUI module.
2. Load plugin list from extraction workspace scripts dir.
3. Run plugin asynchronously and show preview counts.
4. Import extracted rows via `cozoapi` relation mutation APIs.

### Acceptance

1. User flow works: select plugin, run extraction, preview, import.
2. Errors are surfaced in UI without app crash.

## Phase 5: Wire F9 vector search

### Steps

1. Add embedding service backed by geppetto embeddings provider.
2. Update/extend seed schema and migration to include embedding columns and HNSW indices.
3. Implement relation mode selector + query text input + k/ef controls.
4. Execute HNSW Cozo queries and render ranked results.

### Detailed execution sequence

1. Build a query embedding from user input through `internal/geppettohost` embedder.
2. Validate dimension (`384`) before query dispatch.
3. Build relation-mode-specific query (`person`, `relationship`, `behavior`, `event`, `all`).
4. Execute with params (`$q`, `$k`, `$ef`, `$limit`) via `db.ExecScript`.
5. Normalize rows into an F9 result model with deterministic ordering by distance.
6. Surface all runtime/provider/index errors in the F9 status pane without app termination.

### Acceptance

1. Query text produces vector result rows.
2. Relation-specific and all-mode queries work.
3. Missing-key fail-soft message appears if embedding provider unavailable.

## Phase 6: Cleanup and boundary hardening

### Steps

1. Remove old `cozodb-goja/cmd/cozo-tui` and `cozodb-goja/internal/tui` in one hard-cut phase.
2. Update all docs and run scripts to the extraction-side TUI path only.
3. Ensure `cozodb-goja` README positions module as bindings/runtime library.
4. Verify no stale imports reference removed TUI paths.

### Acceptance

1. No active product feature work depends on old TUI path.
2. `cozodb-goja` public scope is clearly infrastructure-focused.

---

## 8. File-Level Move/Refactor Map

## 8.1 TUI relocation map

1. `cozodb-goja/cmd/cozo-tui/main.go`
   -> `2026-02-18--cozodb-extraction/cozo-extraction-tui/cmd/cozo-tui/main.go`
2. `cozodb-goja/internal/tui/app/model.go`
   -> `.../cozo-extraction-tui/internal/tui/app/model.go`
3. `cozodb-goja/internal/tui/screens/*/model.go`
   -> `.../cozo-extraction-tui/internal/tui/screens/*/model.go`
4. `cozodb-goja/internal/tui/seeddata/seed.go`
   -> `.../cozo-extraction-tui/internal/tui/seeddata/seed.go`

## 8.2 Plugin/runtime reuse map

1. `2026-02-18--cozodb-extraction/cozo-relationship-js-runner/plugin_loader.go`
   -> `.../cozo-extraction-tui/internal/plugins/loader.go`
2. Runtime host subset from `cozo-relationship-js-runner/main.go`
   -> `.../cozo-extraction-tui/internal/geppettohost/host.go`
3. JS plugin libs from `cozo-relationship-js-runner/scripts/lib/*`
   -> `.../cozo-extraction-tui/scripts/plugins/lib/*`
4. Template plugin scripts
   -> `.../cozo-extraction-tui/scripts/plugins/*.js`

## 8.3 Keep in `cozodb-goja` map

1. `pkg/cozoapi/*`
2. `pkg/cozoapi/module/*`
3. backend adapters (`cozocgo`, fake backend)
4. tests for bindings/module contracts

---

## 9. Dependency and Tooling Plan

## 9.1 Go workspace plan

Keep one top-level `go.work` for local multi-module development:

```go
use (
  ./cozodb-goja
  ./geppetto
  ./go-go-goja
  ./goja
  ./2026-02-18--cozodb-extraction/cozo-extraction-tui
)
```

## 9.2 Module dependency plan

`cozo-extraction-tui/go.mod` should explicitly require:

1. `github.com/go-go-golems/cozodb-goja`
2. bubbletea/lipgloss/bubbles (if not transitively pinned as desired)
3. geppetto/goja runtime deps needed by plugin host packages

## 9.3 Versioning rule

1. During local development: use `go.work` local module resolution.
2. For CI/release: pin tagged versions of `cozodb-goja` when available.

---

## 10. Testing and Validation Plan

## 10.1 Pre-move baseline

1. Run current `cozodb-goja/cmd/cozo-tui` mem-engine smoke.
2. Capture key interactions (F1-F7) and known outputs/queries.

## 10.2 Post-move parity tests

1. Run relocated TUI with same commands and verify F1-F7 parity.
2. Add automated tests for:
   - schema creation and seed
   - plugin descriptor validation
   - Cozo query builder helpers (F9)

## 10.3 F8/F9 integration tests

1. fixture plugin run returns expected shape.
2. import writes expected row counts.
3. vector query returns non-empty nearest-neighbor rows for seeded vectors.

## 10.4 Regression tests for library boundary

In `cozodb-goja`:

1. run existing `pkg/cozoapi` tests.
2. add at least one test to ensure no hidden dependency on removed internal TUI paths.

---

## 11. Rollout and Cutover Strategy

This plan uses hard cutover only (no dual-run and no compatibility window).

1. Relocated app reaches F1-F7 parity and passes smoke checks.
2. Active commands/docs switch to relocated path.
3. Old TUI paths are removed immediately in the next phase.
4. CI gates fail on any reintroduced references to removed paths.

---

## 12. Risks and Mitigations

## Risk 1: Import churn and package path breakage

Mitigation:

1. do path rewrites with scripted search/replace + compile after each package move.
2. move in small batches (app, one screen at a time if needed).

## Risk 2: Behavior drift in copied TUI code

Mitigation:

1. capture baseline before move.
2. parity smoke checklist for each screen.

## Risk 3: Duplicate code divergence during transition

Mitigation:

1. keep transition short and cut over hard once parity is confirmed.
2. remove old TUI paths immediately after cutover.

## Risk 4: Plugin runtime side effects in-process

Mitigation:

1. isolate plugin execution per VM.
2. enforce input/output contract and timeout bounds.

## Risk 5: Embedding provider availability

Mitigation:

1. fail-soft messaging in UI.
2. allow non-live mode for development with preseeded vectors.

---

## 13. Alternative Organization Variants

## Variant A: Add extraction app module directly under extraction root (recommended)

Pros: simplest mental model.

## Variant B: Create `apps/` and `libs/` structure at workspace root

Pros: explicit monorepo organization.

Cons: larger immediate path churn.

## Variant C: Split extraction runtime core package from TUI module

1. `cozo-extraction-core` (plugin runtime/import logic)
2. `cozo-extraction-tui` (UI only)

Pros: future reuse by CLI/web.
Cons: initial overhead.

Recommendation:

1. start with Variant A.
2. split into Variant C only if second consumer appears.

---

## 14. Proposed Initial Task Breakdown (Execution-Ready)

1. Create `cozo-extraction-tui` module and add to `go.work`.
2. Copy TUI code tree and compile F1-F7 parity.
3. Add extraction-side plugin host package from runner loader/runtime core.
4. Implement F8 screen on moved codebase.
5. Implement F9 screen with embedding + HNSW.
6. Add parity and integration tests.
7. Deprecate old TUI path in `cozodb-goja`.

Each task should include:

1. code commit,
2. diary update,
3. ticket task check-off,
4. changelog entry.

---

## 15. Open Questions Requiring Early Decision

1. Module naming: final import path for extraction-side TUI module?
2. Do we include run telemetry (`run_recorder`) in phase 1 or postpone?
3. Should plugin script directory be fixed or configurable via CLI flag/env?
4. Should embedding service support both OpenAI and local/Ollama in first relocation release?

Decide these before Phase 1 code churn starts.

---

## 16. Concrete Start Commands

```bash
# workspace root
cd /home/manuel/workspaces/2026-02-24/cozodb-goja-init

# create extraction-side module
mkdir -p 2026-02-18--cozodb-extraction/cozo-extraction-tui/cmd/cozo-tui
cd 2026-02-18--cozodb-extraction/cozo-extraction-tui
go mod init github.com/manuel/cozo-extraction-tui

go get github.com/go-go-golems/cozodb-goja@latest

# return to root and include in workspace
cd /home/manuel/workspaces/2026-02-24/cozodb-goja-init
go work use ./2026-02-18--cozodb-extraction/cozo-extraction-tui

# verify module graph
cd 2026-02-18--cozodb-extraction/cozo-extraction-tui
go list ./...
```

Parity smoke after move:

```bash
cd /home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui
go run ./cmd/cozo-tui --engine mem
```

---

## 17. Final Recommendation

Proceed with relocation now, using a **new extraction-side app module** and retaining `cozodb-goja` as the reusable bindings package. This aligns exactly with your desired ownership model and provides the cleanest foundation for finishing F8/F9 with maximal reuse of existing extraction/plugin assets.

In practical terms:

1. Move app code to extraction workspace.
2. Keep bindings in `cozodb-goja`.
3. Reuse runner/plugin pieces where contract-compatible.
4. Remove old TUI path immediately after relocated parity validation.

This gets you one coherent extraction product surface without sacrificing long-term library reuse.

---

## 18. API Boundary Contract (What Must Stay Stable)

To keep the reorganization sustainable, define an explicit contract between the extraction-side app and `cozodb-goja`.

## 18.1 Library-owned contract (`cozodb-goja`)

The following should be considered reusable API, versioned conservatively:

1. `pkg/cozoapi`:
   - `Open(...)`, `DB`, `Policy`, `PreparedQuery`, `QueryOptions`
   - relation helper methods (`Rel(name).Put/Insert/Update/Rm/Del/Get`)
2. backend adapters:
   - `pkg/cozoapi/cozocgo`
   - `pkg/cozoapi/fakebackend`
3. JS module:
   - `pkg/cozoapi/module` and the `require("cozodb")` behavior

Change policy for this layer:

1. avoid breaking signatures unless major-version bump,
2. preserve query semantics and result object shape,
3. keep tests for this layer independent of TUI code.

## 18.2 App-owned contract (`cozo-extraction-tui`)

The extraction module owns these surfaces and can iterate faster:

1. Bubble Tea screens and navigation model.
2. extraction workflow UX (plugin selection, preview, import).
3. embedding strategy defaults and runtime provider policy.
4. script/plugin discovery conventions.

Change policy for this layer:

1. no backward compatibility commitment; hard cutovers are expected by default,
2. prioritize delivery velocity and clarity of operational runbooks.

## 18.3 Cross-layer anti-patterns to avoid

1. Do not move TUI screens back into `cozodb-goja/pkg/*`.
2. Do not import extraction module internals into `cozodb-goja`.
3. Do not mix extraction-specific CLI flags into library APIs.
4. Do not let `cozodb-goja` tests depend on extraction app fixtures.

These rules prevent “accidental monolith” regression after relocation.

---

## 19. Component-by-Component Reuse Decisions

This section is an explicit keep/adapt/retire table for implementation planning.

## 19.1 `cozodb-goja/internal/tui/*` (~3.6k LOC)

Decision: **Move (copy+adapt), then retire old location**.

Rationale:

1. logic is app-specific and belongs with extraction workflows,
2. cannot be imported cross-module due to `internal/` boundary,
3. moving unblocks CO-05 F8/F9 work in the right ownership area.

Expected effort:

1. low-to-medium for code movement,
2. medium for import path rewriting and parity validation.

## 19.2 `cozodb-goja/pkg/cozoapi/*` (~2.6k LOC)

Decision: **Keep as library, do not move**.

Rationale:

1. already reusable and tested,
2. desired long-term boundary aligns with your stated goal.

Expected effort:

1. low for extraction-side integration (import-only),
2. ongoing maintenance in library repo area.

## 19.3 `cozo-relationship-js-runner/plugin_loader.go`

Decision: **Adapt into extraction module runtime package**.

Rationale:

1. canonical plugin contract enforcement already implemented,
2. high-confidence reuse with minimal business-logic changes.

Expected effort:

1. low for initial integration,
2. medium when adding test coverage and package refactoring.

## 19.4 `cozo-relationship-js-runner/main.go`

Decision: **Partial reuse only**.

Reuse:

1. goja + require registry setup pattern,
2. geppetto module registration and engine option wiring,
3. script root/module loading conventions.

Do not reuse directly:

1. glazed command definitions and CLI-only parameter sections,
2. benchmark/report command plumbing for first relocation pass.

## 19.5 `cozo-relationship-js-runner/run_recorder.go` + `eval_report.go`

Decision: **Phase-2 optional reuse**.

Rationale:

1. valuable for telemetry and reproducibility,
2. not blocking relocation itself,
3. can increase migration risk if included too early.

Plan:

1. finish relocation and F8/F9 first,
2. then integrate recorder behind feature flag.

## 19.6 JS plugin libs (`scripts/lib/*`)

Decision: **Direct reuse with minor path/config updates**.

Rationale:

1. already aligned with `geppetto/plugins` contract,
2. directly supports F8 plugin execution.

Expected effort:

1. low (mostly script-root and default prompt/config wiring).

## 19.7 Python/TS explorer stack

Decision: **Retire from active path; keep as historical reference**.

Rationale:

1. subprocess architecture conflicts with Go-native target,
2. includes pseudo-embedding behavior we do not want.

Use only for:

1. schema/query inspiration,
2. dataset/sample fixtures where useful.

---

## 20. Milestone and Staffing Plan

The timeline below assumes one primary engineer with occasional review support.

## Milestone M0: Workspace bootstrap (0.5 day)

1. create module directory and `go.mod`,
2. update `go.work`,
3. verify base import of `cozodb-goja`.

Exit criteria:

1. module compiles with placeholder main.

## Milestone M1: F1-F7 relocation parity (1.5-2 days)

1. move TUI code tree,
2. rewrite imports,
3. run parity smoke checklist.

Exit criteria:

1. moved TUI starts and all existing screens function.

## Milestone M2: Plugin runtime integration (1-1.5 days)

1. adapt loader/runtime host packages from runner,
2. add fixture plugin and loader tests.

Exit criteria:

1. extraction plugin can run from relocated app runtime service layer.

## Milestone M3: F8 extraction monitor (1.5-2 days)

1. add extraction screen and async commands,
2. preview+import flow working,
3. error states and status updates implemented.

Exit criteria:

1. end-to-end extraction import usable from TUI.

## Milestone M4: F9 vector search (1-1.5 days)

1. embedding service wiring,
2. HNSW query execution and result rendering,
3. relation selector + tuning controls.

Exit criteria:

1. semantic search flow works in app.

## Milestone M5: Stabilization and boundary cleanup (1 day)

1. docs/runbooks refreshed,
2. old TUI path deprecated,
3. optional deletion scheduled.

Exit criteria:

1. team-default run path is extraction-side app module.

Total estimate: roughly **6-8 working days** for one engineer, depending on test depth and telemetry scope.

## Review checkpoints (recommended)

1. After M1: verify no hidden coupling to old `internal/tui` path.
2. After M3: review import transaction semantics and rollback behavior.
3. After M4: review vector query correctness and performance defaults (`k`, `ef`).
4. After M5: verify old path deprecation notices and commands.

---

## 21. Implementation Readiness Checklist

Before touching production code, ensure all boxes are true:

1. CO-06 task list reflects phases M0-M5.
2. The relocation module name is agreed and documented.
3. The team agrees on the dual-run cutoff date.
4. Baseline smoke commands are captured in diary.
5. A reviewer is assigned for boundary/API changes in `cozodb-goja`.
6. F8/F9 acceptance criteria are agreed from CO-05 plan + this document.

When all six are checked, implementation can start with low ambiguity.
