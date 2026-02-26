---
Title: Diary
Ticket: CO-10
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
    - Path: ttmp/2026/02/25/CO-10--phase-6-hard-cutover-remove-legacy-tui-and-harden-library-boundary/design/01-implementation-plan-phase-6-hard-cutover-and-boundary-cleanup.md
      Note: Phase 6 plan referenced by diary
    - Path: ttmp/2026/02/25/CO-10--phase-6-hard-cutover-remove-legacy-tui-and-harden-library-boundary/tasks.md
      Note: Phase 6 task queue referenced by diary
ExternalSources: []
Summary: Implementation diary for CO-10
LastUpdated: 2026-02-25T12:35:00-05:00
WhatFor: Track phase 6 hard cutover and boundary hardening
WhenToUse: Use when reviewing CO-10 execution progress
---


# Diary

## Goal

Track CO-10 implementation decisions and execution history once Phase 6 begins.

## Step 1: Diary Initialization and Queueing

CO-10 implementation is pending earlier phase completion. This entry initializes the diary and records current dependency ordering.

### Prompt Context

**User prompt (verbatim):** "alright, implement it now, task by task, committing as you go, keeping a diary in each ticket."

**Assistant interpretation:** Keep ticket-level diaries active from the start, even for queued phases.

**Inferred user intent:** Enforce strong implementation governance across all tickets.

**Commit (code):** N/A

### What I did
- Created CO-10 diary document.
- Marked phase as queued behind CO-07 through CO-09 delivery.

### Why
- Needed to satisfy per-ticket diary requirement.

### What worked
- Diary framework is now in place for cutover/hardening entries.

### What didn't work
- N/A

### What I learned
- Explicit queue-state notes avoid confusion about perceived inactivity.

### What was tricky to build
- Keeping this precise without implying Phase 6 started early.

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- Add first implementation entry immediately after first CO-10 commit.

### Code review instructions
- Confirm diary file presence and that future entries reference real CO-10 commits.

### Technical details
- Phase dependency: hard cutover ticket starts after functional migration tickets are done.

## Step 2: Hard-Cutover Closure, Guardrail Wiring, and Ticket Closeout

The phase was closed by combining already-merged legacy deletion work with explicit boundary guardrails and final ticket hygiene. The key objective was to prevent accidental reintroduction of `cozodb-goja` runtime app code while keeping the relocated extraction-side app as the only canonical path.

I validated both repos with tests, ran boundary scans with `rg`, added a guard script + CI hook, and synchronized CO-10/CO-05/CO-06 docs so closure checklists are consistent with the hard-cutover state.

### Prompt Context

**User prompt (verbatim):** "ok, close it up."

**Assistant interpretation:** Finish remaining CO-10 closure work now, including validations, guardrails, and ticket bookkeeping.

**Inferred user intent:** Formally complete Phase 6 with enforceable safeguards and updated documentation.

**Commit (code):** `860592d` — "CO-10: close hard-cutover with legacy-path guardrails"

### What I did
- Verified no active legacy references in `cozodb-goja` with `rg ... --glob '!ttmp/**'`.
- Ran `go test ./... -count=1` in:
  - `cozodb-goja`
  - `2026-02-18--cozodb-extraction/cozo-extraction-tui`
- Ran an interactive launch smoke in PTY:
  - `go run ./cmd/cozo-tui --runtime-engine mem` (manual interrupt after launch).
- Added guard script:
  - `ttmp/.../CO-10.../scripts/01-guard-no-legacy-tui-paths.sh`
- Added `Makefile` target:
  - `guard-no-legacy-tui-paths`
- Added CI workflow step in `.github/workflows/push.yml` to run guard before tests.
- Updated CO-10 task checklist/changelog for completion evidence.
- Updated CO-05 and CO-06 docs with canonical relocated-path notes.
- Ran doc hygiene:
  - `docmgr doc relate --ticket CO-10 ...`
  - `docmgr doctor --ticket CO-10 --stale-after 30`

### Why
- Hard cutover is only robust if regressions fail fast in automation, not only by reviewer memory.

### What worked
- Guard script catches path regressions and passes with current tree.
- Tests are green in both modules after cleanup.
- `docmgr doctor` is clean for CO-10 after related-file cleanup.

### What didn't work
- First guard implementation failed locally because an empty leftover `internal/tui/` directory existed in the working tree:
  - `error: legacy path exists: internal/tui/`
- First `docmgr doc relate` removal command failed due `--remove-files` argument format:
  - `Error: Too many arguments`

### What I learned
- Guard checks should assert on legacy **source files**, not empty untracked directories.
- `docmgr doc relate --remove-files` expects comma-separated values in a single argument.

### What was tricky to build
- Distinguishing actionable code references from historical ticket docs required explicit `rg` scoping (`--glob '!ttmp/**'` and non-markdown filters).

### What warrants a second pair of eyes
- CI coverage scope: current guard is in `cozodb-goja` pipeline; if future workflows split, ensure this check remains mandatory for PR gates.

### What should be done in the future
- N/A

### Code review instructions
- Start with:
  - `ttmp/.../CO-10.../scripts/01-guard-no-legacy-tui-paths.sh`
  - `Makefile`
  - `.github/workflows/push.yml`
  - `ttmp/.../CO-10.../tasks.md`
  - `ttmp/.../CO-10.../changelog.md`
- Validate with:
  - `make guard-no-legacy-tui-paths`
  - `go test ./... -count=1` in both modules
  - `docmgr doctor --ticket CO-10 --stale-after 30`

### Technical details
- `docmgr doctor` initially warned about stale related file `cozodb-goja/cmd/cozo-tui/main.go`; resolved by removing obsolete related-file entries and adding relocated command path.

## Step 3: Hard-Cut `geppetto/plugins` Contract to Local `cozo/plugins` Ownership

This step completed a no-backward-compatibility cutover for the extractor plugin descriptor helper module. We removed active-path dependency on `require("geppetto/plugins")` and moved plugin-contract runtime ownership into the relocated extraction runtime.

### Prompt Context

**User prompt (verbatim):** "remove the defineExtractor stuff entirely from cozodb-goja. Move everything in geppetto/ hard cut over to the cozo-extraction-tui. Add tasks to the ticket."

**Assistant interpretation:** Perform immediate hard cutover: localize plugin helper runtime ownership in extraction module, migrate scripts/tests, and scrub `defineExtractor` usage from `cozodb-goja` ticket artifacts.

**Inferred user intent:** Keep Cozo-specific JS plugin contract mechanics out of shared Geppetto module ownership and keep ticket traceability tight.

### What I did
- In `cozo-extraction-tui`:
  - copied JS `geppetto` module runtime integration into local package:
    - `internal/jsmodules/geppetto/*`
  - removed plugin submodule registration from local copied register path (`module.go`).
  - added new local native module:
    - `internal/jsmodules/cozoplugins/module.go`
    - module name: `cozo/plugins`
    - exports: `EXTRACTOR_PLUGIN_API_VERSION`, `defineExtractorPlugin`, `wrapExtractorRun`.
  - rewired runtime module registration:
    - `internal/geppettohost/runtime.go` now imports local `internal/jsmodules/geppetto` and registers `cozo/plugins`.
  - migrated active code/docs from `geppetto/plugins` to `cozo/plugins`:
    - `scripts/*`
    - `internal/*_test.go`
    - `pkg/doc/*`.
  - removed copied stale `generate.go` from local jsmodule copy to avoid invalid `go generate` paths.
  - ran `gofmt` on edited Go files.
- In `cozodb-goja`:
  - converted ticket probe scripts to explicit descriptor exports (no helper import):
    - `ttmp/.../COJS-01.../scripts/02-07*.js`
  - converted CO-05 template script similarly:
    - `ttmp/.../CO-05.../scripts/01-extractor-plugin-template.js`.

### Why
- Cozo extractor plugin contract helpers are Cozo workflow-specific and should be owned by the Cozo runtime app boundary.

### What worked
- `GOWORK=off go test ./... -count=1` passed in `cozo-extraction-tui`.
- `cozo-plugin-run` fixture smoke passed with `cozo/plugins`:
  - `printf 'Alice met Bob at launch.' | go run ./cmd/cozo-plugin-run --plugin-script ./scripts/fixtures/extractor_fixture.js --plugin-transcript - --plugin-pretty=false`
  - returned valid extraction JSON and metadata.
- `cozodb-goja` now has no non-markdown occurrences of:
  - `defineExtractorPlugin`
  - `wrapExtractorRun`
  - `geppetto/plugins`.

### What didn't work
- Initial targeted test pass failed due missing `go.sum` entry for `go-emrichen` transitively used by `geppetto/pkg/embeddings`:
  - fixed with `GOWORK=off go mod tidy` in `cozo-extraction-tui`.

### What I learned
- Copying the JS module package locally is straightforward, but generated-file scaffolding (`generate.go`) must be pruned/updated to keep repo-local generation sane.

### What was tricky to build
- Preserving existing extraction script ergonomics while hard-cutting module ownership required adding `cozo/plugins` with parity exports before migration.

### What warrants a second pair of eyes
- Decide if any remaining Geppetto module references in historical ticket docs should be normalized or intentionally preserved as historical record.

### What should be done in the future
- Consider a follow-up that migrates plugin scripts from helper wrapper style to plain descriptor exports everywhere for simpler host contract semantics.

### Code review instructions
- Review runtime/module ownership:
  - `2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/geppettohost/runtime.go`
  - `2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/jsmodules/cozoplugins/module.go`
  - `2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/jsmodules/geppetto/module.go`
- Review script migration:
  - `2026-02-18--cozodb-extraction/cozo-extraction-tui/scripts/**/*.js`
  - `cozodb-goja/ttmp/2026/02/24/COJS-01-INITIAL-BUILD--initial-build-plan-for-cozodb-goja-js-api/scripts/*.js`
- Re-run:
  - `GOWORK=off go test ./... -count=1` in `cozo-extraction-tui`
  - `printf 'Alice met Bob at launch.' | GOWORK=off go run ./cmd/cozo-plugin-run --plugin-script ./scripts/fixtures/extractor_fixture.js --plugin-transcript - --plugin-pretty=false`.

### Technical details
- `cozo/plugins` helper semantics were intentionally kept source-compatible with prior helper behavior to keep script-level migration limited to module-path changes in extraction repo.
