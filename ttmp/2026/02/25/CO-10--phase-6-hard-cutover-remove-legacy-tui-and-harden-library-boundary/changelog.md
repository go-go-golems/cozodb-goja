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
