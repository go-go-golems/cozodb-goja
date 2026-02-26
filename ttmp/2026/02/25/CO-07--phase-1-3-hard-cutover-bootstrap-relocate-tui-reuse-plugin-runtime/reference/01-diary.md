---
Title: Diary
Ticket: CO-07
Status: active
Topics:
    - cozodb
    - go
    - goja
    - tui
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/cmd/cozo-plugin-run/main.go
      Note: Workstream E smoke-runner command
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/cmd/cozo-tui/main.go
      Note: Relocated TUI entrypoint committed in CO-07 Workstream B
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/go.mod
      Note: Module/dependency baseline for relocation
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/go.sum
      Note: Resolved dependency lockfile for relocation
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/geppettohost/embedder.go
      Note: Runtime global embed helper
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/geppettohost/errors.go
      Note: Timeout/panic error normalization
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/geppettohost/host.go
      Note: High-level RunExtractorScript entrypoint
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/geppettohost/host_test.go
      Note: End-to-end extractor run integration smoke test
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/geppettohost/options.go
      Note: Host runtime options and defaults
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/geppettohost/runtime.go
      Note: goja runtime setup and module registration
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/geppettohost/runtime_test.go
      Note: Host runtime module registration sanity test
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/plugins/loader.go
      Note: Descriptor validation
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/plugins/loader_test.go
      Note: Loader validation tests
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/plugins/runner.go
      Note: Descriptor load/create/run flow with guarded executor
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/plugins/types.go
      Note: Plugin descriptor and run request structs
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/app/model.go
      Note: Relocated screen router for F1-F7
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/app/model_parity_test.go
      Note: Default backend app hotkey/quit parity checks
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/screens/dashboard/model_parity_test.go
      Note: Dashboard data fetch parity check
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/screens/evolution/model_parity_test.go
      Note: Evolution snapshot parity check
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/screens/network/model_parity_test.go
      Note: Network graph parity check
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/screens/people/model_parity_test.go
      Note: People list/preview parity check
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/screens/query/model_parity_test.go
      Note: Query execution parity check
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/screens/relationships/model_parity_test.go
      Note: Relationships detail parity check
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/screens/timeline/model_parity_test.go
      Note: Timeline parity check
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/seeddata/seed.go
      Note: Relocated seed data initializer
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/testutil/db.go
      Note: Seeded mem DB helper for cozo_cgo parity tests
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/scripts/fixtures/extractor_fixture.js
      Note: Deterministic smoke test extractor fixture
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/scripts/fixtures/transcript_fixture.txt
      Note: Fixture transcript used for smoke run
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/scripts/lib/relationship_constants.js
      Note: Relocated script support library
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/scripts/lib/relationship_extractor_factory.js
      Note: Relocated factory helper
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/scripts/lib/relationship_parsing.js
      Note: Relocated parsing helper
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/scripts/relation_extractor_reflective.js
      Note: Relocated reflective extractor script
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/scripts/relation_extractor_template.js
      Note: Relocated extractor template script
ExternalSources: []
Summary: Implementation diary for CO-07 phase 1-3 execution
LastUpdated: 2026-02-25T12:30:00-05:00
WhatFor: Chronological build/debug/validation notes for relocation and runtime foundation work
WhenToUse: Use to review implementation steps and reproduce results
---





# Diary

## Goal

Track CO-07 implementation task-by-task with exact commits, commands, failures, and validation guidance.

## Step 1: Bootstrap Extraction-Side Module (Workstream A)

This step established the new extraction-side module shell so that relocation could happen in a controlled way. The focus was creating a clean Go module boundary under `2026-02-18--cozodb-extraction` and proving the scaffold compiled before copying feature code.

The bootstrap commit intentionally stayed minimal and isolated to reduce rollback risk and simplify review.

### Prompt Context

**User prompt (verbatim):** "alright, implement it now, task by task, committing as you go, keeping a diary in each ticket."

**Assistant interpretation:** Begin execution immediately, complete tasks in checklist order, commit per logical milestone, and maintain diary documentation.

**Inferred user intent:** Get deterministic, auditable progress with granular commits and complete ticket hygiene.

**Commit (code):** `ebb3cb0` — "co-07: bootstrap cozo-extraction-tui module scaffold"

### What I did
- Created `2026-02-18--cozodb-extraction/cozo-extraction-tui`.
- Initialized module with `go mod init github.com/manuel/cozo-extraction-tui`.
- Added `cmd/cozo-tui/main.go` placeholder scaffold.
- Added module to workspace via `go work use ./2026-02-18--cozodb-extraction/cozo-extraction-tui` from workspace root.
- Ran initial build checks (`go list ./...`, `go test ./... -count=1`).

### Why
- Establishing module boundaries first prevents large relocation + dependency failures from mixing with scaffolding errors.

### What worked
- Module initialization and compilation succeeded.
- Workspace recognized the new module.

### What didn't work
- N/A for this step.

### What I learned
- The extraction workspace already had the right dependency graph to accept an additional module with minimal friction.

### What was tricky to build
- Ensuring this remained a small, reviewable commit while still satisfying module/workspace wiring requirements.

### What warrants a second pair of eyes
- Module path naming (`github.com/manuel/cozo-extraction-tui`) should be confirmed as final before external consumers depend on it.

### What should be done in the future
- Continue to keep commits small by isolating relocation/runtime/test phases.

### Code review instructions
- Start at `cozo-extraction-tui/go.mod` and `cozo-extraction-tui/cmd/cozo-tui/main.go`.
- Validate with:
  - `cd 2026-02-18--cozodb-extraction/cozo-extraction-tui`
  - `go list ./...`
  - `go test ./... -count=1`

### Technical details
- Workspace update command: `go work use ./2026-02-18--cozodb-extraction/cozo-extraction-tui`.
- Bootstrap scope intentionally excluded relocated TUI files.

## Step 2: Relocate F1-F7 TUI into New Module (Workstream B)

This step moved the existing TUI app/screen/seed packages into the extraction module and rewired imports so the relocated binary builds against `cozodb-goja` package APIs. The objective was functional parity at compile/smoke level without introducing behavior changes.

The relocation was completed as one cohesive commit after formatting and compile/test checks.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue task execution in order and commit each completed milestone with diary updates.

**Inferred user intent:** Preserve momentum while keeping an auditable task/commit trail.

**Commit (code):** `5089259` — "co-07: relocate F1-F7 TUI into extraction module"

### What I did
- Copied TUI sources from `cozodb-goja` into `cozo-extraction-tui`:
  - `cmd/cozo-tui/main.go`
  - `internal/tui/app/model.go`
  - `internal/tui/screens/*/model.go`
  - `internal/tui/seeddata/seed.go`
- Rewrote imports from `github.com/go-go-golems/cozodb-goja/internal/tui/...` to `github.com/manuel/cozo-extraction-tui/internal/tui/...`.
- Kept DB API imports on `github.com/go-go-golems/cozodb-goja/pkg/cozoapi` and `.../cozocgo`.
- Ran formatting and validation:
  - `gofmt -w ./cmd/cozo-tui/main.go ./internal/tui/app/model.go ./internal/tui/seeddata/seed.go ./internal/tui/screens/*/model.go`
  - `go list ./...`
  - `go test ./... -count=1`
  - `timeout 3s script -q -c 'go run ./cmd/cozo-tui --engine mem' /dev/null`
- Committed only `cozo-extraction-tui/**` and explicitly excluded unrelated untracked `.claude/` and `.idea/`.

### Why
- Relocation is required for hard cutover and reuse of extraction-side runtime/plugin stack.

### What worked
- Full relocated package tree compiled.
- Test run passed for all relocated packages.
- Pseudo-TTY smoke launch started the TUI process successfully (terminated by timeout as intended).

### What didn't work
- Initial path assumptions pointed to non-existent absolute paths and caused command errors:
  - `fatal: cannot change to '/home/manuel/workspaces/2026-02-24/2026-02-18--cozodb-extraction': No such file or directory`
  - Root cause: repos were nested under `/home/manuel/workspaces/2026-02-24/cozodb-goja-init/`.
- First `go mod tidy` attempt (earlier in this phase) had transient proxy parse error before succeeding on retry.

### What I learned
- The relocation can remain thin because existing TUI packages are self-contained and primarily depend on `cozoapi` public interfaces.
- Maintaining strict commit scope prevents accidental inclusion of local editor artifacts.

### What was tricky to build
- Running a Bubble Tea app in non-interactive automation requires a pseudo-TTY wrapper (`script`) and timeout guard. Without that, verification is ambiguous.

### What warrants a second pair of eyes
- Runtime behavior parity for interactive screens F1-F7 still needs manual walkthrough (Workstream C), not just compile-time parity.

### What should be done in the future
- Implement Workstream C manual parity checklist and capture screen-by-screen outcomes.

### Code review instructions
- Start in relocated entrypoint and app router:
  - `/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui/cmd/cozo-tui/main.go`
  - `/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/app/model.go`
- Then spot-check any screen package under `internal/tui/screens/` for import rewrites.
- Validate with:
  - `cd /home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui`
  - `go test ./... -count=1`
  - `timeout 3s script -q -c 'go run ./cmd/cozo-tui --engine mem' /dev/null`

### Technical details
- Import rewrite pattern:
  - from `github.com/go-go-golems/cozodb-goja/internal/tui/...`
  - to `github.com/manuel/cozo-extraction-tui/internal/tui/...`
- `go.sum` added by dependency resolution during relocation.

## Step 3: Build Plugin Runtime Foundation (Workstream D)

This step introduced the extraction-side plugin runtime primitives needed for JavaScript descriptor execution outside the old prototype runner binary. It adds two focused packages: `internal/plugins` for descriptor loading/normalization/decoding and `internal/geppettohost` for runtime wiring and guarded execution.

The result is a reusable host API that can execute extractor scripts with registered `cozodb` and geppetto modules, while enforcing timeout and panic recovery behavior.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue executing CO-07 checklist items with granular commits and diary updates.

**Inferred user intent:** Build the migration foundation incrementally and make each step reviewable and reproducible.

**Commit (code):** `213e79f` — "co-07: add plugin loader and geppetto host runtime foundation"

### What I did
- Added `internal/plugins` package:
  - `types.go` (descriptor/run types and constants)
  - `loader.go` (descriptor validation, run-input canonicalization, return decoding)
  - `runner.go` (descriptor require/create/run flow with guarded executor hook)
- Added `internal/geppettohost` package:
  - `options.go` (runtime options + defaults)
  - `runtime.go` (goja + eventloop + require registry + module registration)
  - `host.go` (high-level `RunExtractorScript` API)
  - `embedder.go` (global injection helper)
  - `errors.go` (timeout/panic normalization)
- Registered runtime modules:
  - `cozodb` via `github.com/go-go-golems/cozodb-goja/pkg/cozoapi/module`
  - `geppetto` and `geppetto/plugins` via `gp.Register(...)`
  - optional `runnerdb`/`database` via go-go-goja `DBModule`
- Added timeout interrupt guard (`vm.Interrupt`) and panic recovery wrapper in `ExecuteWithGuards`.
- Updated module dependencies with `go mod tidy`.
- Validated with `go test ./... -count=1`.

### Why
- CO-07 Phase 3 requires extracting runtime/plugin logic from prototype runner into reusable package boundaries in the relocated module.

### What worked
- All new packages compile.
- Runtime now exposes required module registrations.
- Guarded executor path integrates with plugin loader.

### What didn't work
- Initial build failed due `goja.InterruptedError` API misuse:
  - Error: `internal/geppettohost/errors.go:49:16: v.Value (value of type func() interface{}) is not an interface`
  - Fix: switched from `v.Value` to `v.Value()`.

### What I learned
- Interrupt payload extraction depends on goja version API shape (`Value()` accessor).
- Keeping timeout/panic handling in runtime layer allows plugin loader to stay transport-focused.

### What was tricky to build
- Timeout guard and panic normalization must cooperate with goja's panic-based interruption semantics; otherwise errors become ambiguous or interrupts leak across runs.

### What warrants a second pair of eyes
- Runtime lifecycle in `internal/geppettohost/runtime.go` (`loop.Start`/`Stop`, `DBModule` close, interrupt clearing) should be reviewed for edge-case cleanup ordering.

### What should be done in the future
- Add targeted tests for invalid descriptors, canonicalization, timeout handling, and panic normalization (Workstream F).

### Code review instructions
- Start with:
  - `/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/plugins/runner.go`
  - `/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/geppettohost/runtime.go`
- Then inspect:
  - `internal/plugins/loader.go`
  - `internal/geppettohost/errors.go`
- Validate with:
  - `cd /home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui`
  - `go test ./... -count=1`

### Technical details
- Guarded execution hook type: `type GuardedExecutor func(timeout time.Duration, fn func() (goja.Value, error)) (goja.Value, error)`.
- Timeout enforcement uses `vm.Interrupt(&TimeoutError{...})` and `vm.ClearInterrupt()` per run.

## Step 4: Relocate JS Assets and Add Smoke Runner (Workstream E)

This step moved extractor JS assets into the relocated module and added an executable smoke path that does not depend on external model credentials. The smoke command validates host/runtime wiring by executing a local fixture extractor end-to-end.

The fixture plugin emits deterministic structured JSON so output-shape checks are reliable in CI and local development.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue through remaining CO-07 tasks, commit each functional milestone, and preserve diary traceability.

**Inferred user intent:** Prove the new runtime path actually runs scripts, not only compiles.

**Commit (code):** `8696779` — "co-07: relocate extractor scripts and add plugin smoke runner"

### What I did
- Copied script set into relocated module:
  - `scripts/relation_extractor_template.js`
  - `scripts/relation_extractor_reflective.js`
  - `scripts/lib/relationship_constants.js`
  - `scripts/lib/relationship_parsing.js`
  - `scripts/lib/relationship_extractor_factory.js`
- Added local smoke fixtures:
  - `scripts/fixtures/extractor_fixture.js`
  - `scripts/fixtures/transcript_fixture.txt`
- Added `cmd/cozo-plugin-run/main.go` script runner CLI using `internal/geppettohost` + `internal/plugins`.
- Validated relative import paths in moved scripts:
  - `require("./lib/relationship_extractor_factory")` present in both top-level script variants
  - `scripts/lib/*.js` files exist at expected paths
- Executed smoke command:
  - `go run ./cmd/cozo-plugin-run --script ./scripts/fixtures/extractor_fixture.js --transcript ./scripts/fixtures/transcript_fixture.txt --script-root ./scripts`
- Verified decoded payload shape contains `metadata` + `extraction.people[]` + `extraction.relationships[]`.

### Why
- Workstream E requires script relocation and an executable runtime smoke gate to validate host/loader integration.

### What worked
- Fixture plugin executed successfully through runtime host.
- Output JSON decoded as expected and included descriptor metadata.

### What didn't work
- N/A in this step.

### What I learned
- A deterministic local fixture extractor is essential to validate runtime behavior without depending on network/model credentials.

### What was tricky to build
- The smoke command needed to set compatibility globals (`RELATIONSHIP_*`) so moved scripts and future runtime behavior stay aligned with prototype expectations.

### What warrants a second pair of eyes
- CLI contract for `cmd/cozo-plugin-run` should be reviewed before treating it as stable external interface.

### What should be done in the future
- Add this smoke command into CI or makefile targets once phase ticketing calls for automation.

### Code review instructions
- Start with:
  - `/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui/cmd/cozo-plugin-run/main.go`
  - `/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui/scripts/fixtures/extractor_fixture.js`
- Validate smoke run with the command listed above.

### Technical details
- Smoke output metadata confirmed plugin descriptor values:
  - `plugin_api_version = cozo.extractor/v1`
  - `plugin_kind = extractor`
  - `plugin_id = cozo.fixture.extractor`

## Step 5: Add Quality Gates and Fix Descriptor Panic (Workstream F)

This step implemented test coverage for loader validation and host integration, then fixed a panic uncovered by the new tests. The panic occurred when descriptor fields were missing and `String()` was called on undefined values.

After hardening field parsing, both module tests and boundary tests in `cozodb-goja/pkg/cozoapi` passed.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Complete the test/quality workstream and lock down boundary behavior.

**Inferred user intent:** Ensure migration quality is enforced through tests, not just manual checks.

**Commit (code):** `c4eba2f` — "co-07: add plugin/host runtime tests and harden descriptor parsing"

### What I did
- Added `internal/plugins/loader_test.go` covering:
  - missing descriptor id
  - wrong `apiVersion`
  - empty transcript canonicalization rejection
  - valid descriptor fixture execution path
- Added `internal/geppettohost/runtime_test.go` sanity test for module registration:
  - `require("cozodb")`
  - `require("geppetto")`
  - `require("geppetto/plugins")`
- Added `internal/geppettohost/host_test.go` integration smoke test running temporary JS fixture through `Host.RunExtractorScript`.
- Hardened descriptor parsing in `internal/plugins/loader.go` with safe optional field reads.
- Re-ran validation:
  - `cozo-extraction-tui`: `go test ./... -count=1`
  - `cozodb-goja`: `go test ./pkg/cozoapi/... -count=1`

### Why
- Workstream F explicitly requires loader/host tests and boundary stability checks.

### What worked
- Tests now cover key failure paths and basic runtime integration.
- Descriptor parsing is panic-safe for missing fields.
- Boundary package tests in `cozodb-goja` passed.

### What didn't work
- First run of new tests exposed a panic:
  - `panic: runtime error: invalid memory address or nil pointer dereference`
  - Location: `internal/plugins/loader.go` during `DecodeDescriptorMeta`
  - Cause: unsafe `descriptorObj.Get(...).String()` on missing fields
- First valid fixture run failed due missing explicit `apiVersion` in test fixture with shimmed `defineExtractorPlugin`.
  - Fix: set `apiVersion`/`kind` explicitly in loader fixture script.

### What I learned
- Descriptor decoding must treat missing values as first-class validation failures, not string coercions.
- Test shims for `geppetto/plugins` should either mimic defaults or fixture scripts must provide explicit defaults.

### What was tricky to build
- Building a realistic but deterministic loader integration test required a minimal native module shim for `geppetto/plugins` in goja require registry.

### What warrants a second pair of eyes
- Error-message contracts from `DecodeDescriptorMeta` and `LoadAndRunExtractorPlugin` should be reviewed for external caller expectations.

### What should be done in the future
- Expand timeout and panic-path tests to assert exact error typing (`ErrPluginRunTimeout`, `ErrPluginExecutionPanic`).

### Code review instructions
- Start with:
  - `/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/plugins/loader.go`
  - `/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/plugins/loader_test.go`
  - `/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/geppettohost/host_test.go`
- Validate with:
  - `cd /home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui && go test ./... -count=1`
  - `cd /home/manuel/workspaces/2026-02-24/cozodb-goja-init/cozodb-goja && go test ./pkg/cozoapi/... -count=1`

### Technical details
- Field-read helper added: `readOptionalStringField(obj, key)`.
- Loader fixtures use a local `geppetto/plugins` shim in tests to avoid heavy external runtime setup.

## Step 6: Add TUI Parity Test Harness and Record Linker Blocker (Workstream C, Partial)

This step added a parity test harness for the relocated TUI screens and app shell behavior. The shell-level checks (hotkeys and quit semantics) now run in default tests. Screen data parity checks were added behind `cozo_cgo` build tags because they require the native Cozo backend.

Execution of the `cozo_cgo` suite is currently blocked by a local static archive issue in a read-only module cache location.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue through remaining tasks and keep diary/changelog accurate, including blockers.

**Inferred user intent:** Maximize completed scope while keeping unresolved environment constraints explicit.

**Commit (code):** `1ee3a83` — "co-07: add cozo_cgo parity smoke tests for relocated tui"  
**Commit (code):** `007a1ae` — "co-07: make app parity tests run on default fake backend"

### What I did
- Added `cozo_cgo`-gated parity tests for screen fetch/load behavior:
  - `internal/tui/screens/*/model_parity_test.go`
- Added shared seeded-db helper for parity tests:
  - `internal/tui/testutil/db.go`
- Added app-level parity tests:
  - `internal/tui/app/model_parity_test.go`
- Refactored app parity tests to use `fakebackend` so default `go test ./...` validates:
  - F1-F7 hotkey routing
  - quit closes DB outside query screen
  - `q` on query screen does not trigger quit command
- Ran default suite:
  - `go test ./... -count=1` (pass)
- Attempted tagged parity suite:
  - `go test -tags cozo_cgo ./internal/tui/screens/... -count=1` (blocked)
- Attempted local fix:
  - `ranlib /home/manuel/go/pkg/mod/github.com/kraklabs/cie@v0.7.20/lib/libcozo_c.a` (permission denied)

### Why
- Workstream C requires parity evidence for relocated F1-F7 behavior.

### What worked
- Default test suite now includes app-level behavioral parity assertions.
- Cozo-backed screen parity tests are implemented and ready for environments with writable/valid native archive setup.

### What didn't work
- Tagged parity suite failed during link:
  - `/usr/bin/ld: .../libcozo_c.a: error adding symbols: archive has no index; run ranlib to add one`
- `ranlib` remediation failed due permissions:
  - `ranlib: could not create temporary file whilst writing archive: Permission denied`

### What I learned
- Native Cozo test execution depends on local archive hygiene and write access to module-cache artifacts.

### What was tricky to build
- Keeping parity tests valuable without destabilizing default test runs required split strategy: default backend for shell behavior, `cozo_cgo`-gated tests for full screen data parity.

### What warrants a second pair of eyes
- Decide whether to vendor or locally mirror `libcozo_c.a` to a writable path for deterministic `cozo_cgo` CI/testing.

### What should be done in the future
- Resolve native archive indexing path so `go test -tags cozo_cgo ./internal/tui/screens/...` can run and fully close Workstream C.

### Code review instructions
- Start with:
  - `/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/app/model_parity_test.go`
  - `/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/screens/dashboard/model_parity_test.go` (pattern for others)
- Validate:
  - default: `go test ./... -count=1`
  - native (currently blocked): `go test -tags cozo_cgo ./internal/tui/screens/... -count=1`

### Technical details
- Build tags used for native-dependent parity tests:
  - `//go:build cozo_cgo`
  - `// +build cozo_cgo`
