---
Title: 'Phase 1-3 Hard Cutover: Bootstrap, Relocate TUI, Reuse Plugin Runtime'
Ticket: CO-07
Status: complete
Topics:
    - cozodb
    - go
    - goja
    - tui
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go.work
      Note: Workspace module list that will include relocation module
    - Path: cozodb-goja/internal/tui/app/model.go
      Note: Source app router for relocation
    - Path: 2026-02-18--cozodb-extraction/cozo-relationship-js-runner/plugin_loader.go
      Note: Source plugin loader logic for adaptation
ExternalSources: []
Summary: Execution ticket for phases 1-3 foundation with hard cutover posture
LastUpdated: 2026-02-25T22:10:45.772779044-05:00
WhatFor: Track implementation of module bootstrap, TUI relocation, and plugin runtime foundation
WhenToUse: Use during CO-07 implementation and review
---


# CO-07 -- Phases 1-3 Hard Cutover Foundation

## Scope

This ticket implements the relocation foundation:

1. bootstrap extraction-side TUI module,
2. relocate F1-F7 TUI,
3. establish plugin runtime foundation.

## Core Documents

- [Implementation Plan](design/01-implementation-plan-phases-1-3-hard-cutover-foundation.md)
- [Tasks](tasks.md)
- [Changelog](changelog.md)

## Status

Current status: **active**
