---
Title: 'Implementation Plan: Phase 6 Hard Cutover and Boundary Cleanup'
Ticket: CO-10
Status: active
Topics:
    - cozodb
    - go
    - goja
    - tui
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: 2026-02-18--cozodb-extraction/cozo-extraction-tui/cmd/cozo-tui/main.go
      Note: New canonical TUI command path after cutover
    - Path: cozodb-goja/cmd/cozo-tui/main.go
      Note: |-
        Legacy command path to remove in hard cutover
        legacy command path targeted for removal
    - Path: cozodb-goja/go.mod
      Note: |-
        Library dependency surface to verify after TUI removal
        post-cutover library dependency sanity
    - Path: cozodb-goja/internal/tui
      Note: Legacy TUI package tree to remove in hard cutover
    - Path: cozodb-goja/internal/tui/app/model.go
      Note: legacy internal path targeted for removal
    - Path: cozodb-goja/ttmp/2026/02/24/CO-06--relocate-tui-to-cozodb-extraction-workspace-and-keep-cozodb-goja-reusable/design/01-relocation-and-reuse-plan-tui-in-extraction-workspace-cozodb-goja-as-library.md
      Note: Governing relocation architecture and hard-cutover policy
    - Path: go.work
      Note: |-
        Workspace module list to verify post-cutover
        workspace path verification after cutover
ExternalSources: []
Summary: Detailed phase-6 hard-cutover plan to remove legacy TUI paths, enforce new app boundary, and keep cozodb-goja strictly reusable infrastructure
LastUpdated: 2026-02-25T12:23:00-05:00
WhatFor: Execution spec for irreversible cutover cleanup and boundary hardening
WhenToUse: Use when finalizing migration and locking architecture boundaries
---


# CO-10 Implementation Plan (Phase 6 / Hard Cutover Cleanup)

## 1. Objective

Execute a hard cutover with no backward compatibility layer:

1. remove legacy TUI code paths from `cozodb-goja`,
2. ensure relocated extraction-side TUI is the only supported app path,
3. harden module boundaries so `cozodb-goja` remains a reusable bindings package.

This ticket is the architecture finalization step. It prevents drift and duplicate maintenance.

---

## 2. Cutover Policy (Non-Negotiable)

1. No dual-run mode.
2. No wrapper commands forwarding old path to new path.
3. No compatibility shims preserving removed package import paths.
4. Documentation and scripts must point only to relocated app.

If these conditions are not met, cutover is incomplete.

---

## 3. Removal Scope

## 3.1 Code removals in `cozodb-goja`

Remove:

1. `cozodb-goja/cmd/cozo-tui/main.go`
2. entire `cozodb-goja/internal/tui/` tree:
   - `app/`
   - `screens/*`
   - `seeddata/`

Retain:

1. `pkg/cozoapi/*`
2. `pkg/cozoapi/module/*`
3. backend adapters and tests related to bindings.

## 3.2 Doc/runbook updates

Update references in:

1. repo READMEs mentioning old TUI path,
2. CO-05/CO-06 docs where final command references still point to old path,
3. scripts and playbooks that invoke `cozodb-goja/cmd/cozo-tui`.

## 3.3 Workspace checks

1. ensure `go.work` includes extraction-side module,
2. ensure no removed module/package paths are referenced by any active package.

---

## 4. Boundary Hardening Plan

## 4.1 Library boundary statement

Post-cutover, `cozodb-goja` owns only:

1. Cozo backend abstractions and adapters,
2. query/relation APIs,
3. JS module exposure for `require("cozodb")`.

It does not own end-user product UI.

## 4.2 Boundary checks to enforce

1. No `internal/tui` directory in `cozodb-goja`.
2. No `cmd/cozo-tui` in `cozodb-goja`.
3. No references to removed paths in codebase (`rg` guard).
4. No future PR reintroducing TUI app code into `cozodb-goja` without explicit architecture approval.

## 4.3 CI/static guard strategy

Add simple guards:

1. grep-based CI step failing on `cozodb-goja/internal/tui` references,
2. grep-based CI step failing on `cozodb-goja/cmd/cozo-tui` references,
3. compile/test matrix that includes extraction-side module.

---

## 5. Detailed Execution Sequence

## Step 1: Pre-cutover sanity

1. Confirm CO-07, CO-08, CO-09 functionality exists in relocated app.
2. Run relocated TUI smoke (`F1-F9`).
3. Freeze legacy path (no new changes).

## Step 2: Remove legacy code

1. Delete legacy command file.
2. Delete legacy internal TUI directory.
3. Run `gofmt`/cleanup if needed for touched files.

## Step 3: Update references

1. Update README/runbook commands.
2. Update ticket docs where old command path appears.
3. Update helper scripts to relocated path.

## Step 4: Validate compile and tests

1. `go test ./...` in `cozodb-goja`.
2. `go test ./...` in relocated extraction-side module.
3. run relocation app smoke command.

## Step 5: Boundary audit

1. `rg` for old path references in workspace.
2. ensure zero hits for removed code imports.
3. document audit results in CO-10 changelog.

## Step 6: Finalize cutover state

1. set CO-10 tasks complete,
2. update CO-06 status notes,
3. mark legacy path as removed in architecture docs.

---

## 6. Verification Matrix

## 6.1 Functional verification

1. Relocated app boots and navigates all supported screens.
2. F8 extraction and F9 search remain operational post-removal.

## 6.2 Structural verification

1. No `cozodb-goja/internal/tui` files remain.
2. No `cozodb-goja/cmd/cozo-tui/main.go` remains.
3. `cozodb-goja` module still builds and tests successfully.

## 6.3 Documentation verification

1. primary run commands point to relocated app.
2. no docs recommend old command path.
3. changelog clearly records hard cutover decision and effect.

---

## 7. Risk Register

## Risk 1: Hidden dependencies on removed paths

Mitigation:

1. workspace-wide `rg` search and compile checks before merge.

## Risk 2: Developer confusion from stale docs

Mitigation:

1. explicit documentation update pass in same cutover PR,
2. changelog callout with old/new command path examples.

## Risk 3: Reintroduction of app code into library module

Mitigation:

1. CI grep guards,
2. architecture rule in CO-06/CO-10 docs,
3. code-review checklist item for boundary violations.

---

## 8. Rollback Policy

This is a hard cutover ticket; rollback should be avoided.

If emergency rollback is required:

1. revert full cutover commit as a single unit,
2. do not partially reintroduce selected legacy files,
3. open follow-up ticket with root-cause and corrected cutover plan.

---

## 9. Definition of Done (CO-10)

CO-10 is complete only when all are true:

1. Legacy TUI code is physically removed from `cozodb-goja`.
2. Relocated extraction-side app is the only supported TUI path.
3. Workspace compiles/tests pass across both modules.
4. Docs/scripts reference only relocated path.
5. Boundary guards are documented and/or automated.

This closes the relocation program and locks the long-term architecture.
