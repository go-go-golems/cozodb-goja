#!/usr/bin/env bash
set -euo pipefail

CUR_DIR=$(cd "$(dirname "$0")" && pwd)
ROOT_DIR="$CUR_DIR"
while [[ "$ROOT_DIR" != "/" ]]; do
  if [[ -d "$ROOT_DIR/go-go-goja" && -d "$ROOT_DIR/2026-02-18--cozodb-extraction" ]]; then
    break
  fi
  ROOT_DIR=$(dirname "$ROOT_DIR")
done

if [[ "$ROOT_DIR" == "/" ]]; then
  echo "Could not locate workspace root containing go-go-goja and 2026-02-18--cozodb-extraction" >&2
  exit 1
fi

OUT_DIR=$(dirname "$0")
OUT_FILE="$OUT_DIR/01-evidence-scan.out"

{
  echo "# COJS-01 evidence scan"
  date -Iseconds
  echo "Workspace root: $ROOT_DIR"
  echo

  echo "## go-go-goja runtime composition"
  rg -n "WithModules|DefaultRegistryModules|NewRuntime|NewRunner|RunOnLoop|Close\(" "$ROOT_DIR/go-go-goja/engine" "$ROOT_DIR/go-go-goja/pkg/runtimeowner" -S
  echo

  echo "## go-go-goja evaluator wiring"
  rg -n "type Config|EnableModules|New\(|Evaluate\(|Reset\(|Close\(|GetAvailableModules" "$ROOT_DIR/go-go-goja/pkg/repl/evaluators/javascript/evaluator.go" -S
  echo

  echo "## cozo relationship runner runtime patterns"
  rg -n "eventloop.NewEventLoop|NewRunner|require.NewRegistry|RegisterNativeModule|multiTransact|registerNamedRule|run\(script" "$ROOT_DIR/2026-02-18--cozodb-extraction/cozo-relationship-js-runner" -S
  echo

  echo "## cozo-lib-go core APIs"
  rg -n "type CozoDB|func New\(|func \(db \*CozoDB\) Run|ImportRelations|ExportRelations|Backup|Restore|ImportRelationsFromBackup" "$ROOT_DIR/2026-02-18--cozodb-extraction/ttmp/2026/02/18/COZO-01-INITIAL-ASSESSMENT--initial-assessment-python-cozo-extraction-to-go-geppetto-pinocchio/sources/cozo-lib-go-cozo.go" -S
  echo

  echo "## goja concurrency and promise safety"
  rg -n "goroutine-safe|not goroutine-safe|Interrupt|ClearInterrupt|NewPromise|RunString" "$ROOT_DIR/goja/README.md" "$ROOT_DIR/goja/runtime.go" "$ROOT_DIR/goja/builtin_promise.go" -S
} > "$OUT_FILE"

echo "Wrote $OUT_FILE"
