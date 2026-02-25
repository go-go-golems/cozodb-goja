---
Title: 'Bug Report: cozo-lib-go write lock on all mutations'
Ticket: CO-04
Status: active
Topics:
    - cozodb
    - bug
    - go
    - cgo
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cozodb-goja/pkg/cozoapi/cozocgo/adapter_cozo_cgo.go
      Note: CGO adapter migrated from cozo-lib-go to cie/pkg/cozodb
    - Path: cozodb-goja/ttmp/2026/02/24/CO-04--tui-for-cozodb-relationship-explorer/scripts/01-cozo-write-lock-repro.go
      Note: Repro through cozoapi wrapper (5 isolation levels)
    - Path: cozodb-goja/ttmp/2026/02/24/CO-04--tui-for-cozodb-relationship-explorer/scripts/02-cozo-raw-cgo-repro.go
      Note: Repro calling cozo-lib-go directly (bypasses cozoapi entirely)
ExternalSources:
    - https://github.com/cozodb/cozo-lib-go
    - https://github.com/cozodb/cozo/releases/tag/v0.7.5
    - https://pkg.go.dev/github.com/kraklabs/cie/pkg/cozodb?tab=versions
Summary: "Confirmed ABI mismatch in cozo-lib-go v0.7.5 and mitigated in project by migrating adapter to github.com/kraklabs/cie/pkg/cozodb@v0.7.20"
LastUpdated: 2026-02-25T01:08:00-05:00
WhatFor: "Track and resolve the write-lock bug blocking TUI development"
WhenToUse: "When debugging CozoDB write failures through the Go CGO binding"
---

# Bug: cozo-lib-go — All Mutations Fail with "write lock required for read-only query"

## Status

Root cause confirmed and project-level mitigation implemented.

- Upstream/root issue: binding-level C API signature mismatch in `cozo-lib-go@v0.7.5`.
- Local/project status: unblocked by migrating adapter to `github.com/kraklabs/cie/pkg/cozodb@v0.7.20`.

## Summary

Every CozoDB mutation (`:create`, `:put`, `:rm`, `::` system ops) fails with the error:

```
× write lock required for read-only query
```

This happens at the lowest possible level — calling `cozo.CozoDB.Run()` directly from `github.com/cozodb/cozo-lib-go@v0.7.5` — so the bug is not in our `cozoapi` wrapper.

## Environment

| Component | Version |
|---|---|
| Go | 1.25.7 linux/amd64 |
| cozo-lib-go (failing repro) | v0.7.5 |
| cie/pkg/cozodb (mitigation) | v0.7.20 |
| libcozo_c.a | v0.7.5 (`libcozo_c-0.7.5-x86_64-unknown-linux-gnu.a.gz`) |
| OS | Ubuntu 24.04 (Linux 6.8.0-90-generic) |

## Reproduction

### Minimal (scripts/02-cozo-raw-cgo-repro.go)

```go
db, _ := cozo.New("mem", "", nil)
res, err := db.Run(`:create test {id: String => val: String}`, nil)
// err = "× write lock required for read-only query"
// res.Ok = false
```

Both `mem` and `sqlite` engines produce the same error.

### Run the repro scripts

```bash
cd cozodb-goja
CGO_LDFLAGS="-L$PWD/.deps/cozo" GOWORK=off go run -tags cozo_cgo \
  ./ttmp/.../scripts/02-cozo-raw-cgo-repro.go
```

### Full test matrix (scripts/01-cozo-write-lock-repro.go)

All 7 tests fail:

| Test | Path | Result |
|---|---|---|
| 1 | `backend.Exec()` zero opts | FAIL |
| 2 | `backend.Exec()` with `:timeout 3` | FAIL |
| 3 | `DB.ExecScript()` DefaultPolicy (no timeout) | FAIL |
| 4 | `DB.ExecScript()` policy with DefaultTimeoutSec=3 | FAIL |
| 5 | `DB.ExecScript()` explicit TimeoutSec=0 | FAIL |
| 6 | `:put` after `:create` | FAIL (no relation created) |
| 7 | Read query | FAIL (relation_not_found) |

## Analysis

### What we ruled out

- **Not our cozoapi wrapper**: Test 1 bypasses the entire wrapper, calls `cozocgo.backend.Exec()` with empty options — still fails.
- **Not the timeout directive**: Test 1 has no `:timeout` appended — still fails.
- **Not the policy layer**: Test 1 has no policy enforcement at all — still fails.
- **Not engine-specific**: Both `mem` and `sqlite` engines fail identically.
- **Not the query syntax**: The exact same `:create` syntax works in Python (`pycozo`) and in CozoDB's own web console.

### Confirmed root cause

The call chain is:

```
Go: db.Run(query, nil)
  → C: cozo_run_query(db_id, script, "{}")
    → Rust: CozoDB internal query dispatch
```

In `cozo-lib-go@v0.7.5`, the bundled header/wrapper declares and calls:

```
cozo_run_query(int32_t db_id, const char *script_raw, const char *params_raw)
```

But Cozo's C API (`cozo-lib-c/cozo_c.h` in `cozo@v0.7.5` and `v0.7.6`) declares:

```
cozo_run_query(int32_t db_id, const char *script_raw, const char *params_raw, bool immutable_query)
```

That missing 4th parameter is the mutability flag. A 3-arg call against a 4-arg ABI causes the immutable flag to be garbage/incorrect at runtime, which explains deterministic read-only behavior for mutations.

### Proof

Added a direct C API repro script with the correct 4-arg signature:

- `scripts/03-cozo-c-api-4arg-repro.go`

Running it with the same `libcozo_c.a` succeeds for `:create`, `:put`, and subsequent read query:

- `ok=true` for mutations and read.

So the static library itself is not read-only; the failure is in the Go binding call signature.

## Impact

This blocks all TUI development — we can't create schemas, insert data, or run any mutations through the Go binding. Read queries would work if relations existed, but we can't create them.

## Project Fix (Implemented)

Applied mitigation in `cozodb-goja` by replacing the adapter wrapper from `cozo-lib-go` to `cie/pkg/cozodb@v0.7.20`.

Validation after migration:

- `scripts/01-cozo-write-lock-repro.go` now reports `<nil>` for all mutation tests.
- read query at the end returns created data (`headers: [id val]`, rows contain inserted tuple).
- adapter export/import path also works via COJS-02 validation script.

## Workarounds

1. **Applied:** migrated adapter to `github.com/kraklabs/cie/pkg/cozodb@v0.7.20` (correct 4-arg call path).
2. **Alternative:** patch/fork `cozo-lib-go` so `Run` calls `cozo_run_query(..., immutable_query)`.
3. **Alternative:** bypass wrapper and call C API directly from adapter (validated in script 03).
4. **Alternative:** use CozoDB HTTP API for writes until upstream wrapper is fixed.

## Repro Scripts

- `scripts/01-cozo-write-lock-repro.go` — Tests 5 isolation levels through the cozoapi wrapper
- `scripts/02-cozo-raw-cgo-repro.go` — Bypasses cozoapi entirely, calls cozo-lib-go directly
- `scripts/03-cozo-c-api-4arg-repro.go` — Calls C API directly with correct 4th `immutable_query` arg (works)
