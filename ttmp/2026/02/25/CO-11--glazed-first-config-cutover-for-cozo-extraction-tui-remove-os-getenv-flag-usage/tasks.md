# Tasks

## Research and planning

- [x] Create CO-11 ticket workspace and baseline docs.
- [x] Analyze current `cozo-tui` command parsing and env bridge behavior.
- [x] Analyze current `cozo-plugin-run` command parsing.
- [x] Analyze env-dependent runtime paths in `geppettohost`, seed, and F9.
- [x] Analyze Geppetto middleware bootstrap/profile precedence implementation.
- [x] Analyze Geppetto embeddings parsed-values factory path.
- [x] Produce long-form design and implementation plan document.
- [x] Record chronological investigation diary with command evidence.

## Phase 1: Glazed command foundation

- [ ] Create shared Cobra root and command registration package.
- [ ] Implement Glazed command for TUI execution.
- [ ] Implement Glazed command for plugin-run execution.
- [ ] Add command settings and profile settings sections to both commands.
- [ ] Preserve current behavior parity for required args and defaults.

## Phase 2: Cozo middleware chain

- [ ] Implement cozo-specific middleware builder with bootstrap parse.
- [ ] Parameterize app env prefix and app config path.
- [ ] Support profile selection from flags/env/config through profile settings section.
- [ ] Encode explicit precedence policy in code and comments.
- [ ] Add precedence matrix tests (defaults/config/profiles/env/flags).

## Phase 3: Embeddings provider cutover

- [ ] Build embeddings provider from parsed values in command layer.
- [ ] Inject provider into `geppettohost.Options.Embedding` for all runtime paths.
- [ ] Remove runtime env fallback from provider construction path.
- [ ] Replace manual Pinocchio file parsing in app runtime with middleware/profile-driven values.

## Phase 4: Runtime env removal

- [ ] Remove `COZO_TUI_AUTO_MIGRATE_VECTORS` env reads from F9 model.
- [ ] Remove `COZO_TUI_SCRIPT_ROOT` env reads from F9 commands.
- [ ] Pass explicit typed options into F9 constructor and host setup.
- [ ] Pass explicit typed options/provider into seed embedding flow.

## Phase 5: Plugin-run hard cutover

- [ ] Remove standard `flag` usage from plugin-run command.
- [ ] Decode plugin-run settings struct from Glazed parsed values.
- [ ] Keep engine options JSON parse behavior parity.
- [ ] Add tests for required-field and parse-error behavior.

## Phase 6: Cleanup and rollout

- [ ] Remove obsolete env-bridge helpers and dead code.
- [ ] Update Makefile and README invocation examples to new command surface.
- [ ] Add migration guidance for operators.
- [ ] Run full test suite and live seed smoke using parsed-values path.
- [ ] Perform manual interactive TUI smoke in a real TTY session.

## Documentation and delivery

- [x] Relate key evidence files to design and diary docs.
- [ ] Keep changelog synchronized with phase milestones.
- [x] Run `docmgr doctor --ticket CO-11 --stale-after 30` with clean result.
- [x] Upload CO-11 bundle to reMarkable and verify remote listing.
