# Changelog

## 2026-02-25

- Created CO-07 phase ticket for hard-cutover foundation work (phases 1-3).
- Added detailed implementation design document covering module bootstrap, relocation parity, and plugin runtime foundation.
- Added granular task checklist with workstream-level breakdown and validation gates.
- Completed Workstream A bootstrap in extraction repo (commit `ebb3cb0`): module scaffold, `go mod init`, workspace registration, and baseline compile/test checks.
- Completed Workstream B relocation in extraction repo (commit `5089259`): moved F1-F7 TUI packages into `cozo-extraction-tui`, rewrote internal imports, validated via `go test ./...` and pseudo-TTY smoke run on `--engine mem`.
- Added per-ticket diaries for CO-07 through CO-10 and started chronological logging format for implementation tracking.
