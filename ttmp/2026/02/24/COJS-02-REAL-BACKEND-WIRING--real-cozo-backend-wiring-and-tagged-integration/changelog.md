# Changelog

## 2026-02-24

- Initial workspace created


## 2026-02-24

Initialized COJS-02, authored implementation checklist, and created execution diary scaffold.

### Related Files

- /home/manuel/workspaces/2026-02-24/cozodb-goja-init/cozodb-goja/ttmp/2026/02/24/COJS-02-REAL-BACKEND-WIRING--real-cozo-backend-wiring-and-tagged-integration/reference/01-implementation-diary-cojs-02-real-backend-wiring.md — Kickoff diary step added
- /home/manuel/workspaces/2026-02-24/cozodb-goja-init/cozodb-goja/ttmp/2026/02/24/COJS-02-REAL-BACKEND-WIRING--real-cozo-backend-wiring-and-tagged-integration/tasks.md — Execution checklist replaced placeholder


## 2026-02-24

Implemented cozo_cgo adapter via cozo-lib-go, threaded engine/path/options through module open path, added tests, and validated normal plus tagged compile paths. Tagged runtime execution is blocked by missing libcozo_c linker dependency.

### Related Files

- /home/manuel/workspaces/2026-02-24/cozodb-goja-init/cozodb-goja/pkg/cozoapi/cozocgo/adapter_cozo_cgo.go — Tagged adapter implementation and result mapping
- /home/manuel/workspaces/2026-02-24/cozodb-goja-init/cozodb-goja/pkg/cozoapi/module/cozodb.go — Open option decoding extended with engine path options
- /home/manuel/workspaces/2026-02-24/cozodb-goja-init/cozodb-goja/pkg/cozoapi/module/default_open.go — DefaultOpen forwards cozo backend open options and defaults
- /home/manuel/workspaces/2026-02-24/cozodb-goja-init/cozodb-goja/pkg/cozoapi/module/default_open_test.go — Forwarding and default behavior tests


## 2026-02-24

Executed COJS-02 setup scripts; downloaded libcozo_c and successfully ran tagged cozocgo smoke command.

### Related Files

- /home/manuel/workspaces/2026-02-24/cozodb-goja-init/cozodb-goja/ttmp/2026/02/24/COJS-02-REAL-BACKEND-WIRING--real-cozo-backend-wiring-and-tagged-integration/scripts/01-download-libcozo-c.sh — Downloaded native static library for local linking
- /home/manuel/workspaces/2026-02-24/cozodb-goja-init/cozodb-goja/ttmp/2026/02/24/COJS-02-REAL-BACKEND-WIRING--real-cozo-backend-wiring-and-tagged-integration/scripts/02-run-cozocgo-smoke.sh — Executed tagged runtime smoke command with CGO_LDFLAGS


## 2026-02-24

Migrated tagged backend adapter from `cozo-lib-go` to `github.com/kraklabs/cie/pkg/cozodb@v0.7.20` to fix mutation lock failures, added immutable/read-only routing, updated JSON-based export/import handling, and added retroactive tracked validation scripts.

### Related Files

- /home/manuel/workspaces/2026-02-24/cozodb-goja-init/cozodb-goja/pkg/cozoapi/cozocgo/adapter_cozo_cgo.go — Adapter now uses `cie/pkg/cozodb` API and parses export envelopes
- /home/manuel/workspaces/2026-02-24/cozodb-goja-init/cozodb-goja/go.mod — Added `github.com/kraklabs/cie@v0.7.20`
- /home/manuel/workspaces/2026-02-24/cozodb-goja-init/cozodb-goja/ttmp/2026/02/24/COJS-02-REAL-BACKEND-WIRING--real-cozo-backend-wiring-and-tagged-integration/scripts/03-cozocgo-export-import-repro.go — Repro for backend export/import roundtrip
- /home/manuel/workspaces/2026-02-24/cozodb-goja-init/cozodb-goja/ttmp/2026/02/24/COJS-02-REAL-BACKEND-WIRING--real-cozo-backend-wiring-and-tagged-integration/scripts/04-run-cozocgo-write-lock-repro.sh — Wrapper script to run CO-04 write-lock repro
- /home/manuel/workspaces/2026-02-24/cozodb-goja-init/cozodb-goja/ttmp/2026/02/24/COJS-02-REAL-BACKEND-WIRING--real-cozo-backend-wiring-and-tagged-integration/scripts/05-run-cozocgo-export-import-repro.sh — Wrapper script to run export/import repro

## 2026-02-25

Ticket closed

