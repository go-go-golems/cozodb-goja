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

- [x] Add input mode enum (`file`, `manual`)
- [x] Add file path input component
- [x] Add manual transcript textarea component
- [x] Implement key binding `n` for source prompt flow
- [x] Add mode switch command and visual indicator
- [x] Implement transcript read/validation helper
- [x] Validate non-empty transcript before run

## Workstream D: Async Extraction Run Flow

- [x] Add `pluginRunStartedMsg`
- [x] Add `pluginRunSuccessMsg`
- [x] Add `pluginRunErrorMsg`
- [x] Implement `runPluginCmd` with runtime host integration
- [x] Add run-state guard to prevent concurrent runs
- [x] Add status updates for running/success/failure
- [x] Store `lastInput` and `lastResult` on success
- [x] Preserve prior result on failed rerun
- [x] Implement key binding `r` to trigger run

## Workstream E: Result Preview UX

- [x] Add grouped result counts panel (persons/rels/behaviors/events)
- [x] Add preview group selector state
- [x] Implement `tab` cycling across groups
- [x] Add preview row cursor state
- [x] Implement cursor key navigation
- [x] Render selected row detail pane
- [x] Handle empty-group preview gracefully

## Workstream F: Import Preview and Validation

- [x] Create `ImportPreview` struct
- [x] Implement count aggregation by group
- [x] Detect missing mandatory fields per group
- [x] Detect duplicate keys inside extraction payload
- [x] Compute `canImport` decision flag
- [x] Render import preview block in UI
- [x] Disable import key when preview has critical errors

## Workstream G: Import Execution Path

- [x] Create `importer.go` helper
- [x] Map persons extraction rows to relation rows
- [x] Map relationships extraction rows to relation rows
- [x] Map behaviors extraction rows to relation rows
- [x] Map events extraction rows to relation rows
- [x] Implement deterministic upsert via relation API
- [x] Add atomic write path (preferred)
- [x] Add fallback/explicit error summary if partial-write risk
- [x] Add `importStartedMsg`
- [x] Add `importSuccessMsg`
- [x] Add `importErrorMsg`
- [x] Implement key binding `i` for import

## Workstream H: Export Path

- [x] Implement JSON marshaling for `lastResult`
- [x] Add output path resolution policy
- [x] Write export file with predictable naming
- [x] Implement key binding `e` for export
- [x] Add export success/failure status messaging

## Workstream I: Runtime and Error Hardening

- [x] Handle plugin panic as surfaced run error
- [x] Handle malformed plugin output decode error
- [x] Handle unreadable file input errors
- [x] Ensure no screen panic on nil `lastResult`
- [x] Ensure error status is visible and persisted until next action

## Workstream J: Test Coverage

- [x] Add unit tests for import preview builder
- [x] Add unit tests for key normalization helpers
- [x] Add screen update tests for run success path
- [x] Add screen update tests for run error path
- [x] Add screen update tests for import success path
- [x] Add screen update tests for import error path
- [x] Add integration test running fixture plugin
- [x] Add integration test importing fixture payload to test DB

## Workstream K: Quality and Ticket Hygiene

- [x] Run `gofmt` on all CO-08 code changes
- [x] Run module test suite (`go test ./... -count=1`)
- [x] Update CO-08 changelog with implementation progress
- [x] Relate modified files to CO-08 design doc
- [x] Run `docmgr doctor --ticket CO-08 --stale-after 30`
