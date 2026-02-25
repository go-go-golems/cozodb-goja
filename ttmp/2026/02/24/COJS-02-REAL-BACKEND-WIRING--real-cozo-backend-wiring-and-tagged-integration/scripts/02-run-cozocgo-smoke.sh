#!/usr/bin/env bash
set -euo pipefail

# Run a quick tagged smoke command against the cozocgo backend.
# This script expects libcozo_c.a to exist already (see 01-download-libcozo-c.sh).

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "${SCRIPT_DIR}" rev-parse --show-toplevel)"

COZO_DIR="${COZO_DIR:-${REPO_ROOT}/.deps/cozo}"

if [[ ! -f "${COZO_DIR}/libcozo_c.a" ]]; then
  echo "Missing ${COZO_DIR}/libcozo_c.a"
  echo "Run scripts/01-download-libcozo-c.sh first (or set COZO_DIR)."
  exit 1
fi

cd "${REPO_ROOT}"
CGO_LDFLAGS="-L${COZO_DIR}" \
  go run -tags cozo_cgo ./cmd/XXX --eval 'require("cozodb").open({backend:"cozocgo"})'

