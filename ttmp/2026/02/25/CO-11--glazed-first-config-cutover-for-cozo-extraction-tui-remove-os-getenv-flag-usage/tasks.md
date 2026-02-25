# Tasks

## Locked decisions (hard cutover)

- [x] Use hard cutover only (no backwards compatibility path).
- [x] Keep simplest precedence for now: `flags > env > profiles > config > defaults`.
- [x] Remove runtime `os.Getenv` usage from non-test code in cutover scope.
- [x] Remove standard `flag` parsing from command entrypoints in cutover scope.

## Research and planning (completed)

- [x] Create CO-11 ticket workspace and baseline docs.
- [x] Analyze current `cozo-tui` command parsing and env bridge behavior.
- [x] Analyze current `cozo-plugin-run` command parsing.
- [x] Analyze env-dependent runtime paths in `geppettohost`, seed, and F9.
- [x] Analyze Geppetto middleware bootstrap/profile precedence implementation.
- [x] Analyze Geppetto embeddings parsed-values factory path.
- [x] Produce long-form design and implementation plan document.
- [x] Record chronological investigation diary with command evidence.

## Workstream A: command framework cutover

- [x] Add a shared Cobra root and command registration package for cozo app binaries.
- [x] Implement `tui` command as Glazed command (no `flag` package).
- [x] Implement `plugin-run` command as Glazed command (no `flag` package).
- [x] Wire Glazed command settings section for config overlays and diagnostics.
- [x] Wire Glazed profile settings section for profile/profile-file selection.
- [x] Ensure command handlers decode typed settings structs via `DecodeSectionInto`.

## Workstream B: app sections and typed settings

- [x] Create app section for runtime DB/seed fields.
- [x] Create app section for TUI runtime behavior (auto-migrate, script-root, etc).
- [x] Create app section for plugin-run behavior and IO fields.
- [x] Define typed settings structs with `glazed` tags for all app fields.
- [ ] Define one aggregate runtime settings struct used by command execution paths.

## Workstream C: middleware bootstrap and precedence

- [x] Implement cozo-specific middleware builder with bootstrap parse for command settings.
- [x] Implement bootstrap parse for profile settings before profile middleware instantiation.
- [x] Resolve config file overlays in one place and reuse for bootstrap + main chain.
- [x] Insert profile middleware with registry-backed loading semantics.
- [x] Enforce precedence behavior: `flags > env > profiles > config > defaults`.
- [x] Add explicit comments in middleware builder describing reverse execution semantics.

## Workstream D: embeddings provider cutover

- [x] Construct embeddings provider from parsed values in command layer.
- [x] Inject provider into `geppettohost.Options.Embedding` in all call paths.
- [x] Remove env-based lazy provider fallback from `geppettohost` runtime path.
- [x] Remove manual Pinocchio YAML merge logic from app runtime package.
- [x] Keep dimension validation and provider error surfacing behavior intact.

## Workstream E: TUI runtime env removal

- [x] Remove `COZO_TUI_AUTO_MIGRATE_VECTORS` runtime read from F9 model.
- [x] Remove `COZO_TUI_SCRIPT_ROOT` runtime read from F9 commands.
- [x] Pass auto-migrate/script-root as typed options at constructor boundary.
- [x] Ensure seed and F9 use injected settings/provider only.

## Workstream F: plugin-run command hard cutover

- [x] Replace current `cmd/cozo-plugin-run` flag parsing with Glazed command implementation.
- [x] Preserve required field checks (`script`, `transcript`) in decoded settings validation.
- [x] Preserve engine options JSON parse semantics and output format behavior.
- [x] Keep metadata wrapping and pretty output toggles under typed settings.

## Workstream G: delete legacy compatibility paths

- [x] Delete `applyEmbeddingCLIOverrides` env-bridge behavior from `cozo-tui` command path.
- [x] Delete legacy env-helper functions used only for configuration loading.
- [ ] Delete obsolete docs/hints referring to `COZO_TUI_*` runtime-only configuration.
- [x] Remove dead tests that exist only to validate env bridge behavior.

## Workstream H: validation and tests

- [ ] Add middleware precedence tests covering defaults/config/profiles/env/flags.
- [ ] Add bootstrap profile selection tests (profile from config/env/flag).
- [ ] Add command decode tests for `tui` and `plugin-run` settings structs.
- [ ] Add provider construction tests from parsed values (success + failure modes).
- [x] Run `go test ./... -count=1` in `cozo-extraction-tui`.
- [ ] Run native-tagged vector path tests with `.deps` (`cozo_cgo`).
- [ ] Run live seed smoke with real credentials through parsed-values path.
- [ ] Run manual interactive TUI smoke in real TTY.

## Workstream I: rollout docs and ticket hygiene

- [ ] Update README and command help examples for Glazed-only invocation.
- [ ] Add a short cutover note documenting that old env bridge is removed.
- [ ] Relate final implementation files to CO-11 docs.
- [ ] Keep changelog synchronized with phase milestones.
- [x] Run `docmgr doctor --ticket CO-11 --stale-after 30` with clean result.
- [x] Upload CO-11 bundle to reMarkable and verify remote listing.
