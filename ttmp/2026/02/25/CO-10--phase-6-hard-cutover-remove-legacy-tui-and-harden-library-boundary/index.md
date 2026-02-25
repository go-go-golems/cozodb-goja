---
Title: 'Phase 6 Hard Cutover: Remove Legacy TUI and Harden Library Boundary'
Ticket: CO-10
Status: active
Topics:
    - cozodb
    - go
    - goja
    - tui
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cozodb-goja/internal/tui
      Note: Legacy tree targeted for removal
    - Path: cozodb-goja/cmd/cozo-tui/main.go
      Note: Legacy command targeted for removal
ExternalSources: []
Summary: "Phase 6 ticket for hard cutover removal of legacy TUI paths and architecture boundary hardening"
LastUpdated: 2026-02-25T12:23:00-05:00
WhatFor: "Track final migration cleanup and irreversible cutover"
WhenToUse: "Use during CO-10 implementation and closeout"
---

# CO-10 -- Phase 6 Hard Cutover Cleanup

## Scope

Execute the irreversible hard cutover:

1. remove legacy TUI from `cozodb-goja`,
2. enforce relocated extraction-side TUI as canonical app path,
3. lock library boundary.

## Core Documents

- [Implementation Plan](design/01-implementation-plan-phase-6-hard-cutover-and-boundary-cleanup.md)
- [Tasks](tasks.md)
- [Changelog](changelog.md)

## Status

Current status: **active**
