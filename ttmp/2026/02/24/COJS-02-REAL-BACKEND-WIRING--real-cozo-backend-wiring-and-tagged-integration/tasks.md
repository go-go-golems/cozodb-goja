# Tasks

## Implementation Plan

### Phase 1: Ticket setup and scoping

- [x] `P1-T1` Create ticket workspace and diary document.
- [x] `P1-T2` Replace placeholder tasks with execution-level checklist.
- [x] `P1-T3` Record kickoff context in diary and changelog.

### Phase 2: Cozo CGO backend wiring

- [x] `P2-T1` Add `github.com/cozodb/cozo-lib-go` dependency for tagged backend implementation.
- [x] `P2-T2` Implement `pkg/cozoapi/cozocgo` adapter with `Open/Exec/Import/Export/Close`.
- [x] `P2-T3` Map Cozo `NamedRows` to local `cozoapi.CozoResult` and relation export payloads.
- [x] `P2-T4` Apply query option directives (`:limit/:offset/:timeout`) in backend execution path.
- [x] `P2-T5` Preserve default (non-tag) build via stub adapter behavior.

### Phase 3: Module open options and default backend path

- [x] `P3-T1` Extend JS `open()` options to include engine/path/options for `cozocgo`.
- [x] `P3-T2` Thread new options through module decode and `DefaultOpen`.
- [x] `P3-T3` Add tests for options decoding and default-open behavior.

### Phase 4: Validation and tagged checks

- [x] `P4-T1` Run `go test ./...` (workspace mode) and fix issues.
- [x] `P4-T2` Run `GOWORK=off go test ./...` and fix issues.
- [x] `P4-T3` Run `make lint` and fix actionable findings.
- [x] `P4-T4` Attempt `cozo_cgo` tagged test/build path and document blockers if native libs are missing. (`go run -tags cozo_cgo ./cmd/XXX ...` currently fails to link: `cannot find -lcozo_c`.)

### Phase 5: Documentation and commit hygiene

- [x] `P5-T1` Update COJS-02 diary with commands, failures, and fixes.
- [x] `P5-T2` Update COJS-02 changelog with implementation milestones.
- [x] `P5-T3` Relate key files to diary/design docs.
- [x] `P5-T4` Run `docmgr doctor --ticket COJS-02-REAL-BACKEND-WIRING --stale-after 30`.
- [ ] `P5-T5` Commit in focused slices as tasks complete.

## Notes

- Tagged backend target: `-tags cozo_cgo`.
- Expected runtime dependency for tagged path: `libcozo_c` and Cozo headers from `cozo-lib-go`.
