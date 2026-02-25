---
Title: Geppetto Extraction and Vector Search Implementation Guide
Ticket: CO-05
Status: active
Topics:
    - cozodb
    - goja
    - tui
    - go
    - geppetto
    - vectors
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: 2026-02-18--cozodb-extraction/cozo-relationship-js-runner/main.go
      Note: Proven host wiring for geppetto module registration and runtime context injection
    - Path: 2026-02-18--cozodb-extraction/cozo-relationship-js-runner/plugin_loader.go
      Note: |-
        Proven plugin descriptor validation and canonical input normalization pattern
        reference plugin descriptor validation and run normalization
    - Path: 2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/lib/relationship_extractor_factory.js
      Note: Proven extractor session construction (createBuilder + middleware + structured output)
    - Path: cozodb-goja/cmd/XXX/main.go
      Note: |-
        Current goja runtime bootstrap that must grow plugin execution + module registration
        current goja host bootstrap and extension point
    - Path: cozodb-goja/internal/tui/app/model.go
      Note: |-
        F1-F7 are wired; F8/F9 wiring pattern and status-bar integration point
        F8/F9 app-level wiring entrypoint
    - Path: cozodb-goja/internal/tui/seeddata/seed.go
      Note: |-
        Current 4-relation schema; must gain embedding columns and HNSW index bootstrap
        seed schema and index migration target
    - Path: cozodb-goja/pkg/cozoapi/module/cozodb.go
      Note: |-
        JS API surface available to plugins and TUI-side script execution
        JS Cozo API available to plugins
    - Path: geppetto/pkg/embeddings/embeddings.go
      Note: Embeddings provider interface used for TUI-side vector index population
    - Path: geppetto/pkg/embeddings/settings_factory.go
      Note: Provider factory and cache policy options for OpenAI/Ollama embeddings
    - Path: geppetto/pkg/js/modules/geppetto/module.go
      Note: |-
        Native module registration and Options contract used by host runtime
        geppetto module registration and options contract
    - Path: geppetto/pkg/js/modules/geppetto/plugins_module.go
      Note: |-
        Canonical plugin API helpers and run-input normalization semantics
        canonical extractor plugin helper APIs
ExternalSources:
    - https://docs.cozodb.org/en/latest/vector.html
Summary: 'Implementation-ready blueprint for CO-05: plugin loading, geppetto runtime wiring, embedding pipeline, Cozo HNSW schema migration, TUI F8/F9 screens, tests, and rollout runbooks'
LastUpdated: 2026-02-25T10:40:00-05:00
WhatFor: Second implementation document that complements plan 01 with concrete contracts, file-level tasks, pseudocode, and operational details
WhenToUse: Use before and during implementation of extraction monitor (F8) and vector search (F9) to avoid design drift
---


# Geppetto Extraction and Vector Search Implementation Guide (CO-05 / Document 2)

## 1. Executive Summary

This is the second implementation document for CO-05. Document 01 defines the high-level phases. This document turns that into a concrete build specification with code contracts, runtime wiring details, schema/query templates, test gates, and rollout steps.

Main recommendation:

1. Reuse the proven plugin descriptor shape from `cozo-relationship-js-runner` and the canonical helper module `require("geppetto/plugins")`.
2. Add a dedicated host runtime package in `cozodb-goja` that registers both `cozodb` and `geppetto` native modules.
3. Extend the current 4-relation seed schema with embeddings and HNSW indices, while keeping existing F1-F7 behavior stable.
4. Build F8 as a plugin-run/import workflow and F9 as a vector query workflow over relation-specific HNSW indices.
5. Gate rollout with deterministic unit tests, integration tests against `cozocgo`, and explicit environment-driven live-LLM smoke tests.

This plan intentionally avoids speculative abstractions. It is optimized for fast, auditable implementation in the existing repository layout.

---

## 2. Scope and Non-Goals

### In scope

1. JS plugin loading/execution using goja and `geppetto/plugins` descriptor helpers.
2. Geppetto host registration (`require("geppetto")`) in `cozodb-goja` runtime.
3. Embedding-aware Cozo schema + HNSW index bootstrap for `person`, `relationship`, `behavior`, `event`.
4. F8 extraction monitor screen in Bubble Tea with async execution and import preview.
5. F9 vector search screen with relation selector, query embedding, k/ef controls, and ranked results.
6. Test and operational runbooks.

### Out of scope for CO-05

1. Full canonical-entity identity-resolution pipeline from CO-03 long-term design docs.
2. Multi-tenant plugin security sandboxing (separate process/container isolation).
3. Production-grade rate-limit orchestration across many concurrent extraction jobs.
4. Full historical embedding lineage table model (we use minimal per-row model metadata first).

---

## 3. Current-State Snapshot (Evidence)

### 3.1 TUI state

`internal/tui/app/model.go` currently wires seven screens (`F1` through `F7`) and has no extraction/vector screen models. This is the insertion point for F8/F9.

### 3.2 Schema state

`internal/tui/seeddata/seed.go` defines four relations with no vector columns and no HNSW index creation statements. Existing seed flow (`SeedIfEmpty`) is the right place to append index bootstrap.

### 3.3 JS runtime state

`cmd/XXX/main.go` currently constructs a goja runtime + `require` registry and registers only the `cozodb` module. This binary is the closest existing host harness for plugin execution.

### 3.4 Cozo JS API state

`pkg/cozoapi/module/cozodb.go` already provides a strong async JS-facing API:

1. `open`, `exec`, `q`, `cq`, `atomic`, `rel`, `export`, `import`, `close`
2. relation mutation helpers (`put`, `insert`, `update`, `rm`, `del`) through `db.rel(name)`

This means plugin authors can already write DB mutations without waiting for new Cozo wrapper work.

### 3.5 Proven prior implementation we should reuse

`2026-02-18--cozodb-extraction/cozo-relationship-js-runner` already solved the hard parts that CO-05 needs:

1. Plugin descriptor validation and normalization (`plugin_loader.go`)
2. Host runtime geppetto registration and profile/engine wiring (`main.go`)
3. JS-side plugin contract helpers (`require("geppetto/plugins")`)
4. Relationship extractor factory pattern based on `gp.createBuilder(...).buildSession()`.

This prior code is the reference implementation, not just inspiration.

---

## 4. Target Architecture for CO-05

```text
                     +-----------------------------------------+
                     |           Bubble Tea App                |
                     | F8 Extraction Monitor | F9 Vector Search|
                     +-------------------+---------------------+
                                         |
                          +--------------+---------------+
                          |    Extraction/Vector Service |
                          | (plugin host + embed helper) |
                          +------+------------------------+
                                 |
        +------------------------+------------------------+
        |                                                 |
+-------v------------------+                 +------------v----------------+
| goja Runtime Host        |                 | CozoDB (cozocgo backend)    |
| - require("cozodb")      |                 | person/relationship/behavior |
| - require("geppetto")    |                 | /event + embedding columns   |
| - require("geppetto/..." )|                | + HNSW indices               |
+--------------------------+                 +-----------------------------+
```

Core design point: keep orchestration in Go and extraction logic in JS plugins. That gives fast iteration on prompts and strict host control over persistence, safety checks, and UI state.

---

## 5. Plugin Contract and Loader Design

## 5.1 Contract (canonical)

Use this descriptor contract exactly:

```javascript
const { defineExtractorPlugin, wrapExtractorRun } = require("geppetto/plugins");

module.exports = defineExtractorPlugin({
  id: "cozo.relationship-extractor.base",
  name: "Cozo Relationship Extractor (Base)",
  create(hostContext) {
    return {
      run: wrapExtractorRun((input, options) => {
        // input.transcript required
        // input.prompt/profile/timeoutMs/engineOptions optional
        // return JSON object with persons/relationships/behaviors/events
      }),
    };
  },
});
```

Why:

1. `geppetto/plugins` already enforces `apiVersion = cozo.extractor/v1` and `kind = extractor`.
2. `wrapExtractorRun` canonicalizes run input, preventing contract drift across plugins.

## 5.2 Go-side loader responsibilities

Create `pkg/plugins` with three files:

1. `types.go`
2. `loader.go`
3. `runner.go`

Recommended types:

```go
type Loader struct {
    ScriptDir string
    Host      *runtime.Host // new package described later
}

type Descriptor struct {
    ID         string
    Name       string
    APIVersion string
    Kind       string
    Path       string
}

type RunInput struct {
    Transcript    string
    Prompt        string
    Profile       string
    TimeoutMs     int
    EngineOptions map[string]any
}

type ExtractionResult struct {
    Persons       []map[string]any `json:"persons"`
    Relationships []map[string]any `json:"relationships"`
    Behaviors     []map[string]any `json:"behaviors"`
    Events        []map[string]any `json:"events"`
}
```

Validation invariants:

1. Descriptor fields must be present and non-empty.
2. `apiVersion` must be `cozo.extractor/v1`.
3. `kind` must be `extractor`.
4. `create` must return object with `run` function.
5. `run` return must decode to object; optionally accept JSON string and decode.

## 5.3 Runtime lifecycle policy

For now, create a fresh VM per run. Reason: plugin isolation and predictable memory behavior.

Later optimization (phase extension): pooled runtimes keyed by plugin path and immutable host settings.

---

## 6. Geppetto Host Wiring in Cozodb-Goja

## 6.1 New runtime package

Add `pkg/runtime/host.go` to centralize module registration.

Responsibilities:

1. Create goja runtime and require registry.
2. Register `cozodb` module (`module.New(module.DefaultOpen).Register(reg)`).
3. Register geppetto module via `gp.Register(reg, gp.Options{...})`.
4. Optionally expose helper globals (run IDs, trace IDs, plugin metadata).

Pseudocode:

```go
func NewHost(opts HostOptions) (*Host, error) {
    vm := goja.New()
    reg := require.NewRegistry()

    module.New(module.DefaultOpen).Register(reg)

    gp.Register(reg, gp.Options{
        Runner:              opts.Runner,
        GoToolRegistry:      opts.GoToolRegistry,
        DefaultEventSinks:   opts.EventSinks,
        DefaultSnapshotHook: opts.SnapshotHook,
        DefaultPersister:    opts.Persister,
        Logger:              opts.Logger,
    })

    reg.Enable(vm)
    return &Host{vm: vm, reg: reg}, nil
}
```

## 6.2 Engine configuration strategy

Use Geppetto profile/config precedence in plugins:

1. If plugin run input includes `engineOptions`, use `gp.engines.fromConfig(engineOptions)`.
2. Else if input includes profile, use `gp.engines.fromProfile(profile, {timeoutMs})`.
3. Else default to `PINOCCHIO_PROFILE`/host default.

This matches `relationship_extractor_factory.js` from prior work and keeps runtime behavior explicit.

## 6.3 Embedding provider strategy

For CO-05 we need one reliable default:

1. Model: `text-embedding-3-small`
2. Dimensions: 384
3. Provider path: Geppetto settings factory or explicit OpenAI provider

Recommended host helper:

```go
type Embedder interface {
    Embed(ctx context.Context, text string) ([]float32, error)
    EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
    Dimensions() int
    ModelID() string
}
```

Back this with `geppetto/pkg/embeddings` provider instances.

---

## 7. Cozo Schema, Indexing, and Query Plan

## 7.1 Schema migration target (keep 4 existing relations)

Modify `internal/tui/seeddata/seed.go` relation declarations to include embedding columns:

```cozoscript
:create person {
  id: String
  =>
  name: String,
  description: String,
  first_mentioned: String,
  embedding: <F32; 384>,
  embedding_model: String,
  embedding_updated_at: String
}

:create relationship {
  id: String,
  from_person: String,
  to_person: String,
  timestamp: String
  =>
  relationship_type: String,
  description: String,
  sentiment: String,
  strength: Float,
  embedding: <F32; 384>,
  embedding_model: String,
  embedding_updated_at: String
}

:create behavior {
  id: String,
  person_id: String,
  timestamp: String
  =>
  behavior_type: String,
  description: String,
  embedding: <F32; 384>,
  embedding_model: String,
  embedding_updated_at: String
}

:create event {
  id: String,
  timestamp: String
  =>
  description: String,
  embedding: <F32; 384>,
  embedding_model: String,
  embedding_updated_at: String
}
```

Note: we add model/time metadata now so re-embed decisions are explicit.

## 7.2 HNSW index bootstrap

Add these statements after relation creation:

```cozoscript
::hnsw create person:person_embedding_idx {
  dim: 384,
  m: 16,
  dtype: F32,
  fields: [embedding],
  distance: Cosine,
  ef_construction: 200
}

::hnsw create relationship:relationship_embedding_idx {
  dim: 384,
  m: 16,
  dtype: F32,
  fields: [embedding],
  distance: Cosine,
  ef_construction: 200
}

::hnsw create behavior:behavior_embedding_idx {
  dim: 384,
  m: 16,
  dtype: F32,
  fields: [embedding],
  distance: Cosine,
  ef_construction: 200
}

::hnsw create event:event_embedding_idx {
  dim: 384,
  m: 16,
  dtype: F32,
  fields: [embedding],
  distance: Cosine,
  ef_construction: 200
}
```

## 7.3 Migration behavior for existing DB files

Add helper `SeedAndMigrateIfNeeded`:

1. Check relation existence.
2. If relation missing: create from scratch.
3. If relation exists but embedding column/index missing:
   run `::columns` and `::indices`, then execute `::alter`/`::hnsw create` where needed.

Important: never drop existing data in auto-migration path.

## 7.4 Vector query templates for F9

Relation-specific ANN query (person example):

```cozoscript
?[dist, id, name, description] :=
  ~person:person_embedding_idx{
    id, name, description |
    query: vec($q),
    k: $k,
    ef: $ef,
    bind_distance: dist
  }
:order dist
```

Cross-relation unified query (recommended F9 mode):

```cozoscript
?[entity_type, id, label, snippet, dist] :=
  *{
    ?[entity_type, id, label, snippet, dist] :=
      ~person:person_embedding_idx{id, name, description | query: vec($q), k: $k, ef: $ef, bind_distance: dist},
      entity_type = "person",
      label = name,
      snippet = description;

    ?[entity_type, id, label, snippet, dist] :=
      ~relationship:relationship_embedding_idx{id, relationship_type, description | query: vec($q), k: $k, ef: $ef, bind_distance: dist},
      entity_type = "relationship",
      label = relationship_type,
      snippet = description;

    ?[entity_type, id, label, snippet, dist] :=
      ~behavior:behavior_embedding_idx{id, behavior_type, description | query: vec($q), k: $k, ef: $ef, bind_distance: dist},
      entity_type = "behavior",
      label = behavior_type,
      snippet = description;

    ?[entity_type, id, label, snippet, dist] :=
      ~event:event_embedding_idx{id, description | query: vec($q), k: $k, ef: $ef, bind_distance: dist},
      entity_type = "event",
      label = "event",
      snippet = description
  }
:order dist
:limit $limit
```

---

## 8. F8 Extraction Monitor Design

## 8.1 Screen model structure

Add `internal/tui/screens/extraction/model.go` with:

```go
type Model struct {
    db      *cozoapi.DB
    loader  *plugins.Loader
    width   int
    height  int

    pluginList   list.Model
    sourcePath   textinput.Model
    status       string
    running      bool

    lastInput    plugins.RunInput
    lastResult   *plugins.ExtractionResult
    lastError    error

    importPreview ImportPreview
}
```

`ImportPreview` should include per-relation row counts, key collisions, and missing mandatory fields.

## 8.2 Command/message flow

Messages:

1. `pluginsDiscoveredMsg`
2. `runStartedMsg`
3. `runCompletedMsg`
4. `runFailedMsg`
5. `importStartedMsg`
6. `importCompletedMsg`
7. `importFailedMsg`

Commands:

1. `discoverPluginsCmd(dir string)`
2. `runPluginCmd(pluginID string, input RunInput)`
3. `importResultCmd(result ExtractionResult)`

All long operations return `tea.Cmd` closures. UI remains responsive.

## 8.3 Import policy

Import should be idempotent and deterministic:

1. Normalize ids (trim/lowercase, `_` separators where needed).
2. Upsert with `db.rel("...").put(...)` for deterministic updates.
3. Always write embedding metadata fields.
4. For relationship rows with same `(id, from_person, to_person, timestamp)`, overwrite non-key fields.

Pseudo-flow:

```go
func ImportExtractionResult(ctx context.Context, db *cozoapi.DB, res plugins.ExtractionResult, emb Embedder) error {
    tx := []cozoapi.PreparedQuery{}
    // Build relation puts by entity type
    // Optionally generate embeddings per row text
    // Append queries and execute db.Atomic(ctx, tx, &cozoapi.AtomicOptions{Write: cozoapi.PtrBool(true)})
}
```

If one relation write fails, everything fails. No partial import.

## 8.4 F8 keyboard contract

1. `n`: choose source input file
2. `p`: open plugin list overlay
3. `r`: run extraction
4. `i`: import extraction result into DB
5. `e`: export raw extraction JSON
6. `tab`: switch panes
7. `esc`: close overlays

---

## 9. F9 Vector Search Design

## 9.1 Screen model

Add `internal/tui/screens/vsearch/model.go`:

```go
type Model struct {
    db       *cozoapi.DB
    embedder Embedder
    width    int
    height   int

    queryInput textinput.Model
    relation   string // all/person/relationship/behavior/event
    k          int
    ef         int

    running    bool
    status     string
    results    []SearchRow
    lastErr    error
}
```

`SearchRow`:

```go
type SearchRow struct {
    EntityType string
    ID         string
    Label      string
    Snippet    string
    Distance   float64
}
```

## 9.2 Query execution flow

1. User enters text query.
2. Screen calls `embedder.Embed(ctx, queryText)`.
3. Build Cozo params: `q` vector, `k`, `ef`, `limit`.
4. Execute selected CozoScript template through `db.ExecScript` or `db.Q`.
5. Decode rows and sort by distance (server already ordered; client re-check optional).

## 9.3 Tuning defaults

Recommended defaults:

1. `k = 20`
2. `ef = 80`
3. relation selector default = `all`
4. max displayed rows = 50

Guidance:

1. Increase `ef` for recall quality.
2. Keep `k <= ef` unless intentionally broad.

---

## 10. File-by-File Implementation Plan

This is the concrete order that minimizes breakage.

## Phase A: runtime and plugin core

1. Add `pkg/runtime/host.go`.
2. Add `pkg/plugins/types.go`.
3. Add `pkg/plugins/loader.go`.
4. Add `pkg/plugins/runner.go`.
5. Add tests:
   - `pkg/plugins/loader_test.go`
   - `pkg/plugins/runner_test.go`
6. Add fixture plugins under ticket scripts and repo testdata.

## Phase B: embedding and schema

1. Add embedding helper package (recommended `internal/embedding`).
2. Extend `internal/tui/seeddata/seed.go` schema + HNSW creation.
3. Add migration helper functions in same package.
4. Add schema tests (`internal/tui/seeddata/seed_test.go`) that assert columns and indices exist.

## Phase C: app services

1. Add `internal/tui/services/extraction/service.go`.
2. Add `internal/tui/services/vectorsearch/service.go`.
3. Unit-test service layers independent of bubbletea view code.

## Phase D: screens + app wiring

1. Add `internal/tui/screens/extraction/model.go`.
2. Add `internal/tui/screens/vsearch/model.go`.
3. Update `internal/tui/app/model.go`:
   - add screen enums
   - add fields
   - add init/update/view routes
   - add tab labels `F8`, `F9`
4. Add manual smoke path in `cmd/cozo-tui/main.go` to provide plugin dir and embedder config.

## Phase E: CLI/script harness and docs

1. Extend `cmd/XXX/main.go` with plugin-run mode for quick debugging.
2. Add sample extractor plugin and README instructions.
3. Keep CO-05 docs updated (tasks/changelog/diary).

---

## 11. Testing and Validation Strategy

## 11.1 Unit tests (must pass without network)

1. Plugin descriptor validation errors:
   - missing `id`
   - wrong `apiVersion`
   - missing `create`
2. Plugin run normalization:
   - empty transcript rejects
   - timeout default = `120000`
3. Schema builder asserts 4 embedding columns + 4 HNSW statements.
4. Vector query builder emits expected query string for each relation mode.

## 11.2 Integration tests (local cozocgo, no LLM key)

1. Seed DB with zero vectors.
2. Run HNSW query with known vectors.
3. Assert query executes and returns typed rows.

## 11.3 Live tests (optional, gated by env)

Set `OPENAI_API_KEY` and run:

1. plugin extraction smoke with real model
2. embedding generation smoke for F9
3. end-to-end F8 import then F9 search

Skip when key absent.

## 11.4 Regression checks for old screens

After schema change, run smoke checks for F1-F7 queries. Ensure no query assumptions break due to added columns.

---

## 12. Operational Runbook

## 12.1 Required env

1. `OPENAI_API_KEY` for live embeddings/inference
2. optional `PINOCCHIO_PROFILE` for model profile defaults

## 12.2 Startup order

1. Open DB and run `SeedAndMigrateIfNeeded`.
2. Initialize embedder (fail soft if key missing).
3. Initialize plugin loader with plugin directory.
4. Start TUI.

## 12.3 Fail-soft behavior requirements

1. If no API key:
   - F8 extraction run should show actionable error, not crash app.
   - F9 should still run over any existing embeddings.
2. If plugin load fails:
   - other plugins remain available.
   - error shown in plugin browser details pane.
3. If index missing:
   - F9 should report migration-needed message with one-key fix action.

---

## 13. Security and Defensive Design

This integration executes arbitrary JS plugin code in-process. Defensive defaults are required.

1. Only load plugins from configured directory, not arbitrary user path.
2. Reject descriptors that do not match API contract.
3. Do not expose OS-level host functions by default.
4. Time-box each plugin run with context deadline.
5. Bound max transcript size for extraction run.
6. Log plugin ID + run ID + duration for auditing.

Future hardening (post CO-05): separate plugin process boundary.

---

## 14. Reuse Checklist from 2026-02-18 Extraction Work

Reuse directly:

1. `plugin_loader.go` descriptor/meta validation pattern.
2. `canonicalizeExtractorRunInput` field defaults and required checks.
3. `defineExtractorPlugin` + `wrapExtractorRun` helper use in JS scripts.
4. `relationship_extractor_factory.js` builder composition pattern:
   - `createBuilder`
   - `withEngine`
   - `useGoMiddleware("systemPrompt", ...)`
   - structured-output config in seed turn
5. Host registration pattern with default event sinks/snapshot/persister from `main.go`.

Do not copy blindly:

1. command-specific metrics table wiring (not required for TUI MVP).
2. pinocchio profile glue that is CLI-specific unless reused intentionally.

---

## 15. Implementation Backlog with Acceptance Criteria

### Story 1: plugin runtime foundation

Done when:

1. Loader discovers valid plugins.
2. Invalid plugins are rejected with clear errors.
3. Single plugin run returns decoded `ExtractionResult`.

### Story 2: geppetto host registration

Done when:

1. JS can `require("geppetto")` and `require("geppetto/plugins")`.
2. Plugin using builder/session runs through Go host.

### Story 3: schema + migration

Done when:

1. All 4 relations have `embedding` + metadata columns.
2. All 4 HNSW indices exist.
3. Existing DBs migrate without data loss.

### Story 4: F8 extraction monitor

Done when:

1. User can pick plugin + input and run extraction.
2. Results preview shows entity counts and sample rows.
3. Import performs atomic writes.

### Story 5: F9 vector search

Done when:

1. User enters text query and sees ranked ANN results.
2. User can switch relation mode and tune `k`/`ef`.
3. Errors are displayed without TUI crash.

---

## 16. Concrete Getting-Started Commands

```bash
# 1) ensure workspace modules are active
go work use .

# 2) resolve module deps inside cozodb-goja
cd /home/manuel/workspaces/2026-02-24/cozodb-goja-init/cozodb-goja
go mod tidy

# 3) run unit tests first
go test ./... -count=1

# 4) run TUI locally (current app)
go run ./cmd/cozo-tui

# 5) run script host harness
# (extend cmd/XXX with plugin mode as part of CO-05)
go run ./cmd/XXX --script ./scripts/example.js
```

Live embedding/extraction testing:

```bash
export OPENAI_API_KEY=... 
# optional
export PINOCCHIO_PROFILE=4o-mini
```

---

## 17. Risks, Mitigations, and Open Questions

### Risk 1: In-process plugin failures can destabilize TUI

Mitigation:

1. Recover panic at run boundary.
2. Per-run VM instances.
3. Strong error-to-message mapping in screen state.

### Risk 2: Embedding model mismatch with schema dimensions

Mitigation:

1. Validate vector length before write.
2. Store model ID and dimension checks in importer.
3. Reject write when mismatch occurs.

### Risk 3: Index creation conflicts on already indexed DB

Mitigation:

1. Discover via `::indices` first.
2. Create only missing indices.

### Open question A

Should F8 plugin execution happen on the TUI process or move behind a small local RPC worker for future isolation?

### Open question B

Do we need cross-relation score normalization in `all` mode if relation text lengths differ significantly?

### Open question C

Should we keep embedding columns inline forever, or later split into dedicated embedding relations (as in CO-03 hybrid schema) once re-embedding/versioning pressure rises?

---

## 18. Appendix A: Minimal Plugin Template

```javascript
const { defineExtractorPlugin, wrapExtractorRun } = require("geppetto/plugins");
const gp = require("geppetto");

module.exports = defineExtractorPlugin({
  id: "cozo.relationship-extractor.mini",
  name: "Mini Relationship Extractor",
  create() {
    return {
      run: wrapExtractorRun((input) => {
        const engine = gp.engines.fromProfile(input.profile || "", {
          timeoutMs: input.timeoutMs,
        });

        const session = gp
          .createBuilder()
          .withEngine(engine)
          .useGoMiddleware("systemPrompt", {
            prompt: input.prompt || "Extract people, relationships, behaviors, events as JSON.",
          })
          .buildSession();

        const turn = gp.turns.newTurn({
          blocks: [gp.turns.newUserBlock(input.transcript)],
        });

        const out = session.run(turn, { timeoutMs: input.timeoutMs });
        // Parse out.blocks assistant text to JSON in real implementation.
        return {
          persons: [],
          relationships: [],
          behaviors: [],
          events: [],
        };
      }),
    };
  },
});
```

---

## 19. Appendix B: Import Query Sketches

Person upsert via relation API:

```go
rows := []map[string]cozoapi.CozoValue{
  {
    "id":                   "sarah_martinez",
    "name":                 "Sarah Martinez",
    "description":          "Senior scientist and team lead",
    "first_mentioned":      "2023-01",
    "embedding":            embeddingVec,
    "embedding_model":      "text-embedding-3-small",
    "embedding_updated_at": time.Now().Format(time.RFC3339),
  },
}
_, err := db.Rel("person").Put(ctx, rows, cozoapi.RelationMutationOptions{})
```

Vector query via raw script:

```go
params := map[string]any{
  "q":     []float32{...},
  "k":     20,
  "ef":    80,
  "limit": 50,
}
res, err := db.ExecScript(ctx, script, params, nil)
```

---

## 20. Final Delivery Checklist (for implementation phase kickoff)

1. Runtime host package created and tested.
2. Plugin loader package created and tested.
3. Schema migration implemented and tested.
4. F8 and F9 screens implemented and wired.
5. End-to-end extraction import and vector search smoke-tested.
6. CO-05 docs/tasks/changelog kept current after each phase.

This document is the execution spec for that work.
