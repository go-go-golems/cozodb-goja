---
Title: CozoDB Goja JavaScript API Research and Implementation Blueprint
Ticket: COJS-01-INITIAL-BUILD
Status: active
Topics:
    - cozodb
    - goja
    - javascript
    - api
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: 2026-02-18--cozodb-extraction/cozo-relationship-js-runner/main.go
      Note: |-
        Real-world goja runner composition with module registration and host context injection
        Practical runner composition and native module registration example
    - Path: 2026-02-18--cozodb-extraction/cozo-relationship-js-runner/plugin_loader.go
      Note: |-
        Typed descriptor contract and defensive plugin execution pattern
        Descriptor contract validation pattern
    - Path: 2026-02-18--cozodb-extraction/ttmp/2026/02/18/COZO-01-INITIAL-ASSESSMENT--initial-assessment-python-cozo-extraction-to-go-geppetto-pinocchio/sources/cozo-lib-go-cozo.go
      Note: |-
        cgo wrapper surface for embedded Cozo in Go
        Cozo Go wrapper API evidence for backend adapter design
    - Path: cozodb-goja/cmd/XXX/main.go
      Note: Current repository entrypoint placeholder to replace with real CLI commands
    - Path: cozodb-goja/ttmp/2026/02/24/COJS-01-INITIAL-BUILD--initial-build-plan-for-cozodb-goja-js-api/sources/local/01-cozodb-js.md
      Note: |-
        Requested API surface and behavioral target
        Imported requested API contract
    - Path: go-go-goja/engine/factory.go
      Note: |-
        Runtime builder/factory lifecycle and ownership patterns
        Runtime factory composition and ownership lifecycle evidence
    - Path: go-go-goja/engine/module_specs.go
      Note: Explicit module registration model for require() enablement
    - Path: go-go-goja/modules/common.go
      Note: Native module interface and default registry conventions
    - Path: go-go-goja/pkg/repl/evaluators/javascript/evaluator.go
      Note: |-
        Existing reusable evaluator and REPL integration behavior
        Current evaluator wiring and module behavior references
    - Path: go-go-goja/pkg/runtimeowner/runner.go
      Note: |-
        Safe single-owner scheduling for goja runtime access
        Serialized owner scheduling patterns for safe goja access
    - Path: goja/README.md
      Note: |-
        Runtime constraints and embedding caveats
        Goja runtime safety constraints
    - Path: goja/builtin_promise.go
      Note: |-
        Promise scheduling constraints on VM owner loop
        Promise settlement constraints and event loop requirements
    - Path: goja/runtime.go
      Note: |-
        Interrupt and non-concurrent runtime semantics
        Interrupt and runtime execution model details
ExternalSources:
    - local:01-cozodb-js.md
    - local:01-cozo-lib-nodejs-readme.md
    - local:cozo-lib-nodejs-index.js
    - local:02-cozo-lib-wasm-readme.md
    - local:cozodb-sysops.html
    - https://docs.cozodb.org/en/latest/queries.html
    - https://docs.cozodb.org/en/latest/stored.html
    - https://docs.cozodb.org/en/latest/sysops.html
    - https://raw.githubusercontent.com/cozodb/cozo/main/cozo-lib-nodejs/README.md
    - https://raw.githubusercontent.com/cozodb/cozo/main/cozo-lib-nodejs/index.js
    - https://raw.githubusercontent.com/cozodb/cozo/main/cozo-lib-wasm/README.md
Summary: Evidence-driven implementation blueprint for a sandbox-safe CozoDB JavaScript API on goja, with adapter design, runtime ownership model, policy enforcement, and phased build plan.
LastUpdated: 2026-02-24T17:41:00-05:00
WhatFor: Build the initial CozoDB JavaScript API in `cozodb-goja` with concrete interfaces, module boundaries, and test strategy.
WhenToUse: Use this document when implementing the first production-capable API layer, reviewing architecture decisions, or onboarding contributors.
---


# CozoDB Goja JavaScript API Research and Implementation Blueprint

## Executive Summary

This document provides an implementation-ready design for `COJS-01-INITIAL-BUILD`: creating a Go-hosted JavaScript API for CozoDB, exposed inside goja with a modern async developer experience and strong policy control.

The requested API from `01-cozodb-js.md` targets five outcomes:

1. One async API shape for all backends (`node`, `wasm`, future `remote`).
2. First-class CozoScript execution (`exec`, `q`, `cq`, `atomic`).
3. Relation-focused ergonomic operations compiled into documented Cozo mutation/query forms.
4. Sandbox-safe capability model via a host-owned request port.
5. Host policy controls (timeouts, relation restrictions, and system-op restrictions).

The codebase already contains most foundational runtime primitives:

1. Explicit runtime composition and lifecycle (`engine.NewBuilder()`, immutable factory, owned runtime close) in `go-go-goja/engine/factory.go`.
2. Safe single-owner runtime scheduling in `pkg/runtimeowner`.
3. Existing JS evaluator and REPL adapters that can host a custom module.
4. Real runner patterns from `cozo-relationship-js-runner` for event loop ownership, module registration, and typed plugin contracts.
5. Prior Cozo assessment evidence including a working `cozo-lib-go` CGO wrapper surface.

The main work is not inventing a runtime model; it is composing existing pieces into a dedicated `cozodb` native module and an API-layer codec that:

1. normalizes values and errors,
2. compiles query templates and relation helpers into CozoScript,
3. enforces host policies before execution,
4. supports backend-specific capabilities while preserving one JS-facing contract.

This blueprint includes:

1. evidence-backed current-state map,
2. gap analysis against requested API,
3. concrete package layout and API sketches,
4. runtime and policy pseudocode,
5. phased implementation and test strategy,
6. risks, alternatives, and open decisions.

## Problem Statement and Scope

### Requested outcome

The imported specification (`sources/local/01-cozodb-js.md`) defines a TypeScript API with:

1. `exec`, `q`, `cq`, `atomic`, `rel`, `export`, `import`, `close`.
2. `QueryOptions` covering `limit`, `offset`, `timeoutSec`, `immutable`.
3. relation CRUD helpers (`create`, `put`, `insert`, `update`, `rm`, `del`, `get`, `columns`, `indices`, `access`).
4. a transport-safe `port.request()` model for sandbox embedding.

### Why this is non-trivial

The repository `cozodb-goja` is currently scaffold-only (`cmd/XXX/main.go`) and has no Cozo module yet. The work requires bridging four constraints:

1. goja runtime ownership is single-threaded and not goroutine-safe.
2. Cozo bindings differ by backend (Node async callback surface, WASM sync string surface, Go CGO surface).
3. Cozo mutation/query/system semantics must be preserved precisely.
4. Sandbox embedding requires capability boundaries stronger than “pass raw DB object into script”.

### Scope included

1. Embedded-Go architecture using `go-go-goja` patterns.
2. JS API surface and codecs matching requested model.
3. Policy enforcement and sandbox transport model.
4. REPL runner path for development and diagnostics.
5. Phased implementation and test plan.

### Scope excluded for this initial ticket

1. Full remote transport implementation (define interface now, implement later).
2. Production packaging of Cozo static libs for all OSes (design and gating included; full distribution automation deferred).
3. Performance tuning beyond baseline correctness + guardrails.

## Current-State Architecture (Evidence-Backed)

### 1) Runtime ownership and module composition foundation already exists

`go-go-goja/engine/factory.go` implements a builder/factory model with:

1. explicit module registration (`WithModules`) and runtime initializers,
2. immutable build with duplicate ID validation,
3. per-runtime creation that wires goja VM, goja_nodejs eventloop, require module, console, and runtime owner runner,
4. explicit cleanup (`Runtime.Close`).

Evidence:

1. builder composition and validation in `factory.go:60-132`.
2. runtime creation and owner runner wiring in `factory.go:147-177`.
3. shutdown path in `engine/runtime.go:31-49`.

This gives a stable host for a new `cozodb` module without changing core runtime architecture.

### 2) Native module registry is explicit and reusable

`go-go-goja/modules/common.go` provides `NativeModule` with `Name()`, `Doc()`, `Loader(vm, moduleObj)` and global registration via `modules.Register(...)` + `DefaultRegistryModules()` bridge in engine.

Evidence:

1. interface in `modules/common.go:30-34`.
2. registry enablement in `modules/common.go:77-83`.
3. explicit opt-in default registration in `engine/module_specs.go:78-82`.

This directly matches the needed injection model (`require("cozodb")`).

### 3) Runtimeowner runner solves safe cross-goroutine scheduling

`pkg/runtimeowner/runner.go` handles:

1. serialized loop scheduling (`RunOnLoop`),
2. `Call`/`Post` APIs with context, cancellation, and panic recovery,
3. owner-context fast path to avoid deadlock for nested owner calls.

Evidence:

1. runner constructor and invariants in `runner.go:34-45`.
2. call scheduling path in `runner.go:78-101`.
3. post scheduling path in `runner.go:129-148`.
4. panic handling in `runner.go:150-178`.

This is critical because goja itself is not goroutine-safe.

### 4) goja constraints are explicit and must drive design

goja documentation and source make runtime constraints non-negotiable:

1. runtime is not goroutine-safe (`README.md:99-103`),
2. promise resolution functions are not goroutine-safe and must be called on runtime loop (`builtin_promise.go:610-623`),
3. timeout/interruption model is host-driven (`runtime.go:1508-1525`),
4. several runtime setter methods are only safe from VM goroutine or when VM is idle (`runtime.go:2430-2432`).

Therefore, every async backend callback must settle JS promises through owner scheduling.

### 5) Existing runner (`cozo-relationship-js-runner`) demonstrates reusable pattern

`2026-02-18--cozodb-extraction/cozo-relationship-js-runner/main.go` shows real composition:

1. eventloop start/stop,
2. runtimeowner runner creation,
3. require registry with module root,
4. native module registration (`runnerdb`, `database`),
5. host context injection into VM,
6. plugin contract execution via typed descriptor loader.

Evidence:

1. runtime and registry wiring in `main.go:408-447`.
2. host context/injection in `main.go:452-510`.
3. descriptor validation and run in `plugin_loader.go:74-159`.

This pattern is directly portable for a Cozo capability-port implementation.

### 6) Prior Cozo Go integration evidence exists (`cozo-lib-go`)

From the extracted assessment sources:

1. `cozo-lib-go` uses CGO (`#include "cozo_c.h"`, link flags) and exposes `New`, `Run`, `ImportRelations`, `ExportRelations`, `Backup`, `Restore`, `ImportRelationsFromBackup`.
2. `Run` takes query + params map and returns named rows JSON-decoded shape.

Evidence:

1. CGO and link flags in `cozo-lib-go-cozo.go:18-25`.
2. constructor and run in `cozo-lib-go-cozo.go:60-140`.
3. import/export/backup/restore APIs in `cozo-lib-go-cozo.go:142-279`.

This is sufficient for a first embedded-Go adapter.

### 7) Requested API is aligned with real Cozo semantics

From local/remote Cozo sources:

1. query options `:limit`, `:offset`, `:timeout` are first-class (`cozodb-queries.html:477-510`),
2. timeout default is 300s, `0` disables (`cozodb-queries.html:506-510`),
3. script chaining `{...}{...}` forms single transaction boundary (`cozodb-stored.html:432-438`),
4. `:returning` behavior and `_kind` semantics for mutation ops (`cozodb-stored.html:346-351`),
5. system ops start with `::` and must appear alone (`cozodb-sysops.html:265-266`),
6. `::access_level` supports `normal`, `protected`, `read_only`, `hidden` and is guardrail-only (`cozodb-sysops.html:345-355`).

### 8) Upstream Node/WASM adapter differences are concrete

Node (`cozo-lib-nodejs-index.js` / README):

1. `run(script, params, immutable)` exists (`index.js:55-65`),
2. explicit `close()` required (`README.md:54-58`),
3. optional `multiTransact`, callbacks, named rules (`index.js:51-53`, `128-156`, README `117-119`).

WASM (`02-cozo-lib-wasm-readme.md`):

1. API is synchronous and string-based (`run(script, params: string): string`) (`README.md:48-55`),
2. blocks main thread unless moved to worker (`README.md:58-60`).

This validates the requested “one async interface with capability flags” approach.

## Gap Analysis Against Requested API

### Gap A: No `cozodb` module or adapters in this repository

Current `cozodb-goja` repository has no Cozo API code, module package, or tests. Only scaffold command exists.

Impact:

1. no JS-facing API,
2. no backend abstraction,
3. no policy controls.

### Gap B: API contract mismatch in existing evaluator module metadata

`Evaluator.GetAvailableModules()` reports `http`, `path`, `url` as standard modules (`evaluator.go:966-972`), but current local `go-go-goja/modules` directory only includes `database`, `exec`, `fs`, `glazehelp`.

Impact:

1. REPL help may imply unavailable modules,
2. this inconsistency should be corrected or documented while adding `cozodb`.

### Gap C: No Cozo-specific value codec and template compiler

Requested API depends on:

1. safe parameter mapping from template interpolations to `$param`.
2. conversion between JS values, Go values, and Cozo JSON-compatible values.
3. helper result wrappers (`objects()`, `firstObject()`, `scalar()`).

No such codec layer exists yet.

### Gap D: No policy enforcement boundary

Current patterns allow direct module calls. The requested design requires a capability/port boundary where host can enforce:

1. timeout clamping,
2. relation allow/deny lists,
3. system op restrictions.

No policy engine exists yet in repository.

### Gap E: Cozo CGO packaging risk is unaddressed

`cozo-lib-go` integration requires static library presence and linker flags (README evidence in prior ticket source). The current repo has no build tags, make targets, or diagnostics for missing native Cozo libs.

## Proposed Architecture

## High-level design

```text
JavaScript user code in goja
  -> require("cozodb")
      -> JS facade object (exec/q/cq/atomic/rel/export/import/close)
          -> Go module bridge (native module Loader)
              -> API service layer (compile + validate + policy + execute)
                  -> Backend adapter interface
                     -> CozoGoAdapter (cozo-lib-go / cgo)
                     -> CozoNodeAdapter (future)
                     -> CozoWasmAdapter (future worker/port)
```

### Package layout proposal (`cozodb-goja`)

1. `pkg/cozoapi/types.go`
   - core request/response/value/result/query option structs.
2. `pkg/cozoapi/result_helpers.go`
   - row/header helper methods (`Objects`, `FirstObject`, `Scalar`).
3. `pkg/cozoapi/template.go`
   - `q` / `cq` compilation and parameter binding.
4. `pkg/cozoapi/relation_compile.go`
   - compile relation helper methods to CozoScript + params.
5. `pkg/cozoapi/policy.go`
   - timeout clamp, system-op rules, relation-level restrictions.
6. `pkg/cozoapi/backend/interface.go`
   - adapter interface and capability model.
7. `pkg/cozoapi/backend/cozogo/adapter.go`
   - CGO-backed Cozo adapter implementation.
8. `pkg/cozoapi/module/cozodb.go`
   - go-go-goja `NativeModule` loader, exports, and promise-based async wrappers.
9. `pkg/cozoapi/port/port.go`
   - request/response operation model for capability-port transport.
10. `cmd/cozodb-js-repl/main.go`
   - dedicated REPL command with `cozodb` pre-wired.
11. `integration/cozoapi_module_test.go`
   - runtime integration tests with `engine.NewBuilder().WithModules(...)`.

### Core backend interface

```go
type Capabilities struct {
    Persistence   bool
    BackupRestore bool
    Callbacks     bool
    NamedRules    bool
    ImmutableRun  bool
}

type Backend interface {
    ID() string
    Capabilities() Capabilities
    Exec(ctx context.Context, script string, params map[string]any, opts QueryOptions) (CozoResult, error)
    Export(ctx context.Context, relations []string) (map[string]RelationRows, error)
    Import(ctx context.Context, data map[string]RelationRows) error
    Close(ctx context.Context) error
}
```

### JS API contract in module export

The module should expose `create(opts)` and return a db handle with methods matching imported spec.

Recommended runtime API shape for goja:

1. `const db = require("cozodb").open({ backend: "cozogo", ... })`
2. `await db.exec(...)`, `await db.q`...
3. `db.rel("users").put(...)`

Reason:

1. aligns with current require module model,
2. allows multiple db handles with different policies,
3. keeps host policy construction outside JS code if needed.

### Policy model

```go
type Policy struct {
    MaxTimeoutSec      int
    AllowSystemOps     bool
    AllowedRelations   map[string]struct{}
    DeniedRelations    map[string]struct{}
    MaxLimit           int
    MaxOffset          int
    DenyAccessLevelSet bool // optional stricter mode
}
```

Policy enforcement steps:

1. Parse/normalize query options and clamp (`:timeout`, `:limit`, `:offset`).
2. Detect disallowed system ops:
   - any script starting with `::` (system op; must stand alone per docs).
3. Validate relation names for helper-generated operations.
4. Optionally deny `::access_level` to prevent JS layer changing relation visibility if host wants static policy.

### Template compiler (`q` and `cq`)

Compiler contract:

1. input: template strings + values + optional query options.
2. output: `PreparedQuery { script, params, opts }`.
3. guarantees:
   - every interpolation becomes `$pN` parameter,
   - no raw string interpolation of user values into script body,
   - Cozo params map stays JSON-compatible.

Pseudo-implementation:

```go
func CompileTemplate(strings []string, values []any, opts QueryOptions) (PreparedQuery, error) {
    var b strings.Builder
    params := map[string]any{}
    for i, part := range strings {
        b.WriteString(part)
        if i < len(values) {
            key := fmt.Sprintf("p%d", i)
            b.WriteString("$" + key)
            params[key] = NormalizeValue(values[i])
        }
    }
    return PreparedQuery{Script: b.String(), Params: params, Opts: opts}, nil
}
```

### `atomic` implementation

Use Cozo chained queries in one script transaction (`stored.html:432-438`).

Compile:

1. each prepared query `q` -> `applyOpts(q.script, q.opts)`
2. wrap each script in `{ ... }`
3. join into single script and merge params (namespaced to avoid collisions).

Collision-safe merge strategy:

1. rewrite each param key with query prefix (`q0_p1`, `q1_p0`, ...).
2. rewrite `$pN` tokens in per-query script accordingly.

This yields deterministic, portable transactions without requiring Node-only `multiTransact`.

### Relation helper compilation

Map helper methods to CozoScript patterns using stored-relation docs:

1. `create(spec)` -> `:create <name> { ... }` or `:replace` when requested.
2. `put/insert/update/rm/del` -> generated input relation + mutation op + optional `:returning`.
3. `get(key)` -> `?[...] := *rel{...}` with `:limit 1`.
4. `columns()` -> `::columns <name>` (standalone system op script).
5. `indices()` -> `::indices <name>` (standalone).
6. `access(level)` -> `::access_level <level> <name>` (subject to policy).

Important Cozo constraint: system ops must appear alone (`sysops.html:265-266`).

Therefore `columns/indices/access` must execute as separate scripts, not inside chained non-system query blocks.

### Error model

Normalize backend errors to a stable JS object:

```ts
{ message: string; code?: string; display?: string; details?: unknown }
```

Adapter-specific notes:

1. Node adapter returns JSON-parsed errors in upstream (`index.js:59-61`, `72`, `84`, etc.).
2. Cozo-go adapter returns `QueryError` map with display/message fields (`cozo-lib-go-cozo.go:46-58`, `136-138`).

Unify both via one `NormalizeError(err any) CozoError` function.

### Result helper model

Requested helpers (`objects`, `firstObject`, `scalar`) should be implemented as lightweight wrappers around `headers` + `rows`.

`objects()` rules:

1. map row columns by header index,
2. if headers length != row length, either:
   - strict mode: return error,
   - default mode: map min(len(headers), len(row)) and ignore extra cells.

`scalar()` rule:

1. return first cell of first row if exists.

### Port/capability transport model

Define transport-neutral operation envelope:

```go
type OpRequest struct {
    Op        string
    Script    string
    Params    map[string]any
    Opts      QueryOptions
    Relations []string
    Data      map[string]RelationRows
}

type OpResponse struct {
    OK     bool
    Result any
    Error  *CozoError
}
```

Host-owned model:

1. JS-facing db handle in untrusted runtime calls `port.request(req)`.
2. host service validates policy and dispatches to backend.
3. response is pure JSON-compatible object.

This matches requested portability: local embedded Cozo, worker-hosted WASM, or remote RPC can all share the same interface.

### REPL runner strategy

Build `cmd/cozodb-js-repl` on top of existing `go-go-goja` bobatea adapter:

1. create evaluator with modules enabled,
2. register `cozodb` module in default registry (or custom `WithModules` composition),
3. provide bootstrapping helper script (e.g., `const db = require("cozodb").open({...})`).

Also keep a non-UI `cmd/cozodb-repl` variant for CI and quick scripting.

## Implementation Phases (File-Level)

## Phase 0: Repository prep and command scaffolding

1. Replace `cmd/XXX/main.go` with real root command(s):
   - `cozodb-goja repl`
   - `cozodb-goja run <script.js>`
2. Add package skeletons listed in proposed layout.
3. Add build tags for CGO adapter package if needed.

Deliverable:

1. compile-ready command and package skeletons with TODO stubs.

## Phase 1: Core types, codec, template compiler

1. Implement `CozoValue` normalization and `QueryOptions` handling.
2. Implement `CompileTemplate` and prepared query type.
3. Implement result helpers and tests.

Tests:

1. param substitution correctness,
2. no direct value interpolation,
3. helper behavior on empty/sparse outputs.

## Phase 2: Backend interface + Cozo-go adapter

1. Implement backend interface in `pkg/cozoapi/backend/interface.go`.
2. Implement `cozogo` adapter using `cozo-lib-go` surface.
3. Add capability reporting (`persistence`, `backupRestore`, etc.).
4. Add adapter-focused tests where CGO env available; otherwise integration tests gated by build tag/env.

## Phase 3: Policy engine and script validators

1. Implement option clamping and defaulting.
2. Implement system-op detection and allow/deny gating.
3. Implement relation name allow/deny checks for helper APIs.
4. Add tests for policy denial cases and timeout rewriting.

## Phase 4: Relation helpers and atomic compiler

1. Implement relation ops compilation.
2. Implement `atomic` chained query assembler with namespaced params.
3. Add transaction behavior tests against mem backend.

Tests:

1. generated script shape,
2. `:returning` behavior pass-through,
3. atomic all-or-nothing semantics for induced error case.

## Phase 5: Native goja module wiring

1. Implement module loader exposing promise-based async methods.
2. Ensure all backend calls happen through runtimeowner-safe scheduling for promise settlement.
3. Add integration tests using `engine.NewBuilder().WithModules(...)` + `require("cozodb")`.

## Phase 6: REPL integration and usability

1. Add `cmd/cozodb-js-repl` command.
2. Add startup helpers and docs for quick exploratory usage.
3. Verify `close()` lifecycle and command shutdown cleanup.

## Phase 7: Packaging and operational hardening

1. Add health diagnostics for missing Cozo native libraries.
2. Add Makefile checks and CI job matrix with CGO-enabled path.
3. Add fallback docs if Cozo libs unavailable.

## Test and Validation Strategy

### Unit tests

1. `template_test.go`
   - interpolation mapping,
   - escaping edge cases,
   - deterministic param naming.
2. `policy_test.go`
   - timeout clamping,
   - blocked system ops,
   - relation allowlist behavior.
3. `relation_compile_test.go`
   - `create/put/insert/update/rm/del/get` compile outputs.
4. `result_helpers_test.go`
   - object/scalar extraction behavior.

### Integration tests

1. module bootstrap:
   - require `cozodb`, open mem db, run simple query.
2. atomic chain:
   - success case returns last query result,
   - failure case rolls back.
3. policy:
   - deny `::remove`, allow normal query,
   - enforce max timeout.
4. lifecycle:
   - explicit `close()` idempotence,
   - command shutdown closes runtime.

### Optional end-to-end tests

1. REPL scripted session executes module operations.
2. capability port mock transport parity test (direct vs port dispatch produce same result shape).

## Risks, Tradeoffs, and Mitigations

### Risk 1: CGO dependency friction across environments

Description:

1. `cozo-lib-go` requires native static libraries and linker flags.

Mitigation:

1. isolate adapter behind build tags,
2. startup diagnostics with actionable remediation,
3. CI job that explicitly validates CGO path.

### Risk 2: Deadlocks or race conditions around promise resolution

Description:

1. goja promise settle functions must be invoked on VM owner loop.

Mitigation:

1. enforce single settlement path via runtimeowner `Post`,
2. add race tests similar to `pkg/runtimeowner/runner_race_test.go` style,
3. avoid blocking owner thread on operations that schedule back into owner.

### Risk 3: Overly naive script validation for policy

Description:

1. text-based checks for system ops/relation names can miss edge cases.

Mitigation:

1. treat helper-generated scripts as policy-safe by construction,
2. for raw `exec`, start conservative (deny `::` unless explicitly enabled),
3. add stricter parser-based policy mode later if needed.

### Risk 4: API drift between requested contract and real backend capabilities

Description:

1. Node-only features (`immutable`, callbacks, named rules, multiTransact) are not universal.

Mitigation:

1. publish `capabilities` on db handle,
2. keep portable core API and optional feature guards,
3. avoid exposing non-portable operations as unconditional methods.

### Tradeoff: Pure “port-only” API vs direct module methods

Decision:

1. expose direct methods for ergonomics, internally route through service dispatch API that is also used by port mode.

Benefit:

1. same policy and backend logic is reused regardless of embedding style.

## Alternatives Considered

### Alternative A: Reuse existing `database` module and map Cozo onto SQL-like calls

Rejected because:

1. current `database` module API is SQL-centric and generic,
2. CozoScript semantics (`:create`, `:put`, `:rm`, chained transactions, system ops) need dedicated abstraction,
3. requested template-tag and relation helper API would be awkward on SQL-shaped interface.

### Alternative B: Skip relation helpers and expose raw `exec` only

Rejected because:

1. misses explicit user request,
2. increases script boilerplate and repeated schema/mutation fragments,
3. loses ability to centralize policy around relation names.

### Alternative C: Implement Node-style interactive transactions (`multiTransact`) first

Rejected for initial scope because:

1. not portable across backends,
2. chained-script transactions already provide required atomicity in a portable form.

### Alternative D: Expose system ops freely

Rejected because:

1. conflicts with sandbox-safe objective,
2. docs explicitly frame access levels as guardrails, not security boundaries.

## Open Questions

1. Should `exec(rawScript)` be enabled by default in sandbox mode, or require explicit `allowRawExec` policy?
2. Should `rel.access(...)` be disabled by default even when other system ops are allowed?
3. For JS numeric keys, do we enforce string-only key fields by convention checks at helper layer?
4. Do we want strict row/header arity validation in `objects()` by default?
5. What is the target initial backend for this repo:
   - `cozogo` only (recommended for first milestone), or
   - `cozogo` + placeholder remote adapter interface?

## Recommended Initial Milestone Definition

“Done” for COJS-01 should mean:

1. JS module `require("cozodb")` available in runtime.
2. Working methods: `exec`, `q`, `cq`, `atomic`, `rel().{create,put,insert,update,rm,get}`, `export`, `import`, `close`.
3. Policy controls for timeout clamp and system-op gating.
4. Integration tests proving transaction behavior and policy enforcement.
5. REPL command wired to module for manual validation.

## Reference Evidence Index

### Requested API and behavior target

1. `sources/local/01-cozodb-js.md:67-123` query options and db handle interface.
2. `sources/local/01-cozodb-js.md:128-160` relation helper interface.
3. `sources/local/01-cozodb-js.md:197-214` atomic transaction chaining model.
4. `sources/local/01-cozodb-js.md:295-311` Node vs WASM adapter expectations.

### Cozo language and operations

1. `cozodb-queries.html:477-510` query options and timeout semantics.
2. `cozodb-stored.html:432-463` script transaction chaining semantics.
3. `cozodb-stored.html:346-351` mutation `:returning` and `_kind` behavior.
4. `cozodb-sysops.html:265-266` system-op script isolation rule.
5. `cozodb-sysops.html:345-355` access levels and guardrail warning.

### Cozo backend APIs

1. `cozo-lib-nodejs-index.js:55-65` run with `immutable` argument.
2. `cozo-lib-nodejs-index.js:51-53` `multiTransact`.
3. `cozo-lib-nodejs-index.js:128-156` callbacks and named rules.
4. `01-cozo-lib-nodejs-readme.md:54-58` explicit close requirement.
5. `02-cozo-lib-wasm-readme.md:48-55` sync string API surface.
6. `02-cozo-lib-wasm-readme.md:58-60` blocking behavior on main thread.

### go-go-goja and goja runtime constraints

1. `go-go-goja/engine/factory.go:147-177` runtime/eventloop/owner wiring.
2. `go-go-goja/pkg/runtimeowner/runner.go:62-148` call/post scheduling semantics.
3. `go-go-goja/pkg/repl/evaluators/javascript/evaluator.go:88-107` evaluator runtime construction.
4. `goja/README.md:99-103` runtime single-goroutine constraint.
5. `goja/builtin_promise.go:610-623` promise settle loop requirement.
6. `goja/runtime.go:1508-1525` interrupt lifecycle semantics.

### Prior Cozo Go integration

1. `.../sources/cozo-lib-go-cozo.go:18-25` CGO integration requirements.
2. `.../sources/cozo-lib-go-cozo.go:93-140` query execution and response decoding.
3. `.../sources/cozo-lib-go-cozo.go:142-279` import/export/backup/restore surface.

