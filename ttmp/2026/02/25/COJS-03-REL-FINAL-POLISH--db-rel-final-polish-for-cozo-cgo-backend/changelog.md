# Changelog

## 2026-02-26

- Created ticket `COJS-03-REL-FINAL-POLISH` for final `db.rel()` polish.
- Added detailed implementation plan with explicit scope for items 1/2/4/5:
  - relation create decode hardening,
  - explicit header-mapped tuple row API,
  - relation error contract standardization,
  - tagged real `cozo_cgo` integration coverage.
- Expanded scaffold `tasks.md` into granular phased checklist.
- Added diary kickoff entry documenting prompt context, scope lock, and execution order.
- Implemented item 1 + 2 hardening in `pkg/cozoapi/module/cozodb.go`:
  - explicit relation create decoding (`keys`/`values` + `Keys`/`Values` aliases, `replace` + `Replace`);
  - explicit tuple payload mapping object `{headers, rows}` with validation;
  - rejection of ambiguous raw tuple arrays with guidance error.
- Added/updated tests for create/mutation tuple mapping:
  - `TestModuleRelCreateSupportsLowercaseAndUppercaseSpecFields`
  - `TestModuleRelMutationTupleRowsRequireExplicitHeaders`
  - `TestModuleRelMutationTuplePayloadWithHeaders`
  - commit: `080be31`.
- Implemented relation-specific error envelope for promise rejections:
  - stable payload fields `message`, `code`, and `operation` for `db.rel()` methods;
  - added test `TestModuleRelErrorPayloadIncludesCodeAndOperation`;
  - commit: `365bd17`.
- Added opt-in real backend integration coverage:
  - `pkg/cozoapi/module/cozodb_rel_cozocgo_integration_test.go`
  - tag gate: `cozo_cgo && cozo_cgo_integration`
  - lifecycle coverage: `create/put/insert/update/rm/del/get/columns/indices/access`
  - commit: `d02c1b0`.
- Added ticket script for integration run:
  - `scripts/01-run-rel-cozocgo-integration.sh`
  - command: `GOWORK=off go test -tags 'cozo_cgo cozo_cgo_integration' ./pkg/cozoapi/module -count=1 -run TestRelLifecycleWithCozoCGOBackend`.
- Captured environment linker limitation:
  - plain `-tags cozo_cgo` run failed in this environment with unresolved `cozo_*` symbols;
  - integration path intentionally kept opt-in to avoid default pipeline breakage.

## 2026-02-25

- Initial workspace created
