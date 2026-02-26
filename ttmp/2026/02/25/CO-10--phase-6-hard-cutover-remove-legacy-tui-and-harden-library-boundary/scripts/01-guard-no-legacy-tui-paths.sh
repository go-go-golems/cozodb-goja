#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "${SCRIPT_DIR}/../../../../../.." && pwd)"

cd "${REPO_ROOT}"

echo "guard: validating legacy TUI paths are removed"

if [[ -e "cmd/cozo-tui/main.go" ]]; then
  echo "error: legacy path exists: cmd/cozo-tui/main.go"
  exit 1
fi

if [[ -e "cmd/cozo-seed/main.go" ]]; then
  echo "error: legacy path exists: cmd/cozo-seed/main.go"
  exit 1
fi

if [[ -d "internal/tui" ]] && find "internal/tui" -type f | grep -q .; then
  echo "error: legacy source files exist under internal/tui/"
  exit 1
fi

# Restrict grep to non-doc active files and ignore ticket workspace docs.
if rg -n --glob '!ttmp/**' --glob '!**/*.md' --glob '!**/*.txt' --glob '!**/.git/**' \
  '(internal/tui|cmd/cozo-tui|cmd/cozo-seed)' .; then
  echo "error: legacy TUI references found in active files"
  exit 1
fi

echo "guard: PASS (no legacy TUI paths or active references)"
