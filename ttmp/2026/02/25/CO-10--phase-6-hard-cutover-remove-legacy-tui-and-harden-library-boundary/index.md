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
    - Path: 2026-02-18--cozodb-extraction/cozo-extraction-tui/cmd/cozo-tui/main.go
      Note: Canonical relocated TUI command path
    - Path: cozodb-goja/.github/workflows/push.yml
      Note: Runs legacy-path guard in CI before tests
    - Path: cozodb-goja/Makefile
      Note: Adds guard-no-legacy-tui-paths target for repeatable local/CI checks
    - Path: cozodb-goja/ttmp/2026/02/25/CO-10--phase-6-hard-cutover-remove-legacy-tui-and-harden-library-boundary/changelog.md
      Note: Records hard-cutover implementation and validation evidence
    - Path: cozodb-goja/ttmp/2026/02/25/CO-10--phase-6-hard-cutover-remove-legacy-tui-and-harden-library-boundary/scripts/01-guard-no-legacy-tui-paths.sh
      Note: Guard script to block legacy TUI path regression
    - Path: cozodb-goja/ttmp/2026/02/25/CO-10--phase-6-hard-cutover-remove-legacy-tui-and-harden-library-boundary/tasks.md
      Note: Marks CO-10 closure checklist complete
ExternalSources: []
Summary: Phase 6 ticket for hard cutover removal of legacy TUI paths and architecture boundary hardening
LastUpdated: 2026-02-25T18:35:00-05:00
WhatFor: Track final migration cleanup and irreversible cutover
WhenToUse: Use during CO-10 implementation and closeout
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

Current status: **completed**
