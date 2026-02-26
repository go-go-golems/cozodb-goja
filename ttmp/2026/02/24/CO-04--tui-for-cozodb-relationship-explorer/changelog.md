# Changelog

## 2026-02-24

- Initial workspace created


## 2026-02-24

Created TUI screen designs with ASCII mockups and Bubbletea model YAML DSL for all 9 screens: Dashboard, People Browser, Relationship Explorer, Relationship Evolution, Network Graph, Timeline, Query Console, Extraction Monitor, Vector Search

### Related Files

- cozodb-goja/ttmp/2026/02/24/CO-04--tui-for-cozodb-relationship-explorer/design/01-tui-screen-designs-and-bubbletea-models.md — Full design document with mockups and models


## 2026-02-24

Added models file hierarchy doc: 40+ files across 5 packages (app, domain, db, screens, components) with YAML model sketches for every struct/interface and a 5-phase build order

### Related Files

- cozodb-goja/ttmp/2026/02/24/CO-04--tui-for-cozodb-relationship-explorer/design/02-models-file-hierarchy.md — Complete file tree with model sketches


## 2026-02-24

Filed bug report: cozo-lib-go v0.7.5 rejects all mutations with 'write lock required for read-only query'. Confirmed at raw CGO level — not a cozoapi wrapper issue. Two repro scripts added.

### Related Files

- cozodb-goja/ttmp/2026/02/24/CO-04--tui-for-cozodb-relationship-explorer/reference/02-bug-report-cozo-lib-go-write-lock-on-all-mutations.md — Bug report with analysis and repro


## 2026-02-24

Implemented project mitigation: migrated tagged adapter from `cozo-lib-go` to `github.com/kraklabs/cie/pkg/cozodb@v0.7.20` and re-ran write-lock repro successfully. Retroactively marked repro `.go` files as tracked scripts (`//go:build ignore`) so they no longer break `go test ./...`.

### Related Files

- cozodb-goja/pkg/cozoapi/cozocgo/adapter_cozo_cgo.go — Adapter migrated to `cie/pkg/cozodb`
- cozodb-goja/ttmp/2026/02/24/CO-04--tui-for-cozodb-relationship-explorer/reference/02-bug-report-cozo-lib-go-write-lock-on-all-mutations.md — Status updated with mitigation and post-fix validation
- cozodb-goja/ttmp/2026/02/24/CO-04--tui-for-cozodb-relationship-explorer/scripts/01-cozo-write-lock-repro.go — Retained as tracked script and used to verify fix
- cozodb-goja/ttmp/2026/02/24/CO-04--tui-for-cozodb-relationship-explorer/scripts/02-cozo-raw-cgo-repro.go — Retained as tracked script for historical failing wrapper reproduction
- cozodb-goja/ttmp/2026/02/24/CO-04--tui-for-cozodb-relationship-explorer/scripts/03-cozo-c-api-4arg-repro.go — Retained as tracked script proving 4-arg C API works

## 2026-02-25

Ticket closed

