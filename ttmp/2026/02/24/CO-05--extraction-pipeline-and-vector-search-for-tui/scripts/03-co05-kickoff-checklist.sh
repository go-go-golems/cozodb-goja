#!/usr/bin/env bash
set -euo pipefail

# CO-05 kickoff checklist (safe/read-only except go test and go run).
# Run from repo root.

cd "$(dirname "$0")/../../../../../../.."  # move to cozodb-goja-init root

printf "== go work ==\n"
go work use .

printf "== module tidy ==\n"
cd cozodb-goja
go mod tidy

printf "== unit tests ==\n"
go test ./... -count=1

printf "== run TUI ==\n"
cat <<'EOM'
Next command (interactive):
  go run ./cmd/cozo-tui
EOM

printf "== env for live embedding/extraction ==\n"
cat <<'EOM'
export OPENAI_API_KEY=...
export PINOCCHIO_PROFILE=4o-mini
EOM
