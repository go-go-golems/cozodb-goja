---
Title: implementation plan db.rel final polish
Ticket: COJS-03-REL-FINAL-POLISH
Status: active
Topics:
    - api
    - cozodb
    - goja
    - javascript
    - cgo
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-25T22:19:59.604777638-05:00
WhatFor: ""
WhenToUse: ""
---

# implementation plan db.rel final polish

## Executive Summary

This ticket performs final `db.rel()` API polish for the `cozo_cgo`-only product direction (plus `fake` backend for deterministic tests). The work closes four high-impact gaps:

1. explicit JS input shape decoding for `rel.create` and options,
2. explicit header-mapped tuple API for relation mutation rows,
3. standardized structured error payloads for relation method promise rejections,
4. build-tagged real-backend integration coverage for the full relation lifecycle.

The implementation intentionally favors deterministic behavior over implicit coercions or backwards-compatibility shims.

## Problem Statement

Current relation method behavior has three practical issues:

1. **Input decoding ambiguity**: struct export behavior hides whether lowercase JS field names (`keys`) or Go-struct-shaped names (`Keys`) are required.
2. **Tuple row ambiguity**: tuple arrays are currently converted to synthetic keys (`c0`, `c1`), which silently diverges from real relation schemas.
3. **Error payload inconsistency**: rejected promises expose only `message`, making robust JS-side error handling brittle.

Additionally, there is insufficient real `cozo_cgo` integration coverage of `db.rel()` behavior.

## Proposed Solution

### A. Hard decode layer for `rel.create`

- Introduce explicit decoders for:
  - relation spec (`keys`/`values`, plus `Keys`/`Values` aliases),
  - create options (`replace`, plus `Replace` alias).
- Return precise validation errors (missing keys, wrong types, empty names).

### B. Explicit tuple payload contract for relation mutations

- Support tuple input only via explicit object payload:
  - `{ headers: string[], rows: any[][] }` (and `Headers`/`Rows` aliases).
- Validate header uniqueness and row length consistency.
- Reject bare tuple arrays (`[[...]]`) with a guidance error to use `{headers, rows}`.

### C. Standardized `db.rel()` promise rejection payload

- Add structured error object fields:
  - `message` (human-readable),
  - `code` (stable machine-friendly identifier),
  - `operation` (relation method context such as `rel.put`).
- Route relation methods through this error envelope.

### D. Tagged real backend integration tests

- Add `//go:build cozo_cgo` integration tests that run full `db.rel()` lifecycle against sqlite-backed `cozo_cgo`.
- Cover CRUD-like operations and metadata ops (`columns`, `indices`, `access`).

## Design Decisions

1. **Explicit decoding over generic `ExportTo`**
Reason: avoids hidden behavior drift and makes JS contract clear.

2. **No raw tuple support without headers**
Reason: synthetic keys are ambiguous and unsafe for production relation schemas.

3. **Alias support for uppercase keys retained in decoder**
Reason: allows controlled migration from older scripts while canonicalizing lowercase JS shape.

4. **Build-tagged real backend tests**
Reason: preserves default CI portability while enabling high-value integration validation when `cozo_cgo` is enabled.

## Alternatives Considered

1. **Keep current implicit tuple behavior (`c0/c1/...`)**
Rejected because it silently produces invalid column mappings in real schemas.

2. **Auto-discover headers from relation metadata for tuple rows**
Rejected for this ticket due added runtime complexity, policy/system-op coupling, and potential ordering mismatches.

3. **Global error envelope migration in one sweep**
Deferred; this ticket scopes guaranteed envelope behavior to `db.rel()` methods first.

## Implementation Plan

1. Implement create-spec and create-options decoders in module layer.
2. Implement tuple payload decoder and row mapping helper.
3. Wire relation methods to use new decode paths and standardized error envelope.
4. Add/extend module unit tests for decoder and error behavior.
5. Add `cozo_cgo` tagged integration tests for relation lifecycle.
6. Validate with focused and tagged test runs.
7. Record diary/changelog evidence and finalize ticket hygiene.

## Open Questions

1. Should `db.rel().get()` return standardized `null` instead of `undefined` for missing rows?
2. Should tuple payload also support optional `strict` mode for rejecting extra object fields?

## References

- `pkg/cozoapi/module/cozodb.go` (current decode + promise behavior)
- `pkg/cozoapi/relation.go` (relation compiler behavior)
- `pkg/cozoapi/module/default_open.go` (`cozo_cgo` default opening path)
- `ttmp/.../COJS-01.../sources/local/01-cozodb-js.md` (original API proposal)
