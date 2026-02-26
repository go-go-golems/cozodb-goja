---
Title: Relocate TUI to CozoDB Extraction Workspace and Keep Cozodb-Goja Reusable
Ticket: CO-06
Status: complete
Topics:
    - cozodb
    - go
    - goja
    - tui
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Relocation plan for moving the Bubble Tea TUI into the extraction workspace while preserving cozodb-goja as reusable Cozo/Goja bindings
LastUpdated: 2026-02-25T22:10:45.657655985-05:00
WhatFor: Plan and track relocation of the Cozo TUI into the extraction workspace while keeping cozodb-goja as reusable infrastructure
WhenToUse: Use when implementing CO-06 migration phases or reviewing architecture/reuse decisions
---


# Relocate TUI to CozoDB Extraction Workspace and Keep Cozodb-Goja Reusable

## Overview

This ticket defines and tracks the relocation of the Bubble Tea TUI from `cozodb-goja` into `2026-02-18--cozodb-extraction`, with a strict boundary that keeps `cozodb-goja` as a reusable bindings/runtime module.

The design outcome is:

1. extraction domain app code (TUI + plugin/embedding/vector workflows) lives with extraction assets,
2. reusable Cozo bindings and JS module functionality remain in `cozodb-goja`,
3. CO-05 feature completion (F8/F9) proceeds in the relocated app path.

## Key Links

- [Design: Relocation and Reuse Plan](design/01-relocation-and-reuse-plan-tui-in-extraction-workspace-cozodb-goja-as-library.md)
- [Reference: Implementation Diary](reference/01-implementation-diary.md)
- [Tasks](tasks.md)
- [Changelog](changelog.md)
- [Preflight Script](scripts/01-relocation-preflight.sh)

## Status

Current status: **active**

## Topics

- cozodb
- go
- goja
- tui

## Tasks

See [tasks.md](tasks.md) for detailed phase tasks and current checkmarks.

## Changelog

See [changelog.md](changelog.md) for research and publication updates.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
