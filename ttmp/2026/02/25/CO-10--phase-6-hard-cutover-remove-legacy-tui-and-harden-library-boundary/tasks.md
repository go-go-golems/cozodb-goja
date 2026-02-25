# Tasks

## Workstream A: Pre-Cutover Readiness Gate

- [ ] Verify CO-07 outputs are complete and stable
- [ ] Verify CO-08 outputs are complete and stable
- [ ] Verify CO-09 outputs are complete and stable
- [ ] Run relocated app smoke test (`F1-F9`)
- [ ] Freeze legacy TUI path (no additional feature changes)

## Workstream B: Remove Legacy Command Path

- [ ] Delete `cozodb-goja/cmd/cozo-tui/main.go`
- [ ] Confirm no references to deleted command in same module
- [ ] Update any script invoking deleted command path

## Workstream C: Remove Legacy Internal TUI Tree

- [ ] Delete `cozodb-goja/internal/tui/app/model.go`
- [ ] Delete `cozodb-goja/internal/tui/screens/dashboard/model.go`
- [ ] Delete `cozodb-goja/internal/tui/screens/people/model.go`
- [ ] Delete `cozodb-goja/internal/tui/screens/relationships/model.go`
- [ ] Delete `cozodb-goja/internal/tui/screens/evolution/model.go`
- [ ] Delete `cozodb-goja/internal/tui/screens/network/model.go`
- [ ] Delete `cozodb-goja/internal/tui/screens/timeline/model.go`
- [ ] Delete `cozodb-goja/internal/tui/screens/query/model.go`
- [ ] Delete `cozodb-goja/internal/tui/seeddata/seed.go`
- [ ] Confirm internal/tui directory removed entirely

## Workstream D: Reference and Documentation Cutover

- [ ] Update root/project README command examples to relocated path
- [ ] Update CO-05 docs where legacy command appears
- [ ] Update CO-06 docs where legacy command appears
- [ ] Update helper scripts to relocated command path
- [ ] Update any onboarding notes referencing old path

## Workstream E: Build/Test Verification

- [ ] Run `go test ./... -count=1` in `cozodb-goja`
- [ ] Run `go test ./... -count=1` in relocated extraction module
- [ ] Run relocated app command smoke after legacy removal
- [ ] Verify no compile errors from removed imports

## Workstream F: Boundary Audit and Guardrails

- [ ] Run `rg` for `cozodb-goja/internal/tui` references across workspace
- [ ] Run `rg` for `cozodb-goja/cmd/cozo-tui` references across workspace
- [ ] Ensure both searches return no active code references
- [ ] Add or update CI guard checks for reintroduced legacy paths
- [ ] Document boundary rules in changelog/design references

## Workstream G: Ticket and Documentation Hygiene

- [ ] Update CO-10 changelog with hard-cutover actions
- [ ] Relate removed/updated files with `docmgr doc relate`
- [ ] Run `docmgr doctor --ticket CO-10 --stale-after 30`
- [ ] Mark CO-10 tasks complete
- [ ] Add final cutover summary note to CO-06 ticket
