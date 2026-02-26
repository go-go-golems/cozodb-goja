---
Title: Glazed-first configuration architecture and implementation plan
Ticket: CO-11
Status: active
Topics:
    - cozo
    - tui
    - glazed
    - geppetto
    - embeddings
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/cmd/cozo-plugin-run/main.go
      Note: |-
        Second CLI entrypoint still on standard flag package
        Flag-based plugin-run command targeted for Glazed cutover
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/cmd/cozo-tui/main.go
      Note: |-
        Current flag parsing and env mutation bridge
        Flag parsing and env bridge targeted for replacement
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/geppettohost/embedder.go
      Note: |-
        Direct os.Getenv and file-path config loading for embeddings
        Current env/config/profile loading implementation
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/geppettohost/host.go
      Note: |-
        Lazy provider construction through env-backed helper
        Lazy env-backed provider creation path
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/screens/vsearch/commands.go
      Note: |-
        Runtime env toggle for script root
        F9 script-root env dependency pending replacement
    - Path: ../../../../../../../2026-02-18--cozodb-extraction/cozo-extraction-tui/internal/tui/screens/vsearch/model.go
      Note: |-
        Runtime env toggle for auto-migration
        F9 auto-migrate env toggle pending replacement
    - Path: ../../../../../../../geppetto/cmd/examples/simple-streaming-inference/main.go
      Note: |-
        Canonical BuildCobraCommand + geppetto middlewares wiring
        Canonical BuildCobraCommand plus custom middlewares wiring
    - Path: ../../../../../../../geppetto/pkg/embeddings/config/settings.go
      Note: |-
        Embeddings config schema with glazed tags
        Embeddings section and glazed tags
    - Path: ../../../../../../../geppetto/pkg/embeddings/settings_factory.go
      Note: |-
        Decoding parsed values into embeddings provider factory
        Parsed-values decode path for embedding provider
    - Path: ../../../../../../../geppetto/pkg/sections/profile_registry_source.go
      Note: |-
        Registry-backed profile middleware behavior and error model
        Registry-backed profile loading behavior
    - Path: ../../../../../../../geppetto/pkg/sections/profile_registry_source_test.go
      Note: |-
        Explicit precedence and profile-order tests
        Precedence assertions for profile/config/env/flags
    - Path: ../../../../../../../geppetto/pkg/sections/sections.go
      Note: |-
        Geppetto middleware bootstrap and precedence implementation
        Reference middleware bootstrap and precedence chain
ExternalSources: []
Summary: Evidence-backed migration design to remove direct flag and os.Getenv usage in cozo-extraction-tui and replace it with Glazed sections, middleware loading, and typed decoded settings structs.
LastUpdated: 2026-02-25T15:40:00-05:00
WhatFor: Guide implementation of Glazed-first config cutover for cozo-extraction-tui binaries and embedding runtime wiring.
WhenToUse: Use before implementing CO-11 to align on precedence, schema boundaries, migration phases, and test strategy.
---


# Glazed-first configuration architecture and implementation plan

## Executive summary

CO-11 addresses a structural configuration problem in `cozo-extraction-tui`: runtime behavior currently depends on a mix of direct `flag` parsing, ad-hoc `os.Setenv`, and deep `os.Getenv` reads across internal packages. This creates hidden global state, weak precedence guarantees, and brittle testing boundaries.

The target is a Glazed-first architecture where:

1. All CLI flags are declared via Glazed field definitions.
2. All config sources are merged by middleware (defaults, profiles, config files, env, args, flags).
3. Command handlers decode typed settings structs from parsed sections.
4. Internal packages receive explicit settings or constructed providers, not process environment lookups.

The migration can be done incrementally without changing core extraction/vector features. The critical engineering choice is whether CO-11 mirrors Geppetto’s current precedence behavior (`flags > env > profiles > config > defaults` in effective value order) or enforces the documented Glazed recommendation (`flags > env > config > profiles > defaults`). This document recommends choosing one explicitly and locking it with tests at the cozo app boundary.

## Problem statement and scope

### Problem

The current implementation has three coupling points that conflict with a deterministic, inspectable configuration model:

1. CLI parsing is done with `flag` in both binaries instead of Glazed command sections.
2. `cozo-tui` translates CLI flags into environment variables (`os.Setenv`), introducing mutable process-global state.
3. runtime packages (`internal/geppettohost`, F9 model/commands) read environment directly (`os.Getenv`), bypassing typed command settings.

This makes behavior harder to reason about and violates the requested direction: no direct `os.Getenv`, all flags/inputs through Glazed sections and struct decode.

### In-scope

1. `cmd/cozo-tui` and `cmd/cozo-plugin-run` command architecture.
2. embedding provider configuration path used by seed/F8/F9 flows.
3. profile/config/env/flag precedence behavior.
4. internal settings plumbing required to eliminate runtime `os.Getenv` calls.

### Out-of-scope

1. Cozo query semantics and schema design.
2. plugin JS runtime behavior unrelated to config loading.
3. changing Geppetto inference/embedding provider internals beyond integration points.

## Current-state architecture (evidence-backed)

### 1) `cozo-tui` uses standard `flag` parsing and process env mutation

`cmd/cozo-tui/main.go` defines all runtime options with `flag.String/Bool/Int` and parses with `flag.Parse()` (`main.go:20-36`). It then calls `applyEmbeddingCLIOverrides(...)`, which writes overrides into process environment keys like `COZO_TUI_EMBEDDINGS_TYPE`, `COZO_TUI_OPENAI_API_KEY`, and `COZO_TUI_PINOCCHIO_PROFILE_FILE` via `os.Setenv` (`main.go:106-161`).

Observed implication: command options are not held as an immutable typed runtime config object; they are projected into global process state for later lazy reads.

### 2) embedding provider defaults and overrides are env/file driven inside runtime package

`internal/geppettohost/embedder.go` builds defaults and then mutates them through:

1. optional direct file loading (`loadPinocchioEmbeddingsConfig`) based on env or home-based search paths (`embedder.go:90-145`), and
2. explicit env overrides (`applyEmbeddingEnvOverrides`) reading `COZO_TUI_*`, `OPENAI_API_KEY`, and `OLLAMA_BASE_URL` (`embedder.go:190-222`).

`host.ensureEmbedder()` lazily calls `defaultEmbeddingProviderFromEnv()` when no explicit provider is injected (`host.go:114-127`).

Observed implication: internal behavior depends on ambient process state at call time and cannot be fully inferred from command input alone.

### 3) additional env reads exist outside embedder path

F9 model and command paths still read env directly:

1. `COZO_TUI_AUTO_MIGRATE_VECTORS` in `vsearch/model.go:73` and parser helper `envEnabled` (`model.go:327-337`).
2. `COZO_TUI_SCRIPT_ROOT` in `vsearch/commands.go:110-115`.

Observed implication: even if embedding configuration were decoded from Glazed once, F9 behavior still depends on runtime env.

### 4) `cozo-plugin-run` is also `flag`-based

`cmd/cozo-plugin-run/main.go` defines and parses its own flag set (`main.go:19-30`), including script path, transcript, timeout, profile, and engine options JSON.

Observed implication: a Glazed-first cutover should include both binaries to avoid mixed conventions and duplicated parsing logic.

### 5) dependency baseline already includes Glazed/Cobra

`cozo-extraction-tui/go.mod` already has `github.com/go-go-golems/geppetto` as direct dependency and `github.com/go-go-golems/glazed` / `github.com/spf13/cobra` as transitive dependencies (`go.mod:12`, `go.mod:42`, `go.mod:77`).

Observed implication: no ecosystem mismatch blocks immediate adoption.

## Geppetto middleware analysis

This section maps exactly how Geppetto currently solves config/profile/env/flag loading using Glazed middlewares, since CO-11 wants equivalent behavior without ad-hoc `os.Getenv`.

### 1) Section construction with typed defaults

`geppetto/pkg/sections/CreateGeppettoSections` creates section objects for chat/client/provider/embeddings/inference and initializes section defaults from `StepSettings` (`sections.go:34-121`).

This is the key pattern CO-11 should copy:

1. define sections,
2. initialize defaults from typed structs,
3. decode back into typed structs after parse.

### 2) Bootstrap parse pattern for profile-selection circularity

`GetCobraCommandGeppettoMiddlewares` contains a deliberate bootstrap stage:

1. parse command settings from Cobra + env + defaults (`sections.go:169-193`),
2. resolve config files from app defaults plus `--config-file` (`sections.go:195-203`),
3. parse profile settings from Cobra + env + config + defaults (`sections.go:205-234`),
4. instantiate profile middleware using resolved profile/profile-file (`sections.go:273-297`).

This aligns with `glaze help implementing-profile-middleware` and `glaze help profiles`, which call out bootstrap parsing as the fix for profile selection circularity.

### 3) Registry-backed profile middleware behavior

`GatherFlagsFromProfileRegistry` in `profile_registry_source.go` resolves profile overlays via Geppetto profile registry and maps them to section fields (`profile_registry_source.go:17-95`).

Error behavior is explicit and strict:

1. non-default missing profile file errors (`profile_registry_source.go:40-47`),
2. missing non-default profile in existing file errors (`profile_registry_source.go:77-80`),
3. default profile missing can no-op for default fallback case.

This provides a robust model CO-11 can reuse instead of bespoke YAML parsing in app code.

### 4) Effective precedence in current Geppetto implementation

Geppetto middleware slice is built in an order that relies on Glazed reverse execution semantics. The code comments explicitly explain this in `sections.go:301-304`.

In the current chain, tested behavior is:

1. profiles override config (`profile_registry_source_test.go:183-186`),
2. env overrides profile (`profile_registry_source_test.go:187-189`),
3. flags override env/profile/config (`profile_registry_source_test.go:190-192`).

So effective precedence today is:

`flags > env > profiles > config > defaults`

This differs from the general documentation phrase used in some Glazed help pages (`flags > env > config > profiles > defaults`). CO-11 must choose the intended behavior explicitly and assert it with local tests.

### 5) Struct decode path for embeddings provider

Geppetto’s embeddings config struct uses `glazed` tags (`embeddings/config/settings.go:11-29`) and `NewSettingsFactoryFromParsedValues` decodes section data from `values.Values` into typed config (`settings_factory.go:170-191`).

This is the precise integration seam CO-11 should use to build providers without reading env directly in runtime packages.

### 6) Canonical command wiring example

`geppetto/cmd/examples/simple-streaming-inference/main.go` demonstrates:

1. command settings struct with `glazed` tags (`main.go:50-59`),
2. field definitions with `fields.New(...)` (`main.go:68-114`),
3. section attachment (`main.go:115-117`),
4. `DecodeSectionInto(values.DefaultSlug, s)` in run path (`main.go:125-132`),
5. Cobra command build with custom middlewares func (`main.go:256-259`).

That is the operational template for CO-11 command cutover.

## Gap analysis

### Gap A: configuration boundary is process-global, not command-scoped

Current: command options are converted to env and read later by unrelated components.

Target: parse once into immutable settings structs and pass dependencies explicitly.

### Gap B: CLI frameworks are inconsistent (`flag` vs Glazed/Cobra)

Current: no uniform schema/section tooling for output, debug parse logs, profile settings sections.

Target: all commands are Glazed commands, enabling `--print-parsed-fields`, profile settings, and config overlays in a standard way.

### Gap C: profile/config/env logic is duplicated manually in app package

Current: `embedder.go` replicates loading/merging logic.

Target: reuse Geppetto middleware + parsed value decode pipeline as source of truth.

### Gap D: precedence policy is not explicit in cozo app

Current: behavior emerges from env read order and manual merges.

Target: precedence declared and tested in one middleware function.

## Proposed architecture

## 1) Command topology

Introduce Glazed/Cobra root with explicit subcommands:

1. `cozo tui`
2. `cozo plugin-run`

Compatibility option:

1. keep `cmd/cozo-tui` and `cmd/cozo-plugin-run` as thin wrappers invoking shared command constructors, or
2. hard cutover to one binary if acceptable.

Given the user preference for hard cutovers in recent tickets, this document assumes hard cutover is acceptable unless release constraints demand wrappers.

## 2) Section model

Use three section classes:

1. App runtime sections (cozo DB, seed behavior, TUI behavior, plugin-run behavior).
2. Geppetto sections from `CreateGeppettoSections()` for embeddings/provider fields.
3. Standard Glazed command/profile/output sections.

Suggested custom sections for cozo app:

1. `cozo-runtime`
2. `cozo-tui`
3. `cozo-plugin-run`

Embed fields currently represented by env variables as typed Glazed fields, for example:

1. `cozo-tui.auto-migrate-vectors` (replaces `COZO_TUI_AUTO_MIGRATE_VECTORS`)
2. `cozo-tui.script-root` (replaces `COZO_TUI_SCRIPT_ROOT` path dependence)
3. `cozo-runtime.seed-only`, `cozo-runtime.seed`
4. `cozo-runtime.engine`, `cozo-runtime.db`

Geppetto/embeddings fields remain in their native sections (`embeddings`, `openai-chat`, etc), decoded by existing factories.

## 3) Typed settings structs

Define strongly typed structs that mirror section slugs and fields:

```go
type CozoRuntimeSettings struct {
    Engine   string `glazed:"engine"`
    DBPath   string `glazed:"db"`
    Seed     bool   `glazed:"seed"`
    SeedOnly bool   `glazed:"seed-only"`
}

type CozoTUISettings struct {
    AutoMigrateVectors bool   `glazed:"auto-migrate-vectors"`
    ScriptRoot         string `glazed:"script-root"`
}

type CozoPluginRunSettings struct {
    ScriptPath       string `glazed:"script"`
    TranscriptPath   string `glazed:"transcript"`
    ScriptRoot       string `glazed:"script-root"`
    ScriptDB         string `glazed:"script-db"`
    Prompt           string `glazed:"prompt"`
    Profile          string `glazed:"profile"`
    TimeoutMs        int    `glazed:"timeout-ms"`
    EngineOptionsRaw string `glazed:"engine-options-json"`
    IncludeMetadata  bool   `glazed:"include-metadata"`
    Pretty           bool   `glazed:"pretty"`
}
```

And an aggregate runtime container used internally:

```go
type AppResolvedSettings struct {
    Runtime  CozoRuntimeSettings
    TUI      CozoTUISettings
    Plugin   CozoPluginRunSettings
    Parsed   *values.Values
}
```

## 4) Middleware chain design

### 4.1 Required behavior

No direct `os.Getenv` in application runtime code. All env/config/profile resolution must happen in middleware parse stage.

### 4.2 Bootstrap + main chain pattern

Follow Geppetto’s bootstrap pattern, but parameterize app identity and env prefix for this app (for example `COZO_TUI`):

```go
func GetCobraCommandCozoMiddlewares(parsedCommand *values.Values, cmd *cobra.Command, args []string) ([]sources.Middleware, error) {
    // 1) bootstrap command settings (resolve config file list)
    // 2) bootstrap profile settings (resolve profile/profile-file)
    // 3) instantiate profile middleware using resolved values
    // 4) return chain in reverse-precedence slice order
}
```

### 4.3 Explicit precedence decision

CO-11 must lock one model:

Option 1 (mirror current Geppetto behavior):
`flags > env > profiles > config > defaults`

Option 2 (follow common docs wording):
`flags > env > config > profiles > defaults`

Recommendation: start with Option 1 for behavioral continuity with Geppetto examples/tests, then revisit only if there is a product-level requirement that config files must override profiles.

## 5) Embeddings provider construction without env reads

Replace `defaultEmbeddingProviderFromEnv()` usage path with parsed-values-driven provider setup.

### 5.1 Current

`Host.ensureEmbedder()` lazy-calls env-backed helper (`host.go:122-127`).

### 5.2 Target

Construct provider before host construction in command run path:

```go
providerFactory, err := embeddings.NewSettingsFactoryFromParsedValues(parsedValues)
if err != nil { return err }
provider, err := providerFactory.NewProvider()
if err != nil { return err }

host, err := geppettohost.New(geppettohost.Options{
    Name:       "cozo-f9-vector-search",
    ScriptRoot: settings.TUI.ScriptRoot,
    Embedding:  provider,
})
```

Then remove env fallback from host internals (or keep strictly as deprecated compatibility path gated behind tests during migration window).

## 6) Internal dependency injection adjustments

To eliminate non-command env reads:

1. `vsearch.New(...)` should accept typed runtime options (at least `AutoMigrate` and `ScriptRoot`).
2. `defaultEmbedText` should use injected host factory or injected provider, not `os.Getenv`.
3. seed path should receive a provider/factory or parsed values context from command layer.
4. `geppettohost` should expose provider-from-parsed-values helper instead of provider-from-env helper.

Suggested direction:

```go
type EmbeddingRuntime struct {
    Provider embeddings.Provider
    ScriptRoot string
}

type HostFactory interface {
    NewEmbeddingHost(name string, scriptRoot string, provider embeddings.Provider) (*geppettohost.Host, error)
}
```

## 7) Config file and profile shape

Use standard section-formatted YAML as produced/consumed by Glazed:

```yaml
cozo-runtime:
  engine: sqlite
  db: /tmp/cozo.db
  seed: true
  seed-only: false

cozo-tui:
  auto-migrate-vectors: true
  script-root: /home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui

embeddings:
  embeddings-type: openai
  embeddings-engine: text-embedding-3-small
  embeddings-dimensions: 384

openai-chat:
  openai-api-key: ${OPENAI_API_KEY}
```

Profile file overlays should remain aligned with Geppetto profile registry schema and profile settings section.

## 8) CLI authoring pattern for CO-11

Based on `glazed-command-authoring` and Geppetto examples, each command should follow this contract:

1. command struct embeds `*cmds.CommandDescription`.
2. settings struct uses `glazed` tags.
3. constructor builds description with `cmds.WithFlags` and `cmds.WithSections`.
4. run method decodes structs via `DecodeSectionInto`.
5. Cobra is built with parser config including custom middlewares func.

## Phased implementation plan

### Phase 1: command foundation and schema definition

1. Introduce new command package layout under `internal/cmd/`.
2. Define shared sections and settings structs for runtime, tui, plugin-run.
3. Build root command + subcommands with `cli.BuildCobraCommandFromCommand`.
4. Add command settings/profile settings sections.

Deliverable: commands run with defaults and `--print-parsed-fields` works.

### Phase 2: cozo middleware chain

1. Add `internal/config/middlewares.go` implementing bootstrap + main chain.
2. Parameterize env prefix (`COZO_TUI`) and config app path.
3. Decide and encode precedence policy; document it in code comments.
4. Add precedence matrix tests similar to Geppetto test style.

Deliverable: deterministic precedence tests passing.

### Phase 3: embeddings plumbing cutover

1. Build embeddings provider from parsed values in command layer.
2. Inject provider via `geppettohost.Options.Embedding` everywhere.
3. Remove or deprecate `defaultEmbeddingProviderFromEnv` path.
4. Move pinocchio/config/profile file selection into profile settings + middleware only.

Deliverable: seed/F8/F9 flows work with no runtime env reads for embedding config.

### Phase 4: F9 and seed runtime cleanup

1. Remove `envEnabled` and script-root env reads from F9 model/commands.
2. Inject config through constructor options.
3. Update app model wiring to pass settings into F9 screen.
4. Ensure seed path uses injected provider and keeps dimension validation.

Deliverable: no `os.Getenv` usage in vsearch and seed runtime paths.

### Phase 5: plugin-run command cutover

1. Replace `cmd/cozo-plugin-run` flag parser with Glazed command implementation.
2. Decode plugin-run settings struct.
3. Keep behavior parity for required fields and output formatting.
4. Add command-level tests for required args and JSON parse errors.

Deliverable: both binaries use Glazed command framework.

### Phase 6: hard cutover and cleanup

1. Delete deprecated env bridge code and legacy helpers.
2. Remove now-unused env variable docs and tests.
3. Update Makefile/README with new command invocation examples.
4. Add migration notes for operators moving from env-heavy workflows.

Deliverable: repository has a single configuration model.

## Detailed task list (implementation-ready)

1. Add `internal/cmd/root.go` with logging and help setup.
2. Add `internal/cmd/tui_command.go` with command description and settings decode.
3. Add `internal/cmd/plugin_run_command.go` with command description and settings decode.
4. Add `internal/config/sections.go` for app sections.
5. Add `internal/config/middlewares.go` for bootstrap and chain.
6. Add `internal/config/middlewares_test.go` precedence matrix tests.
7. Add `internal/config/decoded_settings.go` conversion helpers.
8. Add `internal/geppettohost/provider_from_values.go` using `embeddings.NewSettingsFactoryFromParsedValues`.
9. Refactor `internal/geppettohost/host.go` to require injected provider or explicit provider factory input.
10. Remove env-based loading from `internal/geppettohost/embedder.go` (or isolate legacy fallback behind temporary compatibility flag for one phase).
11. Refactor F9 constructor to accept explicit options and remove env reads.
12. Refactor seed path to accept provider/provider-factory input from command layer.
13. Add `cozo-tui` end-to-end parse test using config + env + flag overrides.
14. Add live smoke test command path using parsed settings only.
15. Update docs and examples with section-formatted config/profile files.

## Testing and validation strategy

### 1) Unit tests

1. section defaults and decode tests for each settings struct.
2. middleware precedence tests across defaults/config/profiles/env/flags.
3. profile bootstrap tests for selection from config and env.
4. provider factory tests from parsed values (openai/ollama failure modes).

### 2) Integration tests

1. seed-only flow with parsed embeddings settings and real provider (env for credentials still allowed through Glazed env middleware, but runtime code should not call `os.Getenv`).
2. F9 query flow with preflight + embed + query path.
3. plugin-run flow with JSON engine options and file inputs.

### 3) Regression matrix

Validate equivalent outcomes for prior env workflows by expressing them via:

1. config file only,
2. profile only,
3. env only,
4. flags only,
5. mixed overlays.

### 4) Diagnostics

Ensure `--print-parsed-fields` is available and documented as first-line debugging tool for configuration disputes.

## Risks, tradeoffs, and alternatives

### Risk 1: precedence mismatch with operator expectations

Because Geppetto implementation and some docs differ on profile-vs-config ordering, ambiguous assumptions may break rollout.

Mitigation:

1. freeze precedence in tests,
2. document explicitly,
3. print effective sources during debug runs.

### Risk 2: over-coupling to Pinocchio naming

Geppetto middleware helper currently hardcodes `PINOCCHIO` prefix and `pinocchio` config path (`sections.go:181`, `sections.go:197`, `sections.go:276`).

Mitigation:

1. implement COZO-specific middleware function locally first,
2. upstream parameterized helper later to reduce drift.

### Risk 3: broad refactor touches multiple runtime paths

This cutover spans command entrypoints, host internals, seed, and F9 screen code.

Mitigation:

1. phase migration,
2. preserve behavior with golden tests,
3. cut legacy bridge only after parity validation.

### Alternative A: keep current env bridge and only swap flag parsing

Pros:

1. smaller short-term change.

Cons:

1. does not satisfy goal (still env-driven internals),
2. keeps hidden mutable global state.

Rejected for CO-11 objective.

### Alternative B: directly call Geppetto `GetCobraCommandGeppettoMiddlewares` from app

Pros:

1. less code duplication.

Cons:

1. hardcoded `PINOCCHIO` app identity and path,
2. may pull unrelated layer assumptions.

Recommended only if Geppetto exposes a parameterized variant.

## Open questions

1. Should CO-11 adopt Geppetto’s tested precedence (`profiles > config`) or the generic docs ordering (`config > profiles`)?
2. Do we keep temporary compatibility aliases for old `COZO_TUI_*` env names, and if yes for how long?
3. Should `cozo-tui` and `cozo-plugin-run` remain separate binaries or become subcommands of a single binary during hard cutover?
4. Do we upstream a parameterized middleware helper into Geppetto in the same sprint or defer to a follow-up ticket?

## Recommended decision set for implementation kickoff

1. Adopt Geppetto precedence for initial cutover (`flags > env > profiles > config > defaults`) to minimize behavior drift.
2. Implement COZO-specific middleware function locally with explicit env prefix and app config path.
3. Remove runtime `os.Getenv` usage from non-test code in CO-11 scope.
4. Keep no backwards compatibility shims unless tests reveal hard operational breakage.

## References

1. `cozo-extraction-tui/cmd/cozo-tui/main.go` (`flag` usage and `os.Setenv` bridge).
2. `cozo-extraction-tui/cmd/cozo-plugin-run/main.go` (`flag`-based plugin runner).
3. `cozo-extraction-tui/internal/geppettohost/embedder.go` (env and file loading logic).
4. `cozo-extraction-tui/internal/geppettohost/host.go` (lazy env-backed provider initialization).
5. `cozo-extraction-tui/internal/tui/screens/vsearch/model.go` (env toggle for auto-migration).
6. `cozo-extraction-tui/internal/tui/screens/vsearch/commands.go` (script root env read).
7. `geppetto/pkg/sections/sections.go` (bootstrap + middleware chain and comments).
8. `geppetto/pkg/sections/profile_registry_source.go` (profile registry middleware implementation).
9. `geppetto/pkg/sections/profile_registry_source_test.go` (ordering assertions and schema assembly tests).
10. `geppetto/pkg/embeddings/config/settings.go` (glazed tags for embeddings settings).
11. `geppetto/pkg/embeddings/settings_factory.go` (parsed-values decode to provider factory).
12. `geppetto/cmd/examples/simple-streaming-inference/main.go` (canonical command wiring).
13. `glaze help middlewares-guide` (reverse middleware execution semantics).
14. `glaze help profiles` (profile-settings and bootstrap guidance).
15. `glaze help implementing-profile-middleware` (bootstrap parse pattern and precedence guidance).
16. `glaze help config-files-quickstart` (overlay and parse provenance patterns).
