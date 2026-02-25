# Tasks

## Workstream A: Bootstrap New Module (Phase 1)

- [x] Create directory `2026-02-18--cozodb-extraction/cozo-extraction-tui`
- [x] Run `go mod init` with approved module path
- [x] Add `cmd/cozo-tui` directory scaffold
- [x] Add minimal placeholder `main.go` that compiles
- [x] Add dependency `github.com/go-go-golems/cozodb-goja`
- [x] Add Bubble Tea dependencies (`bubbletea`, `bubbles`, `lipgloss`)
- [x] Add goja/geppetto dependencies needed by runtime packages
- [x] Add module to top-level `go.work`
- [x] Run `go list ./...` in new module
- [x] Run `go test ./... -count=1` in new module

## Workstream B: Relocate Existing TUI (Phase 2)

- [x] Copy `cozodb-goja/cmd/cozo-tui/main.go` into new module
- [x] Copy `cozodb-goja/internal/tui/app/model.go` into new module
- [x] Copy `cozodb-goja/internal/tui/screens/dashboard/model.go`
- [x] Copy `cozodb-goja/internal/tui/screens/people/model.go`
- [x] Copy `cozodb-goja/internal/tui/screens/relationships/model.go`
- [x] Copy `cozodb-goja/internal/tui/screens/evolution/model.go`
- [x] Copy `cozodb-goja/internal/tui/screens/network/model.go`
- [x] Copy `cozodb-goja/internal/tui/screens/timeline/model.go`
- [x] Copy `cozodb-goja/internal/tui/screens/query/model.go`
- [x] Copy `cozodb-goja/internal/tui/seeddata/seed.go`
- [x] Rewrite imports from `github.com/go-go-golems/cozodb-goja/internal/tui/...` to new module paths
- [x] Verify `cozoapi` imports remain from `github.com/go-go-golems/cozodb-goja/pkg/cozoapi`
- [x] Run `gofmt` on relocated files
- [x] Compile relocated app package tree
- [x] Run relocated TUI in `--engine mem`

## Workstream C: F1-F7 Parity Validation

- [ ] Validate F1 dashboard renders and loads data
- [ ] Validate F2 people browser list and preview
- [ ] Validate F3 relationships list and selection
- [ ] Validate F4 evolution screen navigation and data
- [ ] Validate F5 network graph render
- [ ] Validate F6 timeline render
- [ ] Validate F7 query console executes scripts
- [x] Verify screen switching hotkeys F1-F7
- [x] Verify quit flow and DB close behavior
- [ ] Document parity results in changelog

## Workstream D: Plugin Runtime Foundation (Phase 3)

- [x] Create `internal/plugins/types.go`
- [x] Create `internal/plugins/loader.go`
- [x] Create `internal/plugins/runner.go`
- [x] Port descriptor validation logic from runner `plugin_loader.go`
- [x] Port run-input canonicalization logic
- [x] Port output decoding behavior
- [x] Create `internal/geppettohost/options.go`
- [x] Create `internal/geppettohost/runtime.go`
- [x] Create `internal/geppettohost/host.go`
- [x] Create `internal/geppettohost/embedder.go`
- [x] Create `internal/geppettohost/errors.go`
- [x] Register `cozodb` module in host runtime
- [x] Register `geppetto` and `geppetto/plugins` modules in host runtime
- [x] Add per-run timeout enforcement in runtime execution
- [x] Add panic recovery wrapper around plugin execution

## Workstream E: Script Relocation and Runtime Smoke

- [x] Copy `relation_extractor_template.js` into new scripts path
- [x] Copy `relation_extractor_reflective.js` into new scripts path
- [x] Copy `scripts/lib/relationship_constants.js`
- [x] Copy `scripts/lib/relationship_parsing.js`
- [x] Copy `scripts/lib/relationship_extractor_factory.js`
- [x] Validate relative JS import paths in moved scripts
- [x] Add runtime smoke command (script runner mode) to relocated module
- [x] Execute smoke command against fixture transcript
- [x] Confirm decoded extraction payload shape

## Workstream F: Tests and Quality Gates

- [x] Add `internal/plugins/loader_test.go`
- [x] Add invalid descriptor test (missing id)
- [x] Add invalid descriptor test (wrong apiVersion)
- [x] Add invalid run input test (empty transcript)
- [x] Add valid descriptor fixture test
- [x] Add host runtime creation sanity test
- [x] Add integration smoke test for plugin run path (if feasible)
- [x] Run full module test suite
- [x] Run `go test` for `cozodb-goja/pkg/cozoapi` to ensure boundary stability

## Workstream G: Ticket Hygiene

- [x] Update CO-07 changelog with milestone notes
- [x] Keep design doc aligned with actual file paths implemented
- [x] Relate new/modified code files with `docmgr doc relate`
- [x] Run `docmgr doctor --ticket CO-07 --stale-after 30`
