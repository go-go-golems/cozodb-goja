# Tasks

## Phase 1: Plugin Loader
- [ ] Create `pkg/plugins/types.go` with PluginDescriptor and ExtractionResult
- [ ] Create `pkg/plugins/loader.go` — discover and validate JS plugins
- [ ] Create `pkg/plugins/runner.go` — execute plugin in isolated Goja VM
- [ ] Write test fixture plugin `scripts/test_extractor.js`
- [ ] Write `pkg/plugins/loader_test.go`

## Phase 2: Geppetto Module
- [ ] Add geppetto dependency (`go get github.com/go-go-golems/geppetto`)
- [ ] Create `pkg/geppettomodule/geppetto.go` — expose embed() and complete() to JS
- [ ] Write integration test for embed()
- [ ] Decide: full Geppetto engine API vs thin embedding wrapper

## Phase 3: Schema Migration
- [ ] Add `embedding: <F32; 384>` columns to all 4 relations in seeddata
- [ ] Add HNSW index creation statements to seeddata
- [ ] Add `--embed` flag to `cmd/cozo-seed` for populating vectors
- [ ] Verify existing TUI screens still work with expanded schema
- [ ] Test HNSW queries work through CozoCGO adapter

## Phase 4: Extraction Monitor (F8)
- [ ] Create `internal/tui/screens/extraction/model.go`
- [ ] Implement plugin browser overlay
- [ ] Implement file/plugin input prompt
- [ ] Implement async extraction execution
- [ ] Implement import-to-CozoDB flow
- [ ] Implement JSON export
- [ ] Wire F8 into app model

## Phase 5: Vector Search (F9)
- [ ] Create `internal/tui/screens/vsearch/model.go`
- [ ] Implement query embedding via Geppetto
- [ ] Implement HNSW search query builder
- [ ] Implement index type selector
- [ ] Implement K parameter adjustment
- [ ] Wire F9 into app model
- [ ] Test with real embeddings end-to-end
