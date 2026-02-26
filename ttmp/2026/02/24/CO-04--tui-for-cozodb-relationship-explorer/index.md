---
Title: TUI for CozoDB Relationship Explorer
Ticket: CO-04
Status: complete
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
Summary: Terminal UI for exploring the CozoDB entity extraction database
LastUpdated: 2026-02-25T22:10:45.422782451-05:00
WhatFor: Interactively browse people, relationships, events, and behaviors stored in CozoDB
WhenToUse: After extracting entities with the JS runner, use this TUI to explore and query the results
---


# TUI for CozoDB Relationship Explorer

## Overview

A terminal-based interface (built with Bubbletea) for exploring the CozoDB entity extraction database. 9 screens cover browsing, querying, visualization, extraction, and semantic search.

## Screens

| # | Screen | Key | Status | Purpose |
|---|--------|-----|--------|---------|
| 1 | Dashboard | F1 | Done | Entity counts, top relationships, recent events, sentiment |
| 2 | People Browser | F2 | Done | Filterable/sortable table of all persons with preview |
| 3 | Relationship Explorer | F3 | Done | All relationship records with temporal snapshots |
| 4 | Relationship Evolution | F4 | Done | Drill-down: ASCII strength chart for a single pair |
| 5 | Network Graph | F5 | Done | ASCII circular graph with focus/depth controls |
| 6 | Timeline | F6 | Done | Chronological feed interleaving events, rels, behaviors |
| 7 | Query Console | F7 | Done | CozoScript REPL with results table and history |
| 8 | Extraction Monitor | F8 | CO-05 | Run JS extractors, preview + import results |
| 9 | Vector Search | F9 | CO-05 | HNSW semantic search across all entity types |

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

**7 of 9 screens implemented and working.** Screens F1-F7 are complete with sample data seeding. Remaining screens (F8 Extraction Monitor, F9 Vector Search) require infrastructure tracked in [CO-05](../CO-05--extraction-pipeline-and-vector-search-for-tui/index.md).
