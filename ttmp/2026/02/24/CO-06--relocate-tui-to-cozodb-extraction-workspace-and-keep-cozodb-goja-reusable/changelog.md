# Changelog

## 2026-02-25

- Created CO-06 relocation ticket and baseline documentation workspace.
- Added design document `design/01-relocation-and-reuse-plan-tui-in-extraction-workspace-cozodb-goja-as-library.md` with detailed architecture analysis, option tradeoffs, reuse matrix, phased implementation plan, file move map, risk controls, and milestone schedule.
- Added implementation diary `reference/01-implementation-diary.md` with command-by-command research log and decision rationale.
- Added relocation diagnostics script `scripts/01-relocation-preflight.sh`.
- Updated `tasks.md` with phased M0-M4 execution backlog and research task checklist.
- Ran `docmgr doctor --ticket CO-06 --stale-after 30` successfully (clean report).
- Uploaded bundle `CO-06 TUI Relocation and Reuse Plan` to reMarkable path `/ai/2026/02/25/CO-06` and verified via `remarquee cloud ls`.
- Expanded implementation design with deeper `internal/geppettohost` details and explicit hard-cutover policy (no backward compatibility / no dual-run period).
- CO-10 cutover closeout summary:
  - Legacy runtime path removed from `cozodb-goja` (`cmd/cozo-tui`, `cmd/cozo-seed`, `internal/tui/*`).
  - Canonical app path is now `2026-02-18--cozodb-extraction/cozo-extraction-tui`.
  - Added CI guard target/workflow step in `cozodb-goja` to fail on reintroduced legacy paths.

## 2026-02-24

- Initial workspace created
