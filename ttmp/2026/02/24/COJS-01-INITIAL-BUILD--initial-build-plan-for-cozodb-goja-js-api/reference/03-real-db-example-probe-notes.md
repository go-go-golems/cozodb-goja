# Real DB JS API Probes (2026-02-25)

## Goal
Validate real `cozo_cgo` behavior for JS API examples and capture reproducible probe scripts under ticket `scripts/`.

## Probe scripts (ticket-local)
- `scripts/02-create-relation-probe.js`
- `scripts/03-import-probe.js`
- `scripts/04-put-without-create-probe.js`
- `scripts/05-create-via-exec-probe.js`
- `scripts/06-create-no-replace-probe.js`
- `scripts/02-run-probes.sh`

## Findings
1. `rel(name).create(..., { Replace: true })` rejects on real backend with `Program has no entry`.
2. `rel(name).create(...)` (without `Replace`) succeeds on fresh DB.
3. `rel(name).put(...)` against a missing relation rejects with `Cannot find requested stored relation`.
4. `db.import(...)` requires relation pre-existence.
5. `rel(name).get(...)` currently rejects on real backend (`Symbol 'row' in rule head is unbound`) in this integration path.

## Practical adjustments applied in examples
1. Seed flow uses `create` without `Replace` on a fresh sqlite DB file.
2. Example runner deletes the sqlite file before running all scripts, ensuring deterministic schema creation.
3. Scripts avoid `rel.get` in the happy-path suite and use explicit `exec` queries for lookup reads.
4. Promise-returning values are decoded host-side so real query payloads are emitted as JSON.
