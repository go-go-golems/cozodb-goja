#!/usr/bin/env bash
set -euo pipefail

# Download libcozo_c static library for local cozo_cgo linking.
# This script only downloads and extracts the library; it does not run the app.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "${SCRIPT_DIR}" rev-parse --show-toplevel)"

COZO_VERSION="${COZO_VERSION:-0.7.5}"
COZO_PLATFORM="${COZO_PLATFORM:-x86_64-unknown-linux-gnu}"
COZO_DIR="${COZO_DIR:-${REPO_ROOT}/.deps/cozo}"

mkdir -p "${COZO_DIR}"

URL="https://github.com/cozodb/cozo/releases/download/v${COZO_VERSION}/libcozo_c-${COZO_VERSION}-${COZO_PLATFORM}.a.gz"
ARCHIVE_PATH="${COZO_DIR}/libcozo_c.a.gz"

echo "Downloading: ${URL}"
curl -L "${URL}" -o "${ARCHIVE_PATH}"
gunzip -f "${ARCHIVE_PATH}"

echo
echo "Downloaded: ${COZO_DIR}/libcozo_c.a"
echo "Next step (manual):"
echo "  CGO_LDFLAGS=\"-L${COZO_DIR}\" go run -tags cozo_cgo ./cmd/XXX --eval 'require(\"cozodb\").open({backend:\"cozocgo\"})'"

