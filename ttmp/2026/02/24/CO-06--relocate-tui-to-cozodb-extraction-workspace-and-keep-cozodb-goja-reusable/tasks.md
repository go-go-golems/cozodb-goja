# Tasks

## Research and Planning (2026-02-25)

- [x] Create CO-06 ticket workspace and baseline docs
- [x] Inventory current TUI, bindings, and extraction-runner code boundaries
- [x] Document relocation constraints (notably `internal/tui` import visibility)
- [x] Write 7+ page relocation and implementation plan with reuse matrix and phased execution
- [x] Write implementation diary with command-level chronology and findings
- [x] Add relocation preflight script in ticket `scripts/`
- [x] Upload CO-06 documentation bundle to reMarkable and verify remote listing
- [x] Run final `docmgr doctor --ticket CO-06 --stale-after 30` after all updates

## Implementation Phase M0: Bootstrap Extraction-Side TUI Module

- [ ] Create `2026-02-18--cozodb-extraction/cozo-extraction-tui` module (`go mod init`)
- [ ] Add module to top-level `go.work`
- [ ] Verify import of `github.com/go-go-golems/cozodb-goja/pkg/cozoapi`

## Implementation Phase M1: Relocate Existing F1-F7 TUI

- [ ] Move/copy `cmd/cozo-tui` into extraction module
- [ ] Move/copy `internal/tui/app`, `internal/tui/screens/*`, `internal/tui/seeddata`
- [ ] Rewrite imports from old module-internal paths to new module paths and `cozoapi`
- [ ] Validate F1-F7 parity in relocated app

## Implementation Phase M2: Reuse Plugin Runtime Assets

- [ ] Adapt `cozo-relationship-js-runner/plugin_loader.go` into extraction module package
- [ ] Extract reusable goja/geppetto host wiring from runner `main.go`
- [ ] Port JS plugin library scripts into extraction-side script path
- [ ] Add loader and run-input canonicalization unit tests

## Implementation Phase M3: Complete F8/F9 in Relocated App

- [ ] Implement F8 extraction monitor using relocated plugin runtime
- [ ] Implement import preview and atomic import path
- [ ] Implement F9 vector search with embedding service and HNSW queries
- [ ] Validate end-to-end extraction/import/vector-search workflow

## Implementation Phase M4: Boundary Cleanup

- [ ] Execute hard cutover removal of old TUI location in `cozodb-goja` (no compatibility layer)
- [ ] Switch all run docs/scripts to extraction-side app path only
- [ ] Enforce boundary guards to prevent reintroduction of legacy TUI paths
