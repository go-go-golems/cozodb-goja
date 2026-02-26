# Tasks

## Workstream A: Pre-Cutover Readiness Gate

- [x] Verify CO-07 outputs are complete and stable
- [x] Verify CO-08 outputs are complete and stable
- [x] Verify CO-09 outputs are complete and stable
- [x] Run relocated app smoke test (`F1-F9`)
- [x] Freeze legacy TUI path (no additional feature changes)

## Workstream B: Remove Legacy Command Path

- [x] Delete `cozodb-goja/cmd/cozo-tui/main.go`
- [x] Confirm no references to deleted command in same module
- [x] Update any script invoking deleted command path

## Workstream C: Remove Legacy Internal TUI Tree

- [x] Delete `cozodb-goja/internal/tui/app/model.go`
- [x] Delete `cozodb-goja/internal/tui/screens/dashboard/model.go`
- [x] Delete `cozodb-goja/internal/tui/screens/people/model.go`
- [x] Delete `cozodb-goja/internal/tui/screens/relationships/model.go`
- [x] Delete `cozodb-goja/internal/tui/screens/evolution/model.go`
- [x] Delete `cozodb-goja/internal/tui/screens/network/model.go`
- [x] Delete `cozodb-goja/internal/tui/screens/timeline/model.go`
- [x] Delete `cozodb-goja/internal/tui/screens/query/model.go`
- [x] Delete `cozodb-goja/internal/tui/seeddata/seed.go`
- [x] Confirm internal/tui directory removed entirely

## Workstream D: Reference and Documentation Cutover

- [x] Update root/project README command examples to relocated path
- [x] Update CO-05 docs where legacy command appears
- [x] Update CO-06 docs where legacy command appears
- [x] Update helper scripts to relocated command path
- [x] Update any onboarding notes referencing old path

## Workstream E: Build/Test Verification

- [x] Run `go test ./... -count=1` in `cozodb-goja`
- [x] Run `go test ./... -count=1` in relocated extraction module
- [x] Run relocated app command smoke after legacy removal
- [x] Verify no compile errors from removed imports

## Workstream F: Boundary Audit and Guardrails

- [x] Run `rg` for `cozodb-goja/internal/tui` references across workspace
- [x] Run `rg` for `cozodb-goja/cmd/cozo-tui` references across workspace
- [x] Ensure both searches return no active code references
- [x] Add or update CI guard checks for reintroduced legacy paths
- [x] Document boundary rules in changelog/design references

## Workstream G: Ticket and Documentation Hygiene

- [x] Update CO-10 changelog with hard-cutover actions
- [x] Relate removed/updated files with `docmgr doc relate`
- [x] Run `docmgr doctor --ticket CO-10 --stale-after 30`
- [x] Mark CO-10 tasks complete
- [x] Add final cutover summary note to CO-06 ticket

## Workstream H: JS Plugin Contract Hard Cutover (Geppetto -> Cozo Runtime Ownership)

- [x] Add local `cozo/plugins` JS native module in relocated extraction runtime
- [x] Copy JS `geppetto` module runtime ownership into `cozo-extraction-tui/internal/jsmodules/geppetto`
- [x] Stop importing external `geppetto/pkg/js/modules/geppetto` from runtime wiring
- [x] Migrate active extractor scripts from `require("geppetto/plugins")` to `require("cozo/plugins")`
- [x] Migrate runtime/tests from `geppetto/plugins` assertions to `cozo/plugins`
- [x] Update onboarding/cookbook docs to reflect `cozo/plugins` module path
- [x] Remove `defineExtractorPlugin`/`wrapExtractorRun` usage from `cozodb-goja` ticket script artifacts
- [x] Validate relocation with `go test ./... -count=1` in `cozo-extraction-tui`
- [x] Validate plugin smoke with `cozo-plugin-run` fixture script against stdin transcript
