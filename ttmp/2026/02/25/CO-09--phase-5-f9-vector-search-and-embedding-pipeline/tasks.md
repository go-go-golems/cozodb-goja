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

- [ ] Integrate F9 with `internal/geppettohost` embedder API
- [ ] Add async embed command message type
- [ ] Add embedding result message type
- [ ] Add embedding error message type
- [ ] Validate non-empty query input before embed
- [ ] Validate embedding vector dimension equals 384
- [ ] Add status messages for embedding lifecycle

## Workstream D: Query Builder and Execution

- [ ] Create `query_builder.go` for mode-specific Cozo templates
- [ ] Implement person mode template
- [ ] Implement relationship mode template
- [ ] Implement behavior mode template
- [ ] Implement event mode template
- [ ] Implement all-relations merged mode template
- [ ] Parameterize templates with `$q`, `$k`, `$ef`, `$limit`
- [ ] Add command to execute query against `db.ExecScript`
- [ ] Add query success message type
- [ ] Add query error message type

## Workstream E: Result Decoding and Rendering

- [ ] Create `decoder.go` for result row conversion
- [ ] Decode person mode rows
- [ ] Decode relationship mode rows
- [ ] Decode behavior mode rows
- [ ] Decode event mode rows
- [ ] Decode all-mode rows
- [ ] Normalize into common `SearchRow` view model
- [ ] Add result list rendering
- [ ] Add selected-row detail rendering
- [ ] Add cursor navigation keys

## Workstream F: Schema/Index Preconditions

- [ ] Add startup preflight for vector columns existence
- [ ] Add startup preflight for HNSW indices existence
- [ ] Add migration-needed error message if indices missing
- [ ] Add optional automatic migration hook (if phase policy allows)
- [ ] Add status hint for manual migration command path

## Workstream G: Error Handling and Resilience

- [ ] Handle missing provider/API key errors gracefully
- [ ] Handle missing index errors gracefully
- [ ] Handle query decode mismatches gracefully
- [ ] Preserve prior successful results on failed rerun
- [ ] Ensure F9 never crashes app on runtime errors

## Workstream H: Test Coverage

- [ ] Add query builder unit tests for all modes
- [ ] Add control-bounds unit tests for `k/ef/limit`
- [ ] Add row decoder unit tests per mode
- [ ] Add screen update tests for success path
- [ ] Add screen update tests for embedding error path
- [ ] Add screen update tests for query error path
- [ ] Add integration test with seeded vectors and indices
- [ ] Add env-gated live embedding smoke test

## Workstream I: Validation and Hygiene

- [ ] Run `gofmt` on F9 code changes
- [ ] Run module tests (`go test ./... -count=1`)
- [ ] Verify F9 manual smoke in TUI
- [ ] Update CO-09 changelog with progress details
- [ ] Relate key files with `docmgr doc relate`
- [ ] Run `docmgr doctor --ticket CO-09 --stale-after 30`
