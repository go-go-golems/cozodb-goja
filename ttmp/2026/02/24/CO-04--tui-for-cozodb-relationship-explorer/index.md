---
Title: TUI for CozoDB Relationship Explorer
Ticket: CO-04
Status: active
Topics:
    - cozodb
    - tui
    - bubbletea
    - go
    - goja
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cozo-relationship-js-runner/main.go
      Note: Go CLI runner - TUI shares CozoDB connection and Goja integration
    - Path: cozo-relationship-js-runner/plugin_loader.go
      Note: Plugin descriptor protocol used by Extraction Monitor screen
    - Path: cozo_demo.py
      Note: Schema definitions and query examples to replicate in TUI
    - Path: cozo_advanced_demo.py
      Note: Advanced queries (PageRank, community detection) exposed in TUI
ExternalSources:
    - https://github.com/charmbracelet/bubbletea
    - https://github.com/charmbracelet/lipgloss
    - https://github.com/charmbracelet/bubbles
Summary: "Terminal UI for exploring the CozoDB entity extraction database"
LastUpdated: 2026-02-24T20:00:00-05:00
WhatFor: "Interactively browse people, relationships, events, and behaviors stored in CozoDB"
WhenToUse: "After extracting entities with the JS runner, use this TUI to explore and query the results"
---

# TUI for CozoDB Relationship Explorer

## Overview

A terminal-based interface (built with Bubbletea) for exploring the CozoDB entity extraction database. 9 screens cover browsing, querying, visualization, extraction, and semantic search.

## Screens

| # | Screen | Key | Purpose |
|---|--------|-----|---------|
| 1 | Dashboard | F1 | Entity counts, top relationships, recent events, sentiment |
| 2 | People Browser | F2 | Filterable/sortable table of all persons with preview |
| 3 | Relationship Explorer | F3 | All relationship records with temporal snapshots |
| 4 | Relationship Evolution | -- | Drill-down: strength chart + diff for a single pair |
| 5 | Network Graph | F5 | ASCII graph with community detection |
| 6 | Timeline | F4 | Chronological feed interleaving events, rels, behaviors |
| 7 | Query Console | F6 | CozoScript REPL with results table |
| 8 | Extraction Monitor | F7 | Run JS extractors, preview + import results |
| 9 | Vector Search | F8 | HNSW semantic search across all entity types |

## Documents

| Document | Purpose |
|---|---|
| [TUI Screen Designs](design/01-tui-screen-designs-and-bubbletea-models.md) | ASCII mockups + Bubbletea model YAML DSL for all 9 screens |
| [Models File Hierarchy](design/02-models-file-hierarchy.md) | Concrete Go file tree with struct/interface sketches for every file |
| [Implementation Diary](reference/01-implementation-diary.md) | Chronological record of design and build decisions |

## Tech Stack

- **Go** + **Bubbletea** (TUI framework) + **Lipgloss** (styling) + **Bubbles** (components)
- **CozoDB** (embedded database, cozo-go bindings)
- **Goja** (JS runtime for extraction plugins)
- **Geppetto** (LLM middleware for embeddings)

## Status

Currently in **design phase**. ASCII mockups and model sketches complete. No code written yet.
