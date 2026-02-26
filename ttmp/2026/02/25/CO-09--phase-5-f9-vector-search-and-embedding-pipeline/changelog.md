# Changelog

## 2026-02-25

- Created CO-09 phase ticket for F9 vector search and embeddings.
- Added detailed implementation design document covering embedding integration, HNSW query templates, mode controls, and validation strategy.
- Added granular execution task checklist.
- Added CO-09 diary and related-file links; queued implementation behind earlier phases.
- Completed Workstreams A-B scaffold in extraction repo (commit `1aeafe4`): added `internal/tui/screens/vsearch/model.go`, F9 router/tab/resize wiring, and initial control state/input bindings (query, mode, k, ef, limit, reset).
- Completed Workstreams C-H in extraction repo (commit `dacc589`): added geppettohost embedding API, async F9 embed/query message flow, mode-specific Cozo query builders, row decoder, result/detail rendering, vector schema/index preflight with optional index auto-create hook, and comprehensive tests including env-gated live embedding + `cozo_cgo` integration test.
- Validation: `go test ./... -count=1` passes without `cozo_cgo` tag; `go test -tags cozo_cgo ./internal/tui/screens/vsearch -count=1` currently fails due known linker issue (`libcozo_c.a: archive has no index; run ranlib`).
- Updated CO-09 tasks/diary/file-relations for Step 3 and reran `docmgr doctor --ticket CO-09 --stale-after 30` (passes).
- Attempted local archive-index remediation (`ranlib`) for `github.com/kraklabs/cie@v0.7.20/lib/libcozo_c.a`; tagged builds remain blocked with unresolved `cozo_*` symbols and `nm` reports malformed archive, so manual F9 smoke remains blocked pending a valid static artifact.
- Added `cozo-extraction-tui/Makefile` with repeatable `.deps`-based targets (`test-cgo-vsearch`, `run-cgo-tui`) and fixed parser-incompatible `ModeAll` query template to `or`-branch union form (commit `cb28ba2`).
- Revalidation: `make test-cgo-vsearch` now passes with `CGO_LDFLAGS` pointed at `cozodb-goja/.deps/cozo`; `make run-cgo-tui` reaches DB open but cannot fully execute in this non-interactive session due missing `/dev/tty`.
- Implemented real-embedding seeding path (commit `1ee1b80`): seed schema now includes embedding/metadata columns, seed insertion generates live embeddings through `geppettohost` using Pinocchio profile/config loading (`~/.pinocchio/config.yaml`, `~/.config/pinocchio/profiles.yaml`) with env overrides, and HNSW indices are created during seed.
- Added env-gated live seed smoke (`COZO_TUI_LIVE_SEED_SMOKE=1`) and verified:
  - `CGO_LDFLAGS="-L.../cozodb-goja/.deps/cozo" COZO_TUI_LIVE_SEED_SMOKE=1 go test -tags cozo_cgo ./internal/tui/seeddata -run TestSeedIfEmptyLiveEmbeddingsSmoke -count=1` passes.
- Updated upstream embedding provider behavior in geppetto (commit `4c913db`) so OpenAI embedding requests now send configured `dimensions`, enabling 384-d outputs for this app schema.
- Added explicit embedding CLI flags and non-interactive seed mode in `cozo-tui` (commit `c63f24b`): `--embed-provider`, `--embed-engine`, `--embed-dimensions`, credential/base-url overrides, Pinocchio overrides, and `--seed-only` so seed + real credentials can be validated without TUI `/dev/tty`.
- Added repeatable Make target `run-cgo-seed-only` with `SEED_EMBED_FLAGS` passthrough for live-provider smoke runs from `.deps`-linked native backend.
- Verification:
  - `go test ./cmd/cozo-tui ./internal/geppettohost ./internal/tui/screens/vsearch -count=1` passes.
  - `make run-cgo-seed-only SEED_DB=/tmp/cozo-tui-real-seed-make.db SEED_EMBED_FLAGS='--embed-provider openai --embed-engine text-embedding-3-small --embed-dimensions 384'` completes successfully against real provider settings.

## 2026-02-25

Ticket closed

