#!/usr/bin/env bash
set -euo pipefail

# Opt-in integration run for real Cozo CGO backend relation lifecycle tests.
# Requires cozo_c symbols to be available on linker path.

ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "$ROOT_DIR"

GOWORK=off go test -tags 'cozo_cgo cozo_cgo_integration' ./pkg/cozoapi/module -count=1 -run TestRelLifecycleWithCozoCGOBackend
