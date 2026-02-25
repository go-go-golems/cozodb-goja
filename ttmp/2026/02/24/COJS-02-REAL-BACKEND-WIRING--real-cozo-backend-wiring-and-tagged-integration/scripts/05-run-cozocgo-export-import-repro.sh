#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "${ROOT_DIR}"

CGO_LDFLAGS="-L${ROOT_DIR}/.deps/cozo" GOWORK=off \
  go run -tags cozo_cgo \
  ./ttmp/2026/02/24/COJS-02-REAL-BACKEND-WIRING--real-cozo-backend-wiring-and-tagged-integration/scripts/03-cozocgo-export-import-repro.go
