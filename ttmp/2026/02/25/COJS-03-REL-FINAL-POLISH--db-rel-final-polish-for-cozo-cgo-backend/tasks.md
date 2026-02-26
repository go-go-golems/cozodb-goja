# Tasks

## Phase 1: Ticket Planning and Scope Lock

- [ ] Confirm scope: only `cozo_cgo` + `fake` backend assumptions
- [ ] Define explicit API deltas for `db.rel()` input decoding and error payloads
- [ ] Record implementation order and validation commands in design doc
- [ ] Add diary kickoff entry with prompt context and planned workstreams

## Phase 2: `db.rel().create()` Input Shape Hardening (Item 1)

- [ ] Implement explicit decoder for relation create spec that accepts:
  - [ ] lowercase keys (`keys`, `values`) as canonical JS shape
  - [ ] uppercase compatibility aliases (`Keys`, `Values`) for existing scripts
- [ ] Implement explicit decoder for relation create options that accepts:
  - [ ] `replace`
  - [ ] `Replace` alias
- [ ] Ensure decoder errors are actionable and mention expected fields
- [ ] Add unit tests for lowercase, uppercase alias, and malformed inputs

## Phase 3: Explicit Header-Mapped Row API for Mutations (Item 2)

- [ ] Add decode path for tuple payload object:
  - [ ] `{ headers: string[], rows: any[][] }`
  - [ ] aliases `{ Headers, Rows }`
- [ ] Validate tuple payload:
  - [ ] non-empty headers
  - [ ] unique non-empty header names
  - [ ] every row length equals header length
- [ ] Map tuple rows to object rows using provided headers
- [ ] Reject ambiguous raw tuple arrays (`[[...], [...]]`) with a clear guidance error
- [ ] Preserve object-row payload support (`[{id:..., ...}]`)
- [ ] Add tests for valid and invalid header mapping payloads

## Phase 4: `db.rel()` Error Contract Cleanup (Item 5)

- [ ] Add standardized JS error object builder for promise rejections
- [ ] Include at minimum:
  - [ ] `message`
  - [ ] stable `code`
  - [ ] `operation` context for relation methods
- [ ] Route all `db.rel()` method promise errors through standardized envelope
- [ ] Add tests asserting error object fields for representative failures

## Phase 5: Real `cozo_cgo` Relation Integration Coverage (Item 4)

- [ ] Add build-tagged integration test file (`//go:build cozo_cgo`)
- [ ] Exercise relation lifecycle on real backend:
  - [ ] `create`
  - [ ] `put`
  - [ ] `insert`
  - [ ] `update`
  - [ ] `rm`
  - [ ] `del`
  - [ ] `get`
  - [ ] `columns`
  - [ ] `indices`
  - [ ] `access`
- [ ] Validate return shaping (`objects`, row values) and expected errors where applicable
- [ ] Keep test isolated via temp sqlite path and deterministic setup/teardown

## Phase 6: Validation, Diary, and Commit Hygiene

- [ ] Run focused tests:
  - [ ] `go test ./pkg/cozoapi/module -count=1`
  - [ ] `go test ./pkg/cozoapi -count=1`
- [ ] Run tagged real-backend tests:
  - [ ] `go test -tags cozo_cgo ./pkg/cozoapi/module -count=1`
- [ ] Update diary per implementation step with exact commands/results
- [ ] Update changelog with commit references and outcome summary
- [ ] Run `docmgr doctor --ticket COJS-03-REL-FINAL-POLISH --stale-after 30`
