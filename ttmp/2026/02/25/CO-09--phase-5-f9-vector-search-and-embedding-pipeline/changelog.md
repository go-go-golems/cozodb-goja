# Changelog

## 2026-02-25

- Created CO-09 phase ticket for F9 vector search and embeddings.
- Added detailed implementation design document covering embedding integration, HNSW query templates, mode controls, and validation strategy.
- Added granular execution task checklist.
- Added CO-09 diary and related-file links; queued implementation behind earlier phases.
- Completed Workstreams A-B scaffold in extraction repo (commit `1aeafe4`): added `internal/tui/screens/vsearch/model.go`, F9 router/tab/resize wiring, and initial control state/input bindings (query, mode, k, ef, limit, reset).
