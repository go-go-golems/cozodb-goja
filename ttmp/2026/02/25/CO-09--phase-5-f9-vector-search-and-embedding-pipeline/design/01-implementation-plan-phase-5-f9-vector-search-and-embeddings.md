---
Title: 'Implementation Plan: Phase 5 F9 Vector Search and Embeddings'
Ticket: CO-09
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
    - Path: 2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/geppettohost/embedder.go
      Note: Embedding provider integration point
    - Path: 2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/screens/vsearch/model.go
      Note: Target F9 screen implementation file
    - Path: 2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/seeddata/seed.go
      Note: Schema/index migration and embedding column definitions
    - Path: 2026-02-18--cozodb-extraction/schema_design.md
      Note: |-
        HNSW query/index reference patterns
        HNSW schema and query pattern reference
    - Path: cozodb-goja/pkg/cozoapi/db.go
      Note: |-
        Query execution API used by F9 commands
        query execution integration point
    - Path: cozodb-goja/pkg/cozoapi/types.go
      Note: |-
        Query options/result decoding types
        result decoding model
ExternalSources:
    - https://docs.cozodb.org/en/latest/vector.html
Summary: 'Detailed implementation plan for phase 5: F9 vector search UX, embedding pipeline, HNSW schema/index migration, query builders, and validation'
LastUpdated: 2026-02-25T12:16:00-05:00
WhatFor: Execution spec for implementing semantic vector search in relocated TUI
WhenToUse: Use while implementing and validating CO-09
---


# CO-09 Implementation Plan (Phase 5 / F9 Vector Search)

## 1. Objective

Implement F9 semantic vector search in the relocated TUI with production-grade query flow:

1. generate embeddings from user query text,
2. execute HNSW nearest-neighbor search over Cozo relations,
3. present ranked results with mode/tuning controls,
4. fail safely when provider/index/runtime requirements are missing.

Phase 5 depends on CO-07 (relocation foundation) and benefits from CO-08 imported data flows.

---

## 2. Functional Requirements

## 2.1 F9 user capabilities

1. Enter free-text semantic query.
2. Choose search mode:
   - `all`
   - `person`
   - `relationship`
   - `behavior`
   - `event`
3. Tune ANN controls:
   - `k` (neighbors)
   - `ef` (search breadth)
4. Execute search and view ranked result rows.
5. Inspect selected row details in side pane.
6. Re-run quickly after mode/control change.

## 2.2 F9 system capabilities

1. Embedding generation through `internal/geppettohost` provider bridge.
2. HNSW index existence checks and helpful diagnostics.
3. Query template generation for each mode.
4. Cozo result decoding into strongly typed view models.

---

## 3. Data and Schema Requirements

## 3.1 Required relation columns

All target relations must include:

1. `embedding: <F32; 384>`
2. metadata fields as defined in relocation plan (`embedding_model`, `embedding_updated_at`) if enabled.

## 3.2 Required indices

Required HNSW index names:

1. `person:person_embedding_idx`
2. `relationship:relationship_embedding_idx`
3. `behavior:behavior_embedding_idx`
4. `event:event_embedding_idx`

## 3.3 Migration policy

F9 startup should validate schema/index readiness and provide deterministic behavior:

1. if missing schema/index and migration is allowed -> run migration,
2. if missing schema/index and migration disabled -> block search with actionable status.

---

## 4. Screen Architecture

## 4.1 Package layout

Create:

1. `internal/tui/screens/vsearch/model.go`
2. `internal/tui/screens/vsearch/messages.go`
3. `internal/tui/screens/vsearch/commands.go`
4. `internal/tui/screens/vsearch/query_builder.go`
5. `internal/tui/screens/vsearch/decoder.go`

## 4.2 Model shape

Recommended model:

```go
type Model struct {
    db       *cozoapi.DB
    host     *geppettohost.Host

    width    int
    height   int

    queryInput textinput.Model
    mode       SearchMode
    k          int
    ef         int
    limit       int

    running    bool
    status     string
    lastErr    error

    results    []SearchRow
    cursor     int
    lastVector []float32
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

---

## 5. Query Pipeline

## 5.1 End-to-end execution flow

1. User submits query text.
2. Validate non-empty text and `k/ef/limit` bounds.
3. Embed query text via `host.EmbedText(...)`.
4. Validate embedding dimension is `384`.
5. Build Cozo query template by selected mode.
6. Execute query with params:
   - `$q` -> embedding vector
   - `$k`
   - `$ef`
   - `$limit`
7. Decode rows into `SearchRow`.
8. Render results and status.

## 5.2 Query templates

### Person mode

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
:limit $limit
```

### Relationship mode

```cozoscript
?[dist, id, relationship_type, description] :=
  ~relationship:relationship_embedding_idx{
    id, relationship_type, description |
    query: vec($q),
    k: $k,
    ef: $ef,
    bind_distance: dist
  }
:order dist
:limit $limit
```

### All mode

Use merged union query and normalize into `(entity_type, id, label, snippet, dist)`.

## 5.3 Parameter bounds policy

Defaults:

1. `k=20`
2. `ef=80`
3. `limit=50`

Bounds:

1. `k` min 1 max 200
2. `ef` min `k` max 400
3. `limit` min 1 max 500

---

## 6. UX and Interaction Contract

Bindings:

1. `f9`: open vector search screen.
2. `enter`: execute query.
3. `m`: cycle mode.
4. `[` / `]`: decrement/increment `k`.
5. `{` / `}`: decrement/increment `ef`.
6. `ctrl+l`: clear query and results.
7. arrow keys: navigate result rows.

Status line should always show:

1. mode,
2. k,
3. ef,
4. result count,
5. error/warning state if present.

---

## 7. Error Handling Strategy

## 7.1 Provider/key errors

If embedding provider unavailable (e.g., missing API key):

1. show `Embedding provider unavailable` status,
2. keep screen interactive,
3. do not clear previous successful result set.

## 7.2 Index/schema errors

If HNSW index missing:

1. detect from Cozo error patterns or preflight checks,
2. show migration-needed status with exact remediation command.

## 7.3 Decode/query errors

If row decode fails:

1. log debug detail,
2. show user-facing status with simplified reason,
3. preserve app stability.

---

## 8. Testing Strategy

## 8.1 Unit tests

1. query builder mode tests for all 5 modes,
2. bounds normalization tests for `k/ef/limit`,
3. decoder tests for each relation result shape,
4. mode cycling and key handling tests.

## 8.2 Integration tests

1. seed/migrate DB with vector columns and indices,
2. run known query vector against seeded data,
3. assert result ordering by distance,
4. assert non-empty results in at least one mode.

## 8.3 Live-provider tests (env-gated)

1. set `OPENAI_API_KEY`,
2. run query embedding + HNSW search end-to-end,
3. skip gracefully if key missing.

---

## 9. Performance and Safety Targets

1. initial query latency target (local): < 2s for standard dataset.
2. no blocking UI stall > 100ms on command dispatch.
3. avoid storing query embeddings beyond session unless explicitly needed.
4. sanitize and bound query input length to prevent accidental overload.

---

## 10. Definition of Done (CO-09)

CO-09 completes when:

1. F9 screen is fully wired and reachable by `F9`.
2. Semantic query generates embeddings and executes HNSW search.
3. Mode controls and `k/ef` tuning work.
4. Errors are fail-soft and visible.
5. Unit and integration tests for F9 critical path pass.

CO-09 closes the semantic search gap and prepares final hard-cut cleanup in CO-10.
