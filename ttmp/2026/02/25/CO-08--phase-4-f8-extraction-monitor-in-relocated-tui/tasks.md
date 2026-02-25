# Tasks

## Workstream A: Screen Scaffold and Routing

- [x] Create `internal/tui/screens/extraction/model.go`
- [x] Add extraction screen model constructor
- [x] Add screen enum value in app router
- [x] Add `F8` key routing in app update loop
- [x] Add extraction screen view branch in app view
- [x] Add extraction screen resize branch
- [x] Add status bar tab label `[F8]Extract`

## Workstream B: Plugin Discovery and Selection UX

- [x] Add plugin discovery command on screen init
- [x] Load plugin descriptors from configured plugin directory
- [x] Sort descriptors deterministically
- [x] Add plugin overlay list model
- [x] Add plugin metadata detail pane
- [x] Add empty-state messaging when no plugins found
- [x] Add invalid-plugin diagnostics display
- [x] Implement key binding `p` for overlay toggle
- [x] Implement plugin selection update events

## Workstream C: Transcript Input UX

- [ ] Add input mode enum (`file`, `manual`)
- [ ] Add file path input component
- [ ] Add manual transcript textarea component
- [ ] Implement key binding `n` for source prompt flow
- [ ] Add mode switch command and visual indicator
- [ ] Implement transcript read/validation helper
- [ ] Validate non-empty transcript before run

## Workstream D: Async Extraction Run Flow

- [ ] Add `pluginRunStartedMsg`
- [ ] Add `pluginRunSuccessMsg`
- [ ] Add `pluginRunErrorMsg`
- [ ] Implement `runPluginCmd` with runtime host integration
- [ ] Add run-state guard to prevent concurrent runs
- [ ] Add status updates for running/success/failure
- [ ] Store `lastInput` and `lastResult` on success
- [ ] Preserve prior result on failed rerun
- [ ] Implement key binding `r` to trigger run

## Workstream E: Result Preview UX

- [ ] Add grouped result counts panel (persons/rels/behaviors/events)
- [ ] Add preview group selector state
- [ ] Implement `tab` cycling across groups
- [ ] Add preview row cursor state
- [ ] Implement cursor key navigation
- [ ] Render selected row detail pane
- [ ] Handle empty-group preview gracefully

## Workstream F: Import Preview and Validation

- [ ] Create `ImportPreview` struct
- [ ] Implement count aggregation by group
- [ ] Detect missing mandatory fields per group
- [ ] Detect duplicate keys inside extraction payload
- [ ] Compute `canImport` decision flag
- [ ] Render import preview block in UI
- [ ] Disable import key when preview has critical errors

## Workstream G: Import Execution Path

- [ ] Create `importer.go` helper
- [ ] Map persons extraction rows to relation rows
- [ ] Map relationships extraction rows to relation rows
- [ ] Map behaviors extraction rows to relation rows
- [ ] Map events extraction rows to relation rows
- [ ] Implement deterministic upsert via relation API
- [ ] Add atomic write path (preferred)
- [ ] Add fallback/explicit error summary if partial-write risk
- [ ] Add `importStartedMsg`
- [ ] Add `importSuccessMsg`
- [ ] Add `importErrorMsg`
- [ ] Implement key binding `i` for import

## Workstream H: Export Path

- [ ] Implement JSON marshaling for `lastResult`
- [ ] Add output path resolution policy
- [ ] Write export file with predictable naming
- [ ] Implement key binding `e` for export
- [ ] Add export success/failure status messaging

## Workstream I: Runtime and Error Hardening

- [ ] Handle plugin panic as surfaced run error
- [ ] Handle malformed plugin output decode error
- [ ] Handle unreadable file input errors
- [ ] Ensure no screen panic on nil `lastResult`
- [ ] Ensure error status is visible and persisted until next action

## Workstream J: Test Coverage

- [ ] Add unit tests for import preview builder
- [ ] Add unit tests for key normalization helpers
- [ ] Add screen update tests for run success path
- [ ] Add screen update tests for run error path
- [ ] Add screen update tests for import success path
- [ ] Add screen update tests for import error path
- [ ] Add integration test running fixture plugin
- [ ] Add integration test importing fixture payload to test DB

## Workstream K: Quality and Ticket Hygiene

- [ ] Run `gofmt` on all CO-08 code changes
- [ ] Run module test suite (`go test ./... -count=1`)
- [ ] Update CO-08 changelog with implementation progress
- [ ] Relate modified files to CO-08 design doc
- [ ] Run `docmgr doctor --ticket CO-08 --stale-after 30`
