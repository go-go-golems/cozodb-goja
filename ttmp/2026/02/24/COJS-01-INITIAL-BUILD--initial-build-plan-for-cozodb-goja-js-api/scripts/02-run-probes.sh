#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
COZO_EXTRACT_ROOT="${COZO_EXTRACT_ROOT:-/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-extraction-tui}"
COZO_LIB_DIR="${COZO_LIB_DIR:-/home/manuel/workspaces/2026-02-24/cozodb-goja-init/cozodb-goja/.deps/cozo}"
PROBE_DB="${PROBE_DB:-/tmp/cozo-probe.db}"
TRANSCRIPT="${TRANSCRIPT:-${COZO_EXTRACT_ROOT}/scripts/examples/cozodb-js/00-transcript.txt}"

cd "${COZO_EXTRACT_ROOT}"

COMMON_ARGS=(
  --plugin-transcript "${TRANSCRIPT}"
  --plugin-script-root "${SCRIPT_DIR}"
  --plugin-engine-options-json "{\"backend\":\"cozo_cgo\",\"engine\":\"sqlite\",\"path\":\"${PROBE_DB}\",\"options\":{}}"
)

CGO_LDFLAGS="-L${COZO_LIB_DIR}" go run -tags cozo_cgo ./cmd/cozo-plugin-run \
  --plugin-script "${SCRIPT_DIR}/02-create-relation-probe.js" \
  "${COMMON_ARGS[@]}"

CGO_LDFLAGS="-L${COZO_LIB_DIR}" go run -tags cozo_cgo ./cmd/cozo-plugin-run \
  --plugin-script "${SCRIPT_DIR}/03-import-probe.js" \
  "${COMMON_ARGS[@]}"
