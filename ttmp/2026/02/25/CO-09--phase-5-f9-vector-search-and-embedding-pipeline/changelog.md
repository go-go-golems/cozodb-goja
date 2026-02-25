# Changelog

## 2026-02-25

- Created CO-09 phase ticket for F9 vector search and embeddings.
- Added detailed implementation design document covering embedding integration, HNSW query templates, mode controls, and validation strategy.
- Added granular execution task checklist.
- Added CO-09 diary and related-file links; queued implementation behind earlier phases.
- Completed Workstreams A-B scaffold in extraction repo (commit `1aeafe4`): added `internal/tui/screens/vsearch/model.go`, F9 router/tab/resize wiring, and initial control state/input bindings (query, mode, k, ef, limit, reset).
- Completed Workstreams C-H in extraction repo (commit `dacc589`): added geppettohost embedding API, async F9 embed/query message flow, mode-specific Cozo query builders, row decoder, result/detail rendering, vector schema/index preflight with optional index auto-create hook, and comprehensive tests including env-gated live embedding + `cozo_cgo` integration test.
- Validation: `go test ./... -count=1` passes without `cozo_cgo` tag; `go test -tags cozo_cgo ./internal/tui/screens/vsearch -count=1` currently fails due known linker issue (`libcozo_c.a: archive has no index; run ranlib`).
- Updated CO-09 tasks/diary/file-relations for Step 3 and reran `docmgr doctor --ticket CO-09 --stale-after 30` (passes).
