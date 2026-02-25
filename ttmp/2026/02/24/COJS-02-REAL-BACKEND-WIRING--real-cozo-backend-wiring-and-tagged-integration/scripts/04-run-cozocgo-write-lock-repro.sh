#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "${ROOT_DIR}"

CGO_LDFLAGS="-L${ROOT_DIR}/.deps/cozo" GOWORK=off \
  go run -tags cozo_cgo \
  ./ttmp/2026/02/24/CO-04--tui-for-cozodb-relationship-explorer/scripts/01-cozo-write-lock-repro.go
