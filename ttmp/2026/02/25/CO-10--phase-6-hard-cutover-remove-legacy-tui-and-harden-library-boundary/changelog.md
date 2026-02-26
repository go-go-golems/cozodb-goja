# Changelog

## 2026-02-25

- Created CO-10 phase ticket for hard cutover and boundary cleanup.
- Added detailed implementation design document for irreversible legacy-path removal and boundary hardening.
- Added granular cleanup and verification checklist tasks.
- Added CO-10 diary and related-file links; queued implementation pending CO-07 to CO-09 delivery.
- Executed hard-cutover cleanup commit `0518bd4`:
  - removed `cmd/cozo-tui/main.go`, `cmd/cozo-seed/main.go`, and `internal/tui/*` from `cozodb-goja`.
  - ran `go mod tidy` and validated `go test ./... -count=1` in `cozodb-goja`.
- Added boundary guardrail script:
  - `ttmp/2026/02/25/CO-10--phase-6-hard-cutover-remove-legacy-tui-and-harden-library-boundary/scripts/01-guard-no-legacy-tui-paths.sh`
  - wired `make guard-no-legacy-tui-paths`
  - wired `.github/workflows/push.yml` to run guard before unit tests.
- Verified active-code boundary scans:
  - no legacy references in `cozodb-goja` outside ticket docs (`rg ... --glob '!ttmp/**'` checks).
- Validated relocated runtime path:
  - `go test ./... -count=1` passes in `2026-02-18--cozodb-extraction/cozo-extraction-tui`.
  - interactive TUI launch smoke executed in PTY (`go run ./cmd/cozo-tui --runtime-engine mem`), interrupted manually after launch.
- Updated cross-ticket docs:
  - CO-05 index/tasks now reference relocated canonical command/runtime paths.
  - CO-06 changelog includes final cutover summary and boundary note.

## 2026-02-25

CO-10 closed: hard cutover complete, legacy path guardrails active

## 2026-02-26

- Added follow-up hard-cutover tasks for JS plugin contract runtime ownership split.
- In relocated extraction runtime (`cozo-extraction-tui`):
  - copied JS `geppetto` module integration to local `internal/jsmodules/geppetto`,
  - added local `cozo/plugins` native module (`defineExtractorPlugin`, `wrapExtractorRun`),
  - migrated active scripts/tests/docs from `require("geppetto/plugins")` to `require("cozo/plugins")`,
  - rewired runtime registration to local modules and removed stale copied `generate.go`.
- In `cozodb-goja` ticket artifacts:
  - removed `defineExtractorPlugin`/`wrapExtractorRun` usage from COJS-01 probe scripts and CO-05 template script by converting to explicit descriptor exports.
- Validation:
  - `GOWORK=off go test ./... -count=1` passed in `cozo-extraction-tui`.
  - `cozo-plugin-run` fixture smoke succeeded with `cozo/plugins`.
