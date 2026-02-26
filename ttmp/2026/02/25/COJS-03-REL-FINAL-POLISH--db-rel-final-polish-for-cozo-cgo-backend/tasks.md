# Tasks

## Phase 1: Ticket Planning and Scope Lock

- [x] Confirm scope: only `cozo_cgo` + `fake` backend assumptions
- [x] Define explicit API deltas for `db.rel()` input decoding and error payloads
- [x] Record implementation order and validation commands in design doc
- [x] Add diary kickoff entry with prompt context and planned workstreams

## Phase 2: `db.rel().create()` Input Shape Hardening (Item 1)

- [x] Implement explicit decoder for relation create spec that accepts:
  - [x] lowercase keys (`keys`, `values`) as canonical JS shape
  - [x] uppercase compatibility aliases (`Keys`, `Values`) for existing scripts
- [x] Implement explicit decoder for relation create options that accepts:
  - [x] `replace`
  - [x] `Replace` alias
- [x] Ensure decoder errors are actionable and mention expected fields
- [x] Add unit tests for lowercase, uppercase alias, and malformed inputs

## Phase 3: Explicit Header-Mapped Row API for Mutations (Item 2)

- [x] Add decode path for tuple payload object:
  - [x] `{ headers: string[], rows: any[][] }`
  - [x] aliases `{ Headers, Rows }`
- [x] Validate tuple payload:
  - [x] non-empty headers
  - [x] unique non-empty header names
  - [x] every row length equals header length
- [x] Map tuple rows to object rows using provided headers
- [x] Reject ambiguous raw tuple arrays (`[[...], [...]]`) with a clear guidance error
- [x] Preserve object-row payload support (`[{id:..., ...}]`)
- [x] Add tests for valid and invalid header mapping payloads

## Phase 4: `db.rel()` Error Contract Cleanup (Item 5)

- [x] Add standardized JS error object builder for promise rejections
- [x] Include at minimum:
  - [x] `message`
  - [x] stable `code`
  - [x] `operation` context for relation methods
- [x] Route all `db.rel()` method promise errors through standardized envelope
- [x] Add tests asserting error object fields for representative failures

## Phase 5: Real `cozo_cgo` Relation Integration Coverage (Item 4)

- [x] Add build-tagged integration test file (`//go:build cozo_cgo && cozo_cgo_integration`)
- [x] Exercise relation lifecycle on real backend:
  - [x] `create`
  - [x] `put`
  - [x] `insert`
  - [x] `update`
  - [x] `rm`
  - [x] `del`
  - [x] `get`
  - [x] `columns`
  - [x] `indices`
  - [x] `access`
- [x] Validate return shaping (`objects`, row values) and expected errors where applicable
- [x] Keep test isolated via temp sqlite path and deterministic setup/teardown
- [x] Add ticket script to run opt-in tagged integration test

## Phase 6: Validation, Diary, and Commit Hygiene

- [x] Run focused tests:
  - [x] `go test ./pkg/cozoapi/module -count=1`
  - [x] `go test ./pkg/cozoapi -count=1`
- [x] Run tagged real-backend tests:
  - [x] attempted `go test -tags cozo_cgo ./pkg/cozoapi/module -count=1` (linker symbol failure captured in diary)
  - [x] added opt-in tagged test command/script using `cozo_cgo cozo_cgo_integration`
- [x] Update diary per implementation step with exact commands/results
- [x] Update changelog with commit references and outcome summary
- [x] Run `docmgr doctor --ticket COJS-03-REL-FINAL-POLISH --stale-after 30`
