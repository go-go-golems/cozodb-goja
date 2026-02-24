# Tasks

## Completed

- [x] Initialize ticket `COJS-01-INITIAL-BUILD` with design-doc and diary documents.
- [x] Import `/tmp/cozodb-js.md` with `docmgr import file --file ... --ticket ...`.
- [x] Collect code evidence from `go-go-goja`, `2026-02-18--cozodb-extraction`, and `goja`.
- [x] Pull and store upstream Cozo Node/WASM/sysops sources for adapter/policy validation.
- [x] Produce long-form 8+ page implementation research document.
- [x] Maintain detailed chronological investigation diary.
- [x] Add reproducible evidence script in ticket `scripts/` and capture output.

## Implementation Execution Plan

### Phase 1: Planning and ticket hygiene

- [ ] `P1-T1` Expand this task list into execution-level steps and commit docs checkpoint.
- [ ] `P1-T2` Add diary entry for execution kickoff with prompt context and commit references.
- [ ] `P1-T3` Record planning checkpoint in changelog.

### Phase 2: Core API domain layer (`pkg/cozoapi`)

- [ ] `P2-T1` Create core public types (`CozoValue`, `CozoResult`, `QueryOptions`, `PreparedQuery`, relation row payloads).
- [ ] `P2-T2` Implement result helpers (`Objects`, `FirstObject`, `Scalar`) with tests.
- [ ] `P2-T3` Implement query options normalization/clamping logic with tests.
- [ ] `P2-T4` Implement template compiler for `q`/`cq` interpolation into `$params` with tests.
- [ ] `P2-T5` Implement relation helper compiler (`create`, `put`, `insert`, `update`, `rm`, `del`, `get`, `columns`, `indices`, `access`) with tests.
- [ ] `P2-T6` Implement `atomic` compiler for chained transaction scripts with parameter namespacing and tests.
- [ ] `P2-T7` Implement policy engine for timeout/system-op/relation restrictions with tests.

### Phase 3: Backend abstraction and first adapter wiring

- [ ] `P3-T1` Define backend interface and capabilities contract.
- [ ] `P3-T2` Implement a test backend/fake backend for deterministic unit/integration validation.
- [ ] `P3-T3` Add optional Cozo CGO adapter scaffold behind build tags (`cozo_cgo`) to keep default build green.
- [ ] `P3-T4` Add adapter-level tests for fake backend and compile-only coverage for optional adapter.

### Phase 4: goja native module (`require("cozodb")`)

- [ ] `P4-T1` Implement native module adapter satisfying `modules.NativeModule`.
- [ ] `P4-T2` Implement `open()` returning db handle with `exec`, `q`, `cq`, `atomic`, `rel`, `export`, `import`, `close`.
- [ ] `P4-T3` Implement `db.rel(name)` object methods compiled through core relation compiler.
- [ ] `P4-T4` Ensure asynchronous API behavior is promise-based and runtime-owner safe.
- [ ] `P4-T5` Add integration tests using go-go-goja runtime + `require("cozodb")`.

### Phase 5: CLI and developer workflow

- [ ] `P5-T1` Replace scaffold command with practical CLI entrypoint(s) for script execution and/or REPL bootstrap.
- [ ] `P5-T2` Add examples in README for `require("cozodb")`, `db.rel(...)`, and `atomic`.
- [ ] `P5-T3` Add smoke tests for CLI paths where reasonable.

### Phase 6: Validation, commits, and documentation closure

- [ ] `P6-T1` Run `go test ./...` and fix failures.
- [ ] `P6-T2` Run `GOWORK=off go test ./...` and fix portability issues.
- [ ] `P6-T3` Run `make lint` (or note blockers) and address actionable findings.
- [ ] `P6-T4` Commit in focused slices with meaningful messages and reference diary steps.
- [ ] `P6-T5` Update diary with exact commands, failures, fixes, and commit hashes.
- [ ] `P6-T6` Update changelog and rerun `docmgr doctor --ticket COJS-01-INITIAL-BUILD --stale-after 30`.
