# Changelog

## 2026-02-25

- Created CO-11 ticket workspace for Glazed-first config cutover.
- Added CO-11 design document: `design-doc/01-glazed-first-configuration-architecture-and-implementation-plan.md`.
- Added CO-11 research diary: `reference/01-investigation-diary-glazed-cutover.md`.
- Completed deep code evidence pass across:
  - `cozo-extraction-tui` command and runtime env usage,
  - Geppetto section/middleware bootstrap chain,
  - Geppetto profile registry adapter and precedence tests,
  - Geppetto embeddings parsed-values factory path.
- Captured phased implementation plan and granular task backlog for hard cutover.
- Related core evidence files to CO-11 design and diary docs with `docmgr doc relate`.
- Added ticket vocabulary slugs for `cozo`, `embeddings`, `geppetto`, and `glazed`.
- Validation: `docmgr doctor --ticket CO-11 --stale-after 30` passes cleanly.
- reMarkable delivery:
  - Dry run passed for bundle upload to `/ai/2026/02/25/CO-11`.
  - Uploaded `CO-11 Glazed Config Cutover Research.pdf`.
  - Verified remote listing via `remarquee cloud ls /ai/2026/02/25/CO-11 --long --non-interactive`.
- Locked implementation policy from user direction: simplest path, hard cutover, no backwards compatibility, precedence fixed to `flags > env > profiles > config > defaults` for this iteration.
- Implementation commit `96cc0b9` in `2026-02-18--cozodb-extraction`:
  - Added Glazed command package for `cozo-tui` and `cozo-plugin-run` entrypoints.
  - Added app config sections and custom middleware bootstrap chain.
  - Removed non-test runtime `os.Getenv` usage in F9 paths.
  - Switched seed/F9 embedding flow to explicit parsed-values provider injection.
  - Removed env-based lazy embedding provider fallback from `geppettohost`.
  - Replaced stale env-bridge tests and removed obsolete `cmd/cozo-tui/main_test.go`.
  - Validation passed: `go test ./... -count=1` in `cozo-extraction-tui`.
