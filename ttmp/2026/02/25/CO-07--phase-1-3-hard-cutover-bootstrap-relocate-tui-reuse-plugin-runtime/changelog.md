# Changelog

## 2026-02-25

- Created CO-07 phase ticket for hard-cutover foundation work (phases 1-3).
- Added detailed implementation design document covering module bootstrap, relocation parity, and plugin runtime foundation.
- Added granular task checklist with workstream-level breakdown and validation gates.
- Completed Workstream A bootstrap in extraction repo (commit `ebb3cb0`): module scaffold, `go mod init`, workspace registration, and baseline compile/test checks.
- Completed Workstream B relocation in extraction repo (commit `5089259`): moved F1-F7 TUI packages into `cozo-extraction-tui`, rewrote internal imports, validated via `go test ./...` and pseudo-TTY smoke run on `--engine mem`.
- Added per-ticket diaries for CO-07 through CO-10 and started chronological logging format for implementation tracking.
- Completed Workstream D foundation in extraction repo (commit `213e79f`): added `internal/plugins` loader/runner/types and `internal/geppettohost` runtime host with `cozodb`, `geppetto`, and `geppetto/plugins` registration, including timeout interrupts and panic recovery normalization.
- Completed Workstream E in extraction repo (commit `8696779`): relocated extractor JS scripts, added fixture script/transcript, and added `cmd/cozo-plugin-run` smoke runner command.
- Completed Workstream F in extraction repo (commit `c4eba2f`): added plugin validation tests and host integration tests; hardened descriptor parsing after a panic was exposed by tests.
- Added TUI parity smoke test suite in extraction repo (commit `1ee3a83`) with `cozo_cgo`-gated checks for screen fetch/load behavior plus app hotkey/quit behavior checks.
- Updated app parity checks to run on default backend (commit `007a1ae`) so F1-F7 hotkeys and quit-flow behavior are validated in default `go test` runs.
- Remaining blocker for full `cozo_cgo` parity suite: linker error in read-only module cache archive (`libcozo_c.a` missing index; attempted `ranlib` but permission denied).
- Validation summary:
  - `cozo-extraction-tui`: `go test ./... -count=1` passed
  - `cozodb-goja`: `go test ./pkg/cozoapi/... -count=1` passed
