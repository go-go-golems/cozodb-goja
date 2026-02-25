---
Title: 'Implementation Plan: Phase 4 F8 Extraction Monitor'
Ticket: CO-08
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
    - Path: 2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/geppettohost/host.go
      Note: Runtime host used for plugin execution and embeddings
    - Path: 2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/plugins/loader.go
      Note: Plugin discovery and validation used by F8
    - Path: 2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/app/model.go
      Note: App-level F8 routing and key handling integration point
    - Path: 2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/seeddata/seed.go
      Note: Schema assumptions for import destination relations
    - Path: 2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/lib/relationship_constants.js
      Note: |-
        Extraction payload schema conventions
        extraction payload shape conventions
    - Path: cozodb-goja/pkg/cozoapi/relation.go
      Note: |-
        Relation-level mutation methods used for import flow
        relation mutation semantics for import flow
    - Path: cozodb-goja/ttmp/2026/02/24/CO-05--extraction-pipeline-and-vector-search-for-tui/design/02-geppetto-extraction-and-vector-search-implementation-guide.md
      Note: phase-4 context and prior design details
ExternalSources: []
Summary: 'Detailed implementation plan for F8 extraction monitor: plugin UX, async execution, preview/diff, atomic import flow, and failure-safe behavior'
LastUpdated: 2026-02-25T12:08:00-05:00
WhatFor: Execution spec for implementing extraction monitor behavior in relocated TUI
WhenToUse: Use while building and validating phase 4
---


# CO-08 Implementation Plan (Phase 4 / F8 Extraction Monitor)

## 1. Objective

Implement the F8 Extraction Monitor screen in the relocated TUI so users can:

1. discover and select extraction plugins,
2. provide transcript input,
3. run extraction asynchronously,
4. preview extracted entities before write,
5. import results into Cozo relations with deterministic behavior.

Hard-cutover assumptions:

1. F8 exists only in relocated app module,
2. no fallback integration with legacy `cozodb-goja/internal/tui` paths,
3. plugin runtime comes from CO-07 foundational packages.

---

## 2. Functional Requirements

## 2.1 User-visible capabilities

1. Plugin browser overlay (`p`) with list + metadata panel.
2. Input source selection (`n`) supporting file path and in-memory text buffer mode.
3. Extraction run command (`r`) with visible running status and elapsed time.
4. Result summary panel grouped by entity class:
   - persons
   - relationships
   - behaviors
   - events
5. Detailed preview table/pane for selected group.
6. Import action (`i`) using deterministic write semantics.
7. JSON export action (`e`) for last extraction payload.

## 2.2 Non-functional requirements

1. No UI freeze during plugin execution.
2. Plugin errors must be shown as structured status, not panic traces.
3. Import writes must be transaction-safe from app perspective (no silent partial success).
4. Input normalization must be strict and auditable.

---

## 3. Screen Architecture

## 3.1 Package layout

Create:

1. `internal/tui/screens/extraction/model.go`
2. `internal/tui/screens/extraction/view.go` (optional split)
3. `internal/tui/screens/extraction/messages.go`
4. `internal/tui/screens/extraction/commands.go`
5. `internal/tui/screens/extraction/importer.go`

If code style in current TUI keeps single-file models, collapse into one file but preserve internal logical separation via section comments.

## 3.2 Model shape

Recommended model fields:

```go
type Model struct {
    db       *cozoapi.DB
    loader   *plugins.Loader
    host     *geppettohost.Host

    width    int
    height   int

    // state
    running          bool
    status           string
    lastErr          error
    startedAt        time.Time

    // plugin/input controls
    pluginItems      []plugins.Descriptor
    selectedPluginID string
    pluginOverlay    bool
    sourcePath       string
    inputMode        InputMode // file|manual
    manualInput      textarea.Model

    // result/preview
    lastInput        plugins.RunInput
    lastResult       *plugins.ExtractionResult
    previewGroup     string // persons|relationships|behaviors|events
    previewCursor    int

    // import state
    importPreview    ImportPreview
    importing        bool
}
```

## 3.3 Message model

Message types:

1. `pluginsLoadedMsg{items []plugins.Descriptor}`
2. `pluginRunStartedMsg{pluginID string}`
3. `pluginRunSuccessMsg{result plugins.ExtractionResult, meta RunMeta}`
4. `pluginRunErrorMsg{err error}`
5. `importStartedMsg{}`
6. `importSuccessMsg{summary ImportSummary}`
7. `importErrorMsg{err error}`
8. `statusTickMsg{now time.Time}` (optional elapsed timer)

---

## 4. Input and Plugin Lifecycle

## 4.1 Plugin discovery flow

On screen init:

1. call async `discoverPluginsCmd` with configured plugins directory,
2. validate descriptors and sort by ID,
3. if none, show actionable status (`No plugins found in ...`).

Validation gate:

1. only include descriptors that pass contract checks,
2. collect invalid plugin errors into diagnostics view (do not crash).

## 4.2 Input normalization flow

When run is triggered:

1. resolve transcript source:
   - file mode: read and trim file contents,
   - manual mode: use textarea content.
2. enforce non-empty transcript.
3. set timeout defaults if not provided.
4. pass canonical `RunInput` to plugin runner.

Normalization checks:

1. transcript required,
2. timeout > 0 else default,
3. profile/prompt optional string trim,
4. engineOptions map copied defensively.

## 4.3 Execution flow

1. set `running=true`, status to `Running plugin ...`.
2. dispatch `runPluginCmd`.
3. on success:
   - store `lastResult`,
   - compute `importPreview`,
   - set `running=false`, status success.
4. on error:
   - set `lastErr`, status failed,
   - keep previous `lastResult` untouched.

---

## 5. Import Pipeline Design

## 5.1 Import semantics

Import should be deterministic and idempotent where possible.

Policy:

1. `person`: upsert by `id`.
2. `relationship`: upsert by existing composite key shape (`id, from_person, to_person, timestamp`).
3. `behavior`: upsert by (`id, person_id, timestamp`).
4. `event`: upsert by (`id, timestamp`).

Use relation API methods from `cozoapi`:

1. `db.Rel("person").Put(...)`
2. `db.Rel("relationship").Put(...)`
3. `db.Rel("behavior").Put(...)`
4. `db.Rel("event").Put(...)`

## 5.2 Import preview contract

Before write, compute:

1. row counts per entity type,
2. rows missing mandatory fields,
3. duplicate keys in extraction payload,
4. estimated upsert count.

User-visible preview should include this before `i` action is enabled.

## 5.3 Atomicity and failure handling

Preferred approach:

1. compile per-relation `PreparedQuery` operations,
2. execute via `db.Atomic(...)` in a write transaction.

Fallback if atomic bundling across relation helpers becomes cumbersome:

1. issue relation writes in deterministic sequence,
2. fail fast with explicit summary that partial writes may have occurred,
3. do not silently report success.

Given hard quality bar, implement true atomic batch if feasible in this phase.

## 5.4 Optional embedding-on-import

If embedding generation is available and schema supports fields:

1. generate embeddings for extracted rows before write,
2. attach embedding metadata (`embedding_model`, `embedding_updated_at`) consistently.

If unavailable:

1. import without embedding update,
2. status warns: `Imported data; embeddings pending`.

---

## 6. Key Bindings and Interaction Contract

Binding plan:

1. `n`: new input source / prompt for file path
2. `p`: toggle plugin browser overlay
3. `r`: run extraction
4. `i`: import current extraction
5. `e`: export current extraction JSON
6. `tab`: cycle preview group
7. `up/down`: navigate preview rows
8. `esc`: close overlay or cancel modal state

State guards:

1. ignore `r` if already running,
2. ignore `i` if no valid `lastResult`,
3. disable import if preview has critical validation errors.

---

## 7. Integration into App Router

Changes to app model:

1. add `screenExtraction` enum member.
2. add `extraction extraction.Model` field.
3. wire `f8` key to extraction screen.
4. include screen in resize/update/view routing.
5. include `[F8]Extract` tab in status bar.

Startup wiring:

1. initialize loader and host dependencies in app constructor or main command,
2. pass shared DB handle into extraction model.

---

## 8. Detailed Testing Plan

## 8.1 Unit tests

1. preview builder tests for row counts/validation.
2. import key normalization tests.
3. screen update flow tests for message transitions:
   - start -> success,
   - start -> error.
4. key-binding guard tests.

## 8.2 Integration tests

1. run fixture plugin against fixture transcript.
2. verify extraction payload shape persisted in model.
3. import into test DB and assert row counts for all relations.
4. verify export writes expected JSON file content.

## 8.3 Manual QA checklist

1. discover plugins with valid and invalid descriptors.
2. run extraction from file path.
3. run extraction from manual input.
4. navigate preview by group.
5. import and verify data appears in F1/F2/F3 views.

---

## 9. Failure Scenarios and Mitigation

1. Plugin runtime panic:
   - mitigation: host recovers and returns `pluginRunErrorMsg`.
2. Malformed plugin output:
   - mitigation: strict decode failure with user-facing status.
3. Transcript file not readable:
   - mitigation: validation failure shown in status line.
4. Import collision anomalies:
   - mitigation: preview highlights duplicates before write.
5. DB write errors:
   - mitigation: explicit import error summary with relation context.

---

## 10. Definition of Done (CO-08)

CO-08 is complete when:

1. F8 screen is routed and usable via `F8`.
2. User can run a plugin and inspect grouped extraction results.
3. Import path writes extracted entities into Cozo relations safely.
4. JSON export works for the last successful run.
5. Unit/integration tests pass for F8 critical paths.

This ticket unblocks CO-09 by ensuring production of data that can be searched semantically.
