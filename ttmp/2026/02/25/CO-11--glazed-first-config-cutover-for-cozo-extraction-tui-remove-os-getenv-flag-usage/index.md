---
Title: Glazed-first config cutover for cozo-extraction-tui (remove os.Getenv/flag usage)
Ticket: CO-11
Status: active
Topics:
    - cozo
    - tui
    - glazed
    - geppetto
    - embeddings
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui/cmd/cozo-tui/main.go
      Note: Current entrypoint requiring Glazed migration
    - Path: /home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui/cmd/cozo-plugin-run/main.go
      Note: Current plugin-run entrypoint requiring Glazed migration
    - Path: /home/manuel/workspaces/2026-02-24/cozodb-goja-init/geppetto/pkg/sections/sections.go
      Note: Reference implementation for bootstrap and middleware precedence
ExternalSources: []
Summary: Research and implementation planning ticket for migrating cozo-extraction-tui configuration to Glazed sections/middlewares with typed settings decode and no runtime os.Getenv calls.
LastUpdated: 2026-02-25T17:16:00-05:00
WhatFor: Define and track a hard cutover from flag/env-driven config to Glazed-first command and runtime wiring.
WhenToUse: Use when implementing command/config refactors across cozo-tui, plugin-run, geppettohost, seed, and F9 vector search.
---

# Glazed-first config cutover for cozo-extraction-tui (remove os.Getenv/flag usage)

## Overview

CO-11 defines a hard cutover plan for replacing ad-hoc `flag` and `os.Getenv` usage in `cozo-extraction-tui` with Glazed command sections, middleware-based source loading, and typed settings structs decoded from parsed fields.

## Current status

1. Deep analysis completed and documented.
2. Hard-cutover implementation completed in two code commits (`96cc0b9`, `7fe59cf`) in `2026-02-18--cozodb-extraction`.
3. Validation completed:
   - `go test ./... -count=1`
   - `make test-cgo-vsearch`
   - real seed-only smoke with parsed-values profile/config path
   - manual interactive PTY TUI smoke.
4. Precedence locked and validated for this pass: `flags > env > profiles > config > defaults`.
5. Hard cutover policy enforced: no backwards compatibility path retained.

## Key links

1. Design doc: [design-doc/01-glazed-first-configuration-architecture-and-implementation-plan.md](./design-doc/01-glazed-first-configuration-architecture-and-implementation-plan.md)
2. Diary: [reference/01-investigation-diary-glazed-cutover.md](./reference/01-investigation-diary-glazed-cutover.md)
3. Task tracker: [tasks.md](./tasks.md)
4. Changelog: [changelog.md](./changelog.md)

## Structure

1. `design-doc/` contains the long-form architecture and implementation plan.
2. `reference/` contains the chronological investigation diary and command evidence.
3. `scripts/` reserved for migration helpers and verification scripts during implementation.
