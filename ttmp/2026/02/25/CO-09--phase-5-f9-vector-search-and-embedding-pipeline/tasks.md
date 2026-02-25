# Tasks

## Workstream A: Screen Scaffold and Router Wiring

- [x] Create `internal/tui/screens/vsearch/model.go`
- [x] Add F9 enum entry in app router
- [x] Add `f9` key routing in app update loop
- [x] Add F9 tab in status bar
- [x] Add F9 resize handling in app model
- [x] Add F9 view routing branch

## Workstream B: Input and Control Components

- [x] Add query text input component
- [x] Add mode selector state (`all/person/relationship/behavior/event`)
- [x] Add `k` control state with defaults
- [x] Add `ef` control state with defaults
- [x] Add `limit` state with defaults
- [x] Add mode cycle key binding
- [x] Add `k` increment/decrement bindings
- [x] Add `ef` increment/decrement bindings
- [x] Add clear/reset key binding

## Workstream C: Embedding Service Integration

- [x] Integrate F9 with `internal/geppettohost` embedder API
- [x] Add async embed command message type
- [x] Add embedding result message type
- [x] Add embedding error message type
- [x] Validate non-empty query input before embed
- [x] Validate embedding vector dimension equals 384
- [x] Add status messages for embedding lifecycle

## Workstream D: Query Builder and Execution

- [x] Create `query_builder.go` for mode-specific Cozo templates
- [x] Implement person mode template
- [x] Implement relationship mode template
- [x] Implement behavior mode template
- [x] Implement event mode template
- [x] Implement all-relations merged mode template
- [x] Parameterize templates with `$q`, `$k`, `$ef`, `$limit`
- [x] Add command to execute query against `db.ExecScript`
- [x] Add query success message type
- [x] Add query error message type

## Workstream E: Result Decoding and Rendering

- [x] Create `decoder.go` for result row conversion
- [x] Decode person mode rows
- [x] Decode relationship mode rows
- [x] Decode behavior mode rows
- [x] Decode event mode rows
- [x] Decode all-mode rows
- [x] Normalize into common `SearchRow` view model
- [x] Add result list rendering
- [x] Add selected-row detail rendering
- [x] Add cursor navigation keys

## Workstream F: Schema/Index Preconditions

- [x] Add startup preflight for vector columns existence
- [x] Add startup preflight for HNSW indices existence
- [x] Add migration-needed error message if indices missing
- [x] Add optional automatic migration hook (if phase policy allows)
- [x] Add status hint for manual migration command path

## Workstream G: Error Handling and Resilience

- [x] Handle missing provider/API key errors gracefully
- [x] Handle missing index errors gracefully
- [x] Handle query decode mismatches gracefully
- [x] Preserve prior successful results on failed rerun
- [x] Ensure F9 never crashes app on runtime errors

## Workstream H: Test Coverage

- [x] Add query builder unit tests for all modes
- [x] Add control-bounds unit tests for `k/ef/limit`
- [x] Add row decoder unit tests per mode
- [x] Add screen update tests for success path
- [x] Add screen update tests for embedding error path
- [x] Add screen update tests for query error path
- [x] Add integration test with seeded vectors and indices
- [x] Add env-gated live embedding smoke test

## Workstream I: Validation and Hygiene

- [x] Run `gofmt` on F9 code changes
- [x] Run module tests (`go test ./... -count=1`)
- [ ] Verify F9 manual smoke in TUI
  Blocked in current environment: `go run ./cmd/cozo-tui --engine mem` fails with `cozocgo backend requires build tag cozo_cgo`, and tagged builds hit the known `libcozo_c.a` archive-index linker error.
- [x] Update CO-09 changelog with progress details
- [x] Relate key files with `docmgr doc relate`
- [x] Run `docmgr doctor --ticket CO-09 --stale-after 30`
