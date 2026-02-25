---
Title: Extraction Pipeline and Vector Search for TUI
Ticket: CO-05
Status: active
Topics:
    - cozodb
    - goja
    - tui
    - go
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cozodb-goja/pkg/cozoapi/module/cozodb.go
      Note: Goja module exposing CozoDB API to JavaScript (exec, q, cq, rel, atomic, export/import)
    - Path: cozodb-goja/pkg/cozoapi/module/default_open.go
      Note: Backend factory routing to fake or cozocgo backends
    - Path: cozodb-goja/cmd/XXX/main.go
      Note: Existing JS REPL/runner — needs plugin execution support
    - Path: cozodb-goja/internal/tui/app/model.go
      Note: TUI app shell — F8/F9 slots reserved for extraction monitor and vector search
    - Path: cozodb-goja/internal/tui/seeddata/seed.go
      Note: Current schema (no embedding columns) — needs vector column additions
    - Path: cozodb-goja/pkg/cozoapi/cozocgo/adapter_cozo_cgo.go
      Note: CGO adapter — may need HNSW query exposure
ExternalSources:
    - https://docs.cozodb.org/en/latest/vector.html
    - https://github.com/charmbracelet/bubbletea
    - https://github.com/go-go-golems/go-go-goja
Summary: "Add JS extraction plugin runner and HNSW vector search to the CozoDB TUI (screens F8 and F9)"
LastUpdated: 2026-02-24T23:00:00-05:00
WhatFor: "Complete the TUI with extraction pipeline integration and semantic vector search"
WhenToUse: "When implementing F8 (Extraction Monitor) or F9 (Vector Search) screens, or adding embedding support"
---

# Extraction Pipeline and Vector Search for TUI

## Overview

The CozoDB TUI (CO-04) has 7 of 9 screens implemented. The remaining two require infrastructure that doesn't exist yet:

- **F8 — Extraction Monitor**: Run JS extraction plugins from within the TUI, preview extracted entities, import into CozoDB
- **F9 — Vector Search**: Semantic search over HNSW vector indices using embeddings

Both screens depend on a chain of components: plugin loading, Geppetto LLM integration, embedding generation, and CozoDB HNSW index support.

## What Exists

| Component | Status | Location |
|-----------|--------|----------|
| Goja JS runtime | Done | `pkg/cozoapi/module/cozodb.go` |
| CozoDB module for JS | Done | `exec`, `q`, `cq`, `rel`, `atomic`, `export/import` |
| CozoCGO backend | Done | `pkg/cozoapi/cozocgo/` via `kraklabs/cie@v0.7.20` |
| TUI shell with 7 screens | Done | `internal/tui/` (F1-F7 wired) |
| Schema (4 relations, no vectors) | Done | `internal/tui/seeddata/seed.go` |
| Basic REPL | Done | `cmd/XXX/main.go` |

## What's Missing

| Component | Blocks | Notes |
|-----------|--------|-------|
| Plugin loader (`plugin_loader.go`) | F8 | Discover/validate JS plugins with `cozo.extractor/v1` contract |
| Geppetto Go dependency | F8, F9 | Not in go.mod yet |
| Geppetto Goja module | F8 | Expose `require("geppetto")` to JS plugins |
| Embedding columns in schema | F9 | Add `embedding: <F32; 384>` to all 4 relations |
| HNSW index definitions | F9 | `::hnsw create` for each relation |
| Embedding pipeline | F8, F9 | Call `text-embedding-3-small` to populate vectors |
| HNSW query support in TUI | F9 | `~relation:index{...}` search syntax |

## Documents

| Document | Purpose |
|----------|---------|
| [Implementation Plan](design/01-implementation-plan.md) | Phased build plan with dependencies and acceptance criteria |
| [CO-04 Screen Designs](../CO-04--tui-for-cozodb-relationship-explorer/design/01-tui-screen-designs-and-bubbletea-models.md) | Original ASCII mockups for F8 and F9 |

## Dependencies

- **CO-04**: TUI shell and existing screens (done)
- **Geppetto**: LLM middleware for embedding calls (external dependency, needs `go get`)
- **CozoDB v0.7.5+**: HNSW support built into libcozo_c.a (already have this)

## Status

**Not started.** This ticket picks up where CO-04 left off after implementing screens F1-F7.
