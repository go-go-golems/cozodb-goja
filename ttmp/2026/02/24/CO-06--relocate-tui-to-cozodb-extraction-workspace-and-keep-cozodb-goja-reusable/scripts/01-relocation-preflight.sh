#!/usr/bin/env bash
set -euo pipefail

# CO-06 preflight inventory script.
# Read-only diagnostics for relocation planning.

ROOT="/home/manuel/workspaces/2026-02-24/cozodb-goja-init"
cd "$ROOT"

echo "== workspace =="
pwd

echo "== go.work =="
cat go.work

echo "== extraction root module/workspace =="
if [[ -f 2026-02-18--cozodb-extraction/go.mod ]]; then
  echo "found extraction root go.mod"
else
  echo "no extraction root go.mod"
fi
if [[ -f 2026-02-18--cozodb-extraction/go.work ]]; then
  echo "found extraction root go.work"
else
  echo "no extraction root go.work"
fi

echo "== tUI LOC (current in cozodb-goja/internal/tui) =="
find cozodb-goja/internal/tui -type f -name '*.go' -print0 | xargs -0 wc -l | tail -n 1

echo "== cozoapi LOC (reusable bindings) =="
find cozodb-goja/pkg/cozoapi -type f -name '*.go' -print0 | xargs -0 wc -l | tail -n 1

echo "== internal import constraints =="
rg -n "github.com/go-go-golems/cozodb-goja/internal/tui" cozodb-goja/internal/tui -g '*.go' || true

echo "== plugin loader candidates =="
ls -1 2026-02-18--cozodb-extraction/cozo-relationship-js-runner/plugin_loader.go \
      2026-02-18--cozodb-extraction/cozo-relationship-js-runner/main.go \
      2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/lib/relationship_extractor_factory.js

echo "== done =="
