---
Title: Implementation Plan for Extraction Pipeline and Vector Search
Ticket: CO-05
Status: active
Topics:
    - cozodb
    - goja
    - tui
    - go
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cozodb-goja/pkg/cozoapi/module/cozodb.go
      Note: Existing Goja CozoDB module — plugin loader builds on this
    - Path: cozodb-goja/internal/tui/seeddata/seed.go
      Note: Schema definitions to extend with embedding columns
    - Path: cozodb-goja/internal/tui/app/model.go
      Note: App shell — wire F8/F9 screens here
    - Path: cozodb-goja/cmd/XXX/main.go
      Note: Existing REPL — extend with plugin execution mode
ExternalSources:
    - https://docs.cozodb.org/en/latest/vector.html
Summary: "5-phase build plan: plugin loader, geppetto module, schema migration, extraction monitor screen, vector search screen"
LastUpdated: 2026-02-24T23:00:00-05:00
WhatFor: "Step-by-step implementation guide for completing the remaining TUI screens"
WhenToUse: "Before starting work on F8 or F9 screens"
---

# Implementation Plan: Extraction Pipeline and Vector Search

## Architecture

```
                    ┌─────────────────────────────────┐
                    │         TUI (bubbletea)          │
                    │  F8: Extraction    F9: Vector    │
                    │  Monitor           Search        │
                    └──────┬──────────────┬────────────┘
                           │              │
                    ┌──────▼──────┐  ┌────▼─────────┐
                    │ Plugin      │  │ Embedding    │
                    │ Loader      │  │ Client       │
                    │ (Goja VM)   │  │ (Geppetto)   │
                    └──────┬──────┘  └────┬─────────┘
                           │              │
                    ┌──────▼──────────────▼────────────┐
                    │         CozoDB Backend            │
                    │  4 relations + HNSW indices       │
                    │  person, relationship,            │
                    │  behavior, event                  │
                    │  + embedding: <F32; 384>          │
                    └──────────────────────────────────┘
```

## Phase 1: Plugin Loader

**Goal:** Load, validate, and execute JS extraction plugins from Go.

**What to build:**

1. `pkg/plugins/loader.go` — Plugin discovery and validation
   - Scan a directory for `.js` files
   - Load each into a Goja VM with the cozodb module registered
   - Validate the `module.exports` contract:
     ```javascript
     module.exports = {
         apiVersion: "cozo.extractor/v1",
         kind: "extractor",
         id: "unique.plugin.id",
         name: "Human-Readable Name",
         create(hostContext) {
             return {
                 run(input, options) {
                     return { persons: [...], relationships: [...], ... };
                 }
             };
         }
     }
     ```
   - Return a `[]PluginDescriptor` with metadata + a callable `Run(input string) (ExtractionResult, error)`

2. `pkg/plugins/types.go` — Shared types
   ```go
   type PluginDescriptor struct {
       ID         string
       Name       string
       APIVersion string
       FilePath   string
   }

   type ExtractionResult struct {
       Persons       []map[string]any
       Relationships []map[string]any
       Behaviors     []map[string]any
       Events        []map[string]any
   }
   ```

3. `pkg/plugins/runner.go` — Execute a plugin against input text
   - Create a fresh Goja VM per run (isolation)
   - Register cozodb module + host context
   - Call `create(hostContext).run(inputText, options)`
   - Marshal the JS result back to `ExtractionResult`

**Test:** Write a simple extractor plugin in `scripts/` that hard-codes a result, load it with the loader, call run, verify the Go-side result.

**Dependencies:** None (uses existing Goja + cozodb module).

**Files to create:**
- `pkg/plugins/loader.go`
- `pkg/plugins/types.go`
- `pkg/plugins/runner.go`
- `pkg/plugins/loader_test.go`
- `scripts/test_extractor.js` (test fixture)

---

## Phase 2: Geppetto Module for Goja

**Goal:** Expose Geppetto's LLM and embedding APIs to JavaScript plugins via `require("geppetto")`.

**What to build:**

1. Add Geppetto dependency:
   ```bash
   go get github.com/go-go-golems/geppetto@latest
   ```

2. `pkg/geppettomodule/geppetto.go` — Goja native module
   - Expose to JS:
     ```javascript
     const gp = require("geppetto");

     // Embedding
     const vec = gp.embed("text to embed");  // returns Float64Array(384)

     // LLM completion (for structured extraction)
     const result = gp.complete({
         model: "gpt-4o-mini",
         messages: [{role: "user", content: "..."}],
         responseFormat: { type: "json_schema", schema: {...} }
     });
     ```
   - The Go side calls Geppetto's engine/session API
   - Embedding uses `text-embedding-3-small` (384 dims)

3. Configuration:
   - Geppetto needs an API key (from env `OPENAI_API_KEY` or config file)
   - The TUI passes a Geppetto engine instance to the module at startup
   - When no API key is set, embedding/LLM calls return errors gracefully

**Test:** Write a JS script that calls `gp.embed("hello")`, verify it returns a 384-dim vector.

**Dependencies:** Phase 1 (plugin loader registers geppetto module alongside cozodb).

**Files to create:**
- `pkg/geppettomodule/geppetto.go`
- `pkg/geppettomodule/geppetto_test.go`

**Decision needed:** Whether to use Geppetto's full engine/builder/session API or a simpler direct-call wrapper. The full API is more flexible but heavier. For embedding-only use, a thin wrapper around the OpenAI embeddings endpoint may be simpler.

---

## Phase 3: Schema Migration (Add Vectors)

**Goal:** Extend the 4-relation schema with embedding columns and HNSW indices.

**What to change:**

1. Update `internal/tui/seeddata/seed.go` — Schema with embedding columns:
   ```cozoscript
   :create person {
       id: String
       =>
       name: String,
       description: String,
       first_mentioned: String,
       embedding: <F32; 384>
   }
   ```
   Same for `relationship`, `behavior`, `event` — each gets an `embedding` column.

2. Add HNSW index creation after schema:
   ```cozoscript
   ::hnsw create person:person_embedding_idx {
       dim: 384,
       dtype: F32,
       fields: [embedding],
       distance: Cosine,
       m: 50,
       ef_construction: 200
   }
   ```
   One index per relation (4 total).

3. Update seed data to include zero-vector embeddings (placeholder):
   - Populate with actual embeddings if `OPENAI_API_KEY` is available
   - Otherwise use zero vectors (HNSW still works, just returns arbitrary order)

4. Add `--embed` flag to `cmd/cozo-seed`:
   - When set, call the embedding API for each entity's description
   - Store the resulting 384-dim vector in the `embedding` column
   - Requires `OPENAI_API_KEY`

**Migration for existing databases:**
- Add a `cmd/cozo-migrate` tool or a `--migrate` flag to `cozo-seed`
- Runs `::alter add column` for each relation to add `embedding`
- Then creates HNSW indices

**Dependencies:** Phase 2 (needs embedding client for `--embed`).

**Files to modify:**
- `internal/tui/seeddata/seed.go` — Add embedding columns and HNSW indices
- `cmd/cozo-seed/main.go` — Add `--embed` flag

**Files to create:**
- `cmd/cozo-seed/embed.go` — Embedding helper (calls Geppetto or direct OpenAI)

---

## Phase 4: Extraction Monitor Screen (F8)

**Goal:** Build the TUI screen that runs extraction plugins and imports results.

**What to build:**

1. `internal/tui/screens/extraction/model.go`
   - **Input pane:** Show source file, plugin name, status indicator
   - **Results pane:** Entity summary grouped by type (persons, rels, behaviors, events)
   - **Import preview:** Count of entities to insert, conflict detection
   - **Key bindings:**
     - `n` — New extraction: prompt for file path (textinput) + plugin selection (list)
     - `r` — Re-run last extraction
     - `i` — Import results into CozoDB (runs `:put` queries)
     - `e` — Export results as JSON to file
     - `p` — Toggle plugin browser overlay

2. The extraction runs asynchronously via `tea.Cmd`:
   ```go
   func startExtraction(loader *plugins.Loader, pluginID, inputFile string) tea.Cmd {
       return func() tea.Msg {
           result, err := loader.Run(pluginID, inputFile)
           if err != nil {
               return extractionErrorMsg{err}
           }
           return extractionCompleteMsg{result}
       }
   }
   ```

3. Import builds `:put` queries from the extraction result and executes them.

4. Wire into app model at F8.

**Dependencies:** Phase 1 (plugin loader), Phase 3 (schema with embedding columns).

**Files to create:**
- `internal/tui/screens/extraction/model.go`

**Files to modify:**
- `internal/tui/app/model.go` — Add F8 screen routing
- `cmd/cozo-tui/main.go` — Initialize plugin loader, pass to extraction screen

---

## Phase 5: Vector Search Screen (F9)

**Goal:** Build the TUI screen for semantic HNSW search.

**What to build:**

1. `internal/tui/screens/vsearch/model.go`
   - **Query input:** textinput for natural language query
   - **Index selector:** Cycle through person/relationship/behavior/event
   - **K selector:** Number of nearest neighbors (default 10)
   - **Results table:** Rank, score, entity name/description
   - **Key bindings:**
     - `enter` — Execute search (embed query, run HNSW)
     - `tab` — Cycle index type
     - `+`/`-` — Adjust k

2. Search flow:
   ```
   User types query → embed(query) → 384-dim vector
   → CozoDB: ~person:person_embedding_idx{id | query: $vec, k: 10, ef: 200, score}
   → Join with *person{id, name, description}
   → Display ranked results
   ```

3. The embedding call is a `tea.Cmd` that calls Geppetto:
   ```go
   func embedAndSearch(db *cozoapi.DB, embedder Embedder, query string, index string, k int) tea.Cmd {
       return func() tea.Msg {
           vec, err := embedder.Embed(query)
           // ... build HNSW query with vec as parameter ...
           res, err := db.ExecScript(ctx, hnswQuery, params, nil)
           return searchResultMsg{results, elapsed}
       }
   }
   ```

4. Passing vectors as CozoDB parameters:
   - CozoDB accepts vector literals as `[0.1, 0.2, ...]` in queries
   - Or use parameter binding: `$vec` with the vector as a `[]float64` param
   - Test which approach the Go adapter supports

5. Wire into app model at F9.

**Dependencies:** Phase 2 (Geppetto for embedding), Phase 3 (HNSW indices).

**Files to create:**
- `internal/tui/screens/vsearch/model.go`

**Files to modify:**
- `internal/tui/app/model.go` — Add F9 screen routing
- `cmd/cozo-tui/main.go` — Initialize embedder, pass to vsearch screen

---

## Phase Order and Dependencies

```
Phase 1: Plugin Loader
    │
    ├──→ Phase 2: Geppetto Module
    │        │
    │        ├──→ Phase 3: Schema Migration
    │        │        │
    │        │        ├──→ Phase 4: F8 Extraction Monitor
    │        │        │
    │        │        └──→ Phase 5: F9 Vector Search
    │        │
    │        └──→ Phase 5: F9 Vector Search (also needs Phase 3)
    │
    └──→ Phase 4: F8 Extraction Monitor (also needs Phase 3)
```

**Shortest path to F8:** Phase 1 → Phase 4 (skip Geppetto, use hand-written extractors)
**Shortest path to F9:** Phase 2 → Phase 3 → Phase 5

Phase 1 and Phase 2 can be developed in parallel since they're independent Go packages.

---

## Open Questions

1. **Geppetto version:** Which version of `go-go-golems/geppetto` to use? Need to check compatibility with `go-go-goja@v0.4.0`.

2. **Embedding model:** Stick with `text-embedding-3-small` (384 dims) or allow configurable models? The HNSW indices are dim-specific, so changing models means rebuilding indices.

3. **Plugin directory:** Where should extraction plugins live? Options:
   - `scripts/extractors/` in the repo
   - `~/.config/cozodb-goja/plugins/`
   - User-specified via `--plugins-dir` flag

4. **Vector parameter passing:** Need to verify that the CozoCGO adapter correctly passes `[]float64` as CozoDB vector parameters. May need adapter changes.

5. **Offline mode:** Should F9 gracefully degrade when no API key is available? Could show a message like "Set OPENAI_API_KEY to enable vector search" instead of erroring.

---

## Acceptance Criteria

### Phase 1
- [ ] `plugins.Discover("path/to/dir")` returns valid plugin descriptors
- [ ] `plugins.Run(pluginID, inputText)` returns an `ExtractionResult`
- [ ] Invalid plugins (wrong apiVersion, missing create/run) are rejected with clear errors
- [ ] Unit test with a test fixture plugin passes

### Phase 2
- [ ] `require("geppetto")` works in Goja VM
- [ ] `gp.embed("text")` returns a 384-dim float array
- [ ] `gp.complete({...})` returns structured JSON from LLM
- [ ] Graceful error when no API key is set

### Phase 3
- [ ] Schema has `embedding: <F32; 384>` on all 4 relations
- [ ] 4 HNSW indices created (one per relation)
- [ ] `cozo-seed --embed` populates embeddings from descriptions
- [ ] Existing TUI screens still work (embedding column is optional for display)

### Phase 4 (F8)
- [ ] Plugin browser shows discovered plugins
- [ ] `n` prompts for file + plugin, runs extraction
- [ ] Results pane shows entity counts by type
- [ ] `i` imports results into CozoDB
- [ ] `e` exports results as JSON
- [ ] Status indicator shows running/completed/error

### Phase 5 (F9)
- [ ] Text query is embedded via Geppetto
- [ ] HNSW search returns ranked results with cosine similarity scores
- [ ] Index selector cycles through person/relationship/behavior/event
- [ ] K parameter adjustable
- [ ] Results display entity name, score, description
