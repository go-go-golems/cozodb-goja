---
Title: Implementation Diary
Ticket: CO-04
Status: active
Topics:
    - cozodb
    - tui
    - bubbletea
    - go
    - goja
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Chronological diary of designing and building the CozoDB TUI"
LastUpdated: 2026-02-24T20:00:00-05:00
WhatFor: "Track progress, decisions, and discoveries while building the TUI"
WhenToUse: "Consult when reviewing how the TUI was designed or why decisions were made"
---

# Implementation Diary -- CO-04 CozoDB TUI

## 2026-02-24 20:00 -- Design Phase

### What happened

- Brainstormed 9 TUI screens based on the CozoDB entity extraction data model
- Created ASCII mockups for each screen
- Sketched bubbletea Model/Msg/Cmd structure in YAML DSL for each screen
- Documented shared components (status bar, help overlay, spinner)
- Defined application-level model with screen routing

### Design decisions

**Screen selection**: Chose 9 screens to cover all major interaction patterns:
1. **Dashboard** — overview/landing, aggregate queries
2. **People Browser** — filterable table, preview pane
3. **Relationship Explorer** — temporal composite key browsing
4. **Relationship Evolution** — drill-down, ASCII chart, diff mode
5. **Network Graph** — ASCII graph rendering, community detection
6. **Timeline** — interleaved multi-entity chronological view
7. **Query Console** — CozoScript REPL with results table
8. **Extraction Monitor** — Goja plugin integration, import workflow
9. **Vector Search** — HNSW semantic search via Geppetto embeddings

**Navigation model**: F1-F8 for direct screen jumping + tab cycling. This avoids a nested menu system. Some screens (Evolution, Detail views) are drill-downs reached via `enter` from browser screens.

**YAML DSL for models**: Each screen's model is described as:
- `model:` — the bubbletea Model struct fields, referencing bubbles components where applicable
- `msgs:` — the message types that screen handles
- `cmds:` — the async commands (CozoDB queries, embeddings, etc.)
- `keybindings:` — vim-style key mappings

**CozoDB query integration**: All queries written as CozoScript embedded in the YAML. This maps directly to `db.Run()` calls in Go. The async pattern wraps each query in a `tea.Cmd` that returns a `tea.Msg` with results.

**Network graph**: Most ambitious screen. Force-directed ASCII layout is impractical for large graphs, so the design includes depth limiting and focus mode. Community detection via CozoDB's built-in `CommunityDetectionLouvain` fixed rule.

### What's next

- Review designs on reMarkable for markup/feedback
- Start implementation with the simplest screens first (Dashboard, People Browser)
- Implement the CozoDB connection wrapper and query helper
- Build shared components (status bar, help overlay)
- Tackle the harder screens (Network Graph, Query Console) last
