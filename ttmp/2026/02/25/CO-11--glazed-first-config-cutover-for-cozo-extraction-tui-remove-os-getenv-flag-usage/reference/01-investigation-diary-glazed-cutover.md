---
Title: 'Investigation diary: glazed cutover'
Ticket: CO-11
Status: active
Topics:
    - cozo
    - tui
    - glazed
    - geppetto
    - embeddings
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/cmd/cozo-plugin-run/main.go
      Note: |-
        Verified second binary still using flag package
        Diary evidence for legacy flag usage
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/cmd/cozo-tui/main.go
      Note: |-
        Verified flag parsing plus env-bridge behavior
        Diary evidence for current env bridge
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/geppettohost/embedder.go
      Note: |-
        Catalogued direct os.Getenv usage and profile/config file merge logic
        Diary evidence for direct os.Getenv usage
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/geppettohost/host.go
      Note: Verified lazy provider construction path
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/screens/vsearch/commands.go
      Note: Verified environment-based script root behavior
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/screens/vsearch/model.go
      Note: |-
        Verified environment toggle in UI model
        Diary evidence for env-enabled behavior
    - Path: ../../../../../../../geppetto/pkg/embeddings/config/settings.go
      Note: Verified glazed-tagged embeddings config structure
    - Path: ../../../../../../../geppetto/pkg/embeddings/settings_factory.go
      Note: Verified parsed-values decode entrypoint for provider construction
    - Path: ../../../../../../../geppetto/pkg/sections/profile_registry_source.go
      Note: Captured profile registry adapter semantics
    - Path: ../../../../../../../geppetto/pkg/sections/profile_registry_source_test.go
      Note: |-
        Captured tested precedence behavior and schema assembly helper usage
        Diary evidence for effective precedence behavior
    - Path: ../../../../../../../geppetto/pkg/sections/sections.go
      Note: |-
        Captured Geppetto bootstrap and precedence middleware implementation
        Diary evidence for bootstrap middleware approach
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/commands/tui.go
      Note: Glazed TUI command execution path and parser option-order fix
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/commands/plugin_run.go
      Note: Glazed plugin-run command execution path and parser option-order fix
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/config/settings.go
      Note: Typed sections and aggregate execution settings decode helpers
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/config/middlewares.go
      Note: Cozo middleware bootstrap/preference chain implementation
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/config/middlewares_test.go
      Note: Precedence and profile bootstrap tests for cozo middleware chain
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/config/embeddings_test.go
      Note: Provider construction tests from parsed values
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/commands/decode_test.go
      Note: Command typed-settings decode coverage for tui and plugin-run
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/Makefile
      Note: Repeatable cozo_cgo validation + real-seed run targets
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/README.md
      Note: Glazed-only invocation examples and hard-cutover note
ExternalSources: []
Summary: Chronological research diary for CO-11 covering discovery, evidence capture, architectural decisions, and delivery steps.
LastUpdated: 2026-02-25T17:05:00-05:00
WhatFor: Preserve command-level and middleware-level evidence supporting Glazed-first cutover planning.
WhenToUse: Use when implementing CO-11 tasks or reviewing why the chosen migration architecture was selected.
---


# Investigation diary: glazed cutover

## Goal

Produce a deep, implementation-ready migration design for replacing `flag` + `os.Getenv` configuration with Glazed sections, middleware loading, and typed settings decode across the cozo extraction TUI project.

## Context

The user requested a hard move away from direct environment lookups and non-Glazed flags, with explicit focus on how Geppetto currently solves middleware/config/profile loading.

## Chronological log

## Step 1: Skill and workflow initialization

Commands:

```bash
sed -n '1,260p' /home/manuel/.codex/skills/glazed-command-authoring/SKILL.md
sed -n '1,320p' /home/manuel/.codex/skills/ticket-research-docmgr-remarkable/SKILL.md
docmgr status --summary-only
```

Findings:

1. `glazed-command-authoring` requires command structs, section definitions, `DecodeSectionInto`, and `cli.BuildCobraCommand*` integration.
2. `ticket-research-docmgr-remarkable` requires ticket creation, evidence-first writing, diary, `docmgr doctor`, and reMarkable upload.
3. Current doc workspace is healthy (`stale=0`).

Decision:

Use both skills in sequence exactly as requested.

## Step 2: Glazed topical docs reconnaissance

Commands:

```bash
glaze help
glaze help middlewares-guide
glaze help sections-guide
glaze help parsing-fields
glaze help profiles
glaze help implementing-profile-middleware
glaze help config-files-quickstart
```

Findings:

1. Middlewares execute in reverse of the slice order provided to execution.
2. Profiles documentation emphasizes bootstrap parsing for profile-selection circularity.
3. Config quickstart confirms overlay model and parse provenance tooling (`--print-parsed-fields`).
4. The docs describe precedence in conceptual terms but implementation ordering must be verified in code.

Decision:

Treat Glazed docs as conceptual baseline, then validate exact precedence from Geppetto source/tests before making recommendations.

## Step 3: Current cozo-extraction-tui code evidence collection

Commands:

```bash
rg -n "os\.Getenv|COZO_TUI_|flag\." /home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui -g'*.go'
nl -ba .../cozo-extraction-tui/cmd/cozo-tui/main.go | sed -n '1,280p'
nl -ba .../cozo-extraction-tui/cmd/cozo-plugin-run/main.go | sed -n '1,320p'
nl -ba .../cozo-extraction-tui/internal/geppettohost/embedder.go | sed -n '1,460p'
nl -ba .../cozo-extraction-tui/internal/geppettohost/host.go | sed -n '1,220p'
nl -ba .../cozo-extraction-tui/internal/tui/screens/vsearch/model.go | sed -n '1,420p'
nl -ba .../cozo-extraction-tui/internal/tui/screens/vsearch/commands.go | sed -n '1,260p'
```

Findings:

1. `cozo-tui` uses `flag.Parse()` and then mutates process env through `os.Setenv` bridge helper.
2. `cozo-plugin-run` is entirely `flag`-based.
3. `geppettohost/embedder.go` contains direct env and profile/config file loading logic.
4. `host.ensureEmbedder()` lazily initializes provider via env-backed helper.
5. F9 still reads env for auto-migration and script root.

Decision:

Migration scope must include both command entrypoints and internal host/F9 runtime dependency injection.

## Step 4: Geppetto middleware implementation analysis

Commands:

```bash
rg -n "MiddlewaresFunc|CobraParserConfig|CreateGeppettoSections|GetCobraCommandGeppettoMiddlewares|PINOCCHIO|profile" /home/manuel/workspaces/2026-02-24/cozodb-goja-init/geppetto/pkg/sections -g'*.go'
nl -ba .../geppetto/pkg/sections/sections.go | sed -n '1,360p'
nl -ba .../geppetto/pkg/sections/profile_registry_source.go | sed -n '1,260p'
nl -ba .../geppetto/pkg/sections/profile_registry_source_test.go | sed -n '90,340p'
```

Findings:

1. Geppetto performs command/profile bootstrap parse before creating profile middleware.
2. Middleware chain comments explicitly discuss reverse execution and why config/profile order is arranged as coded.
3. Tests assert profile override over config, env override over profile, and flags override env.
4. Implementation is currently hardcoded around `PINOCCHIO` prefix and `pinocchio` app path.

Decision:

CO-11 should reuse the pattern, not the hardcoded helper function directly.

## Step 5: Embeddings section/decode path verification

Commands:

```bash
nl -ba .../geppetto/pkg/embeddings/config/settings.go | sed -n '1,240p'
nl -ba .../geppetto/pkg/embeddings/settings_factory.go | sed -n '1,280p'
nl -ba .../geppetto/cmd/examples/simple-streaming-inference/main.go | sed -n '1,190p'
nl -ba .../geppetto/cmd/examples/simple-streaming-inference/main.go | sed -n '245,320p'
```

Findings:

1. Geppetto embeddings config already has `glazed` tags for provider/model/dimensions/API/base URLs.
2. `embeddings.NewSettingsFactoryFromParsedValues` already decodes sections into a typed config.
3. Example command shows canonical command struct + decode + custom middlewares integration.

Decision:

CO-11 should construct embeddings providers from parsed values in command layer and inject provider into host, eliminating env lookups from host runtime.

## Step 6: Ticket creation and doc scaffolding

Commands:

```bash
docmgr ticket create-ticket --ticket CO-11 --title "Glazed-first config cutover for cozo-extraction-tui (remove os.Getenv/flag usage)" --topics cozo,tui,glazed,geppetto,embeddings
docmgr doc add --ticket CO-11 --doc-type design-doc --title "Glazed-first configuration architecture and implementation plan"
docmgr doc add --ticket CO-11 --doc-type reference --title "Investigation diary: glazed cutover"
```

Results:

1. Ticket created at `ttmp/2026/02/25/CO-11--glazed-first-config-cutover-for-cozo-extraction-tui-remove-os-getenv-flag-usage`.
2. Design doc and diary doc created.

## Step 7: Synthesis and recommendation drafting

Work performed:

1. Mapped concrete gaps between current code and target model.
2. Drafted target architecture for Glazed-first section model and typed settings decode.
3. Defined phased migration plan from command foundation through hard cleanup.
4. Captured risk of precedence mismatch and hardcoded app identity in Geppetto helper.

Key decision captured:

Adopt Geppetto-tested precedence for first cutover (`flags > env > profiles > config > defaults`) unless product policy explicitly requires config overriding profiles.

## Step 8: Bookkeeping and validation

Commands:

```bash
docmgr doc relate --doc ttmp/2026/02/25/CO-11--.../design-doc/01-glazed-first-configuration-architecture-and-implementation-plan.md --file-note "/abs/path:reason" ...
docmgr doc relate --doc ttmp/2026/02/25/CO-11--.../reference/01-investigation-diary-glazed-cutover.md --file-note "/abs/path:reason" ...
docmgr doctor --ticket CO-11 --stale-after 30
docmgr vocab add --category topics --slug cozo --description "..."
docmgr vocab add --category topics --slug embeddings --description "..."
docmgr vocab add --category topics --slug geppetto --description "..."
docmgr vocab add --category topics --slug glazed --description "..."
docmgr doctor --ticket CO-11 --stale-after 30
```

Results:

1. Related file mappings were updated for design and diary documents.
2. First doctor run produced topic vocabulary warnings for new slugs.
3. Added missing topic vocabulary entries and reran doctor.
4. Final doctor result: `✅ All checks passed`.

## Step 9: reMarkable delivery

Commands:

```bash
remarquee status
remarquee cloud account --non-interactive
remarquee upload bundle --dry-run ttmp/2026/02/25/CO-11--.../index.md ttmp/2026/02/25/CO-11--.../design-doc/01-glazed-first-configuration-architecture-and-implementation-plan.md ttmp/2026/02/25/CO-11--.../reference/01-investigation-diary-glazed-cutover.md ttmp/2026/02/25/CO-11--.../tasks.md ttmp/2026/02/25/CO-11--.../changelog.md --name \"CO-11 Glazed Config Cutover Research\" --remote-dir \"/ai/2026/02/25/CO-11\" --toc-depth 2
remarquee upload bundle ttmp/2026/02/25/CO-11--.../index.md ttmp/2026/02/25/CO-11--.../design-doc/01-glazed-first-configuration-architecture-and-implementation-plan.md ttmp/2026/02/25/CO-11--.../reference/01-investigation-diary-glazed-cutover.md ttmp/2026/02/25/CO-11--.../tasks.md ttmp/2026/02/25/CO-11--.../changelog.md --name \"CO-11 Glazed Config Cutover Research\" --remote-dir \"/ai/2026/02/25/CO-11\" --toc-depth 2
remarquee cloud ls /ai/2026/02/25/CO-11 --long --non-interactive
```

Results:

1. Dry run succeeded with expected file bundle and destination.
2. Upload succeeded: `CO-11 Glazed Config Cutover Research.pdf`.
3. Remote listing confirmed file presence under `/ai/2026/02/25/CO-11`.

## Step 10: hard-cutover policy lock and task restructuring

Prompt context:

1. User requested the simplest path and explicitly confirmed hard cutover with no backwards compatibility.
2. User then requested task creation aligned to that policy.

Actions:

1. Replaced the CO-11 task list with a hard-cutover implementation backlog.
2. Added explicit locked decisions in tasks:
   - no backwards compatibility,
   - precedence fixed to `flags > env > profiles > config > defaults`,
   - remove `flag` and runtime `os.Getenv` usage in scope.
3. Updated index current-status and changelog to reflect policy lock.

Outcome:

1. Ticket now has implementation tasks that match the confirmed direction exactly.
2. Remaining work is implementation-only; research phase is complete.

## Step 11: implementation slice 1 (command/runtime hard cutover)

I executed the first implementation slice directly in `2026-02-18--cozodb-extraction/cozo-extraction-tui` and committed it as one coherent runtime cutover unit. The scope was to move command entrypoints to Glazed, remove non-test runtime env lookups in the cutover surface, and force embedding provider injection instead of lazy env fallback.

This step intentionally prioritized “code compiles + tests green” before adding new precedence/decode tests and docs updates, so subsequent slices can focus on validation depth and operator-facing documentation without mixing structural migration changes.

### Prompt Context

**User prompt (verbatim):** "ok, now implement it , task by task, committing as you go, keeping a diary"

**Assistant interpretation:** Implement CO-11 in incremental commits, update task checkboxes, and keep a detailed ticket diary with command/test evidence.

**Inferred user intent:** Complete the hard-cutover migration with auditable progress and no hidden partial work.

**Commit (code):** `96cc0b9` — "cozo-tui: cut over to glazed commands and injected embed providers"

### What I did

1. Added Glazed command implementations:
   - `internal/commands/tui.go`
   - `internal/commands/plugin_run.go`
2. Added app config package:
   - `internal/config/settings.go`
   - `internal/config/middlewares.go`
   - `internal/config/embeddings.go`
3. Replaced entrypoint `flag` parsing with Glazed/Cobra execution:
   - `cmd/cozo-tui/main.go`
   - `cmd/cozo-plugin-run/main.go`
4. Added explicit option plumbing:
   - `internal/tui/app/model.go` (`NewWithOptions`)
   - `internal/tui/screens/vsearch/model.go` (`Options`, `NewWithOptions`)
5. Removed runtime env fallback for embeddings:
   - `internal/geppettohost/host.go` (`ensureEmbedder` now returns unavailable if missing provider)
   - `internal/geppettohost/embedder.go` reduced to `embedGlobals` only
6. Switched seed path to explicit embedding provider injection:
   - `internal/tui/seeddata/seed.go` (`Options{EmbeddingProvider ...}`)
   - updated callsites/tests in `internal/tui/testutil/db.go` and `internal/tui/seeddata/seed_live_cozo_cgo_test.go`
7. Removed stale env-bridge test coverage:
   - deleted `cmd/cozo-tui/main_test.go`
   - replaced `internal/geppettohost/embedder_test.go` with tests that match new behavior
8. Fixed compile/test fallout and validated:
   - `go test ./... -count=1` in `cozo-extraction-tui` passed.

### Why

1. Hard cutover requires deleting compatibility behaviors, not layering shims.
2. Embedding provider ownership must be command-boundary/parsed-values driven for deterministic config semantics.
3. F9 and seed needed constructor-level inputs so runtime code stops reading process env in non-test paths.

### What worked

1. New command layer cleanly decodes typed settings sections and executes without `flag`.
2. Removing env fallback in host surfaced missing provider paths explicitly and did not break unit tests after targeted updates.
3. App/vsearch option plumbing removed F9 env dependencies while keeping behavior intact.
4. Full non-tagged test suite passed after updates.

### What didn't work

1. First `go test ./... -count=1` failed with compile errors:
   - `undefined: app.NewWithOptions`
   - `undefined: app.Options`
   - `undefined: vsearch.Options`
2. Stale geppettohost tests failed because removed symbols no longer existed:
   - `undefined: loadPinocchioEmbeddingsConfig`
   - `undefined: defaultEmbeddingProviderFromEnv`
   - `undefined: applyEmbeddingEnvOverrides`
3. Resolution:
   - Added app/vsearch options APIs and rewired callers.
   - Replaced old env-centric tests with behavior-valid tests for current API.

### What I learned

1. The migration boundary is cleanest when all embedding construction remains in command/config layer and runtime receives fully built dependencies.
2. F9 model defaults were tightly coupled to env toggles; option constructors are the least invasive replacement.
3. Keeping one large structural commit for this slice was practical because compile failures were all directly caused by the same architectural move.

### What was tricky to build

1. Cause: `internal/commands/tui.go` was written against APIs that did not exist yet (`app.NewWithOptions`, `vsearch.Options`).
2. Symptoms: immediate build failure on command package and both binaries.
3. Approach:
   - Added app-level options wrapper.
   - Added vsearch options struct + constructor.
   - Replaced env-dependent default embed logic with options-driven host construction.
4. Result: compile restored and tests green.

### What warrants a second pair of eyes

1. Middleware ordering in `internal/config/middlewares.go` should be reviewed against expected reverse execution semantics.
2. Plugin-run setting names/prefixes should be validated against intended CLI UX (prefixed vs non-prefixed flags).
3. Seed “needs provider” decision for mem/seed modes should be rechecked for desired product behavior.

### What should be done in the future

1. Add explicit precedence and profile-bootstrap tests (next CO-11 slice).
2. Update Makefile/README/help examples to Glazed-prefixed flags.
3. Run tagged/live/manual smoke checks after `.deps`/credential setup verification.

### Code review instructions

1. Start in command boundary:
   - `cozo-extraction-tui/internal/commands/tui.go`
   - `cozo-extraction-tui/internal/commands/plugin_run.go`
2. Then inspect config chain:
   - `cozo-extraction-tui/internal/config/settings.go`
   - `cozo-extraction-tui/internal/config/middlewares.go`
3. Then runtime dependency injection:
   - `cozo-extraction-tui/internal/geppettohost/host.go`
   - `cozo-extraction-tui/internal/tui/seeddata/seed.go`
   - `cozo-extraction-tui/internal/tui/screens/vsearch/model.go`
   - `cozo-extraction-tui/internal/tui/screens/vsearch/commands.go`
4. Validate with:
   - `go test ./... -count=1`

### Technical details

Commands run in this step:

```bash
go test ./... -count=1
git commit -m "cozo-tui: cut over to glazed commands and injected embed providers"
```

## Step 12: implementation slice 2 (validation, docs, operator runbook)

I executed the second implementation slice to close the remaining CO-11 validation and rollout tasks. This step added explicit tests for middleware precedence/profile bootstrap, typed settings decode, and provider construction, then updated Makefile and README for repeatable Glazed-first operator workflows.

During live validation, I hit two runtime issues (`--profile-file` missing and embedding dimension mismatch at seed time) and addressed both in this same slice so the documented runbook reflects what actually works in the current codebase.

### Prompt Context

**User prompt (verbatim):** (same as Step 11)

**Assistant interpretation:** Continue implementing unfinished CO-11 tasks and keep committing/reporting each completed slice.

**Inferred user intent:** Finish the hard-cutover ticket end-to-end with executable validation commands and accurate operator docs.

**Commit (code):** `7fe59cf` — "cozo-tui: add precedence/decode tests and glazed run docs"

### What I did

1. Added validation coverage:
   - `internal/config/middlewares_test.go`
   - `internal/config/embeddings_test.go`
   - `internal/commands/decode_test.go`
2. Added aggregate execution settings decode helpers:
   - `internal/config/settings.go`
3. Fixed parser option-order bug so profile flags are registered:
   - `internal/commands/tui.go`
   - `internal/commands/plugin_run.go`
4. Updated operator tooling:
   - `Makefile` (Glazed runtime flags + real-seed target)
   - `README.md` (hard-cutover usage + precedence note)
5. Ran full validation suite:
   - `go test ./... -count=1`
   - `make test-cgo-vsearch`
   - `make run-cgo-seed-only-real ...`
   - Manual PTY run of `cozo-tui` and clean exit with `q`.

### Why

1. CO-11 required explicit precedence/decode/provider tests, not just passing runtime behavior.
2. The Makefile and README needed to reflect post-cutover CLI so teammates stop using removed legacy flags.
3. Live smoke commands had to be repeatable and documented against real profile/config credentials path.

### What worked

1. Added tests passed and validated intended precedence model.
2. `.deps` native vector test path remained green.
3. Real seed-only run succeeded with parsed-values profile/config overlays after dimension override.
4. Manual interactive TUI startup and quit worked in a real PTY.

### What didn't work

1. First real-seed run failed:
   - `error: unknown flag: --profile-file`
   - Root cause: `WithParserConfig(...)` was applied after `WithProfileSettingsSection()`, overwriting `EnableProfileSettingsSection`.
2. First real-seed run with profile defaults failed:
   - `seed database: embed person sarah_martinez: expected 384 embedding dimensions, got 1536`
   - Root cause: seed schema enforces 384-dim vectors while selected profile defaulted to 1536.
3. Resolution:
   - Reordered command builder options so profile section remains enabled.
   - Ran seed with `SEED_EMBED_FLAGS='--embeddings-dimensions 384'`.

### What I learned

1. In Glazed builder options, `WithParserConfig` is a full struct assignment and can silently clear prior toggles.
2. Real-profile smoke testing is mandatory because defaults in shared profile files can violate local schema invariants.
3. For this codebase, embedding dimensions need explicit operator control in seed workflows.

### What was tricky to build

1. Cause: command option order interaction was non-obvious and only surfaced at runtime (`--profile-file` missing).
2. Symptoms: help output excluded profile flags and real-seed target failed immediately.
3. Approach:
   - Reproduced with direct `--help` output.
   - Verified Glazed option application semantics.
   - Reordered options and reran real-seed flow.
4. Result: profile settings flags became available and parsed-values profile loading worked.

### What warrants a second pair of eyes

1. Help output currently shows geppetto ai flag alias annotations that include `--profile`; verify this UX is acceptable.
2. Seed schema dimension is fixed at 384 while profile defaults can be 1536; decide if schema should be parameterized in a later ticket.
3. Parsed-fields output from shared config sources can carry sensitive metadata; ensure operator guidance avoids accidental secret logging.

### What should be done in the future

1. Add a dedicated command/help section explaining recommended profile names for 384-dim seed workflows.
2. Add a safety guard that warns early when configured dimensions differ from schema-required dimensions.

### Code review instructions

1. Start with runtime bug fix:
   - `internal/commands/tui.go`
   - `internal/commands/plugin_run.go`
2. Review tests:
   - `internal/config/middlewares_test.go`
   - `internal/config/embeddings_test.go`
   - `internal/commands/decode_test.go`
3. Review operator docs:
   - `Makefile`
   - `README.md`
4. Re-run:
   - `go test ./... -count=1`
   - `make test-cgo-vsearch`
   - `make run-cgo-seed-only-real SEED_DB=/tmp/cozo-extraction-tui-seed-real3.db SEED_EMBED_FLAGS='--embeddings-dimensions 384'`

### Technical details

Commands run in this step:

```bash
go test ./... -count=1
make test-cgo-vsearch
make run-cgo-seed-only-real SEED_DB=/tmp/cozo-extraction-tui-seed-real3.db SEED_EMBED_FLAGS='--embeddings-dimensions 384'
CGO_LDFLAGS="-L$(pwd)/../../cozodb-goja/.deps/cozo" go run -tags cozo_cgo ./cmd/cozo-tui --runtime-engine sqlite --runtime-db /tmp/cozo-extraction-tui-seed-real3.db --config-file /home/manuel/.pinocchio/config.yaml --profile-file /home/manuel/.config/pinocchio/profiles.yaml --profile default --embeddings-dimensions 384
git commit -m "cozo-tui: add precedence/decode tests and glazed run docs"
```

## Quick reference

### Immediate implementation guidance

1. Build Glazed commands for `tui` and `plugin-run` first.
2. Add cozo-specific middleware bootstrap chain with explicit precedence tests.
3. Build embeddings provider from `values.Values` and inject into host options.
4. Remove runtime `os.Getenv` access from non-test code.

### Precedence assertion template

```go
// expected: flags > env > profiles > config > defaults
if got := resolve("--field flag", "ENV=env", "profile=profile", "config=file", "default=base"); got != "flag" {
    t.Fatalf("precedence mismatch: %q", got)
}
```

## Usage examples

### Developer validation runbook for CO-11 implementation phase

```bash
# 1) verify parsed layers
cozo tui --config-file ./config.yaml --print-parsed-fields

# 2) validate precedence edge cases
COZO_TUI_EMBEDDINGS_ENGINE=env-engine cozo tui --config-file ./config.yaml --embeddings-engine flag-engine --print-parsed-fields

# 3) run non-interactive seed smoke using typed settings path
cozo tui --engine sqlite --db /tmp/cozo-glazed.db --seed --seed-only --embeddings-type openai --embeddings-engine text-embedding-3-small --embeddings-dimensions 384
```

## What was tricky

1. Geppetto documentation and Geppetto tests imply different profile-vs-config expectations unless you inspect code and comments carefully.
2. Reverse middleware execution order can create subtle misreads when only scanning slice construction.

## What to watch during implementation

1. Avoid fallback helper paths that silently call `os.Getenv`; they can mask incomplete migration.
2. Keep all settings decode at command boundary and pass typed options explicitly into constructors.
3. Add parse provenance assertions in tests, not just final value assertions.

## Related

1. [01-glazed-first-configuration-architecture-and-implementation-plan.md](../design-doc/01-glazed-first-configuration-architecture-and-implementation-plan.md)
2. [tasks.md](../tasks.md)
3. [changelog.md](../changelog.md)
