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
