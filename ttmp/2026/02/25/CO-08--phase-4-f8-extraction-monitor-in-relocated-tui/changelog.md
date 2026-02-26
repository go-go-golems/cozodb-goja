# Changelog

## 2026-02-25

- Created CO-08 phase ticket for F8 extraction monitor implementation.
- Added detailed F8 implementation design document with architecture, message flow, import pipeline, and test plan.
- Added granular workstream task checklist for execution.
- Added CO-08 diary and related-file links; queued implementation behind CO-07 completion.
- Completed Workstream A scaffold in extraction repo (commit `ebe1940`): added `internal/tui/screens/extraction/model.go` and wired F8 routing/view/resize/status tab in app router.
- Completed Workstream B scaffold in extraction repo (commit `7487462`): added async plugin discovery, deterministic descriptor sorting, overlay list/detail panes, invalid-plugin diagnostics, and `p`-driven selection events.
- Completed Workstream C scaffold in extraction repo (commit `6f646a1`): added file/manual transcript modes, source path input, manual textarea, `n` mode switching, and non-empty transcript validation before run trigger.
- Completed Workstream D scaffold in extraction repo (commit `c9d1e13`): added async run message types, runtime-host run command wiring, concurrent-run guard, status transitions, and last-input/result retention semantics.
- Completed Workstream E scaffold in extraction repo (commit `5ccb221`): added grouped result counts panel, preview group/cursor state, `tab` group cycling, row navigation, selected-row detail rendering, and empty-group handling.
- Completed Workstream F scaffold in extraction repo (commit `8ae63a5`): added `ImportPreview` aggregation/validation (missing fields, duplicates, canImport), rendered preview block, and gated import action when validation fails.
- Completed Workstream G scaffold in extraction repo (commit `afb697c`): added `importer.go`, relation-row mapping for all groups, atomic `db.Import` write path, explicit import message lifecycle, and `i` key execution flow.
- Completed Workstream H scaffold in extraction repo (commit `30ce81f`): added JSON export command, export-path naming policy under `./exports`, and export success/failure status messaging on `e`.
- Workstream I hardening status: panic/decode/file-read errors are surfaced via run/import/export status channels; nil-result guards are in place for run/import/export actions.
- Completed Workstream J test coverage in extraction repo (commit `0b54ccb`): added import-preview/normalization unit tests, run/import state transition tests, fixture-plugin integration test, and importer integration test with fake DB.
- Quality/hygiene: `gofmt` + `go test ./... -count=1` were run after each implementation slice; CO-08 design doc file relations updated; `docmgr doctor --ticket CO-08 --stale-after 30` passes.
