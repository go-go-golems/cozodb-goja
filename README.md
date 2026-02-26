# cozodb-goja

`cozodb-goja` is an initial Go + goja implementation of the CozoDB JavaScript API contract described in ticket `COJS-01-INITIAL-BUILD`.

## Status

- Implemented core API domain layer in `pkg/cozoapi`:
  - `exec`, `q`, `cq`, `atomic`
  - relation compiler + helpers, including `db.rel(name).put/insert/update/rm/del/get/columns/indices/access`
  - query policy enforcement (limits/timeouts/system-op and relation restrictions)
- Implemented native goja module adapter in `pkg/cozoapi/module`:
  - `require("cozodb").open({ backend })`
  - db handle methods: `exec`, `q`, `cq`, `atomic`, `rel`, `export`, `import`, `close`
- Added fake backend + tests for deterministic behavior.
- Added optional `cozo_cgo` adapter scaffold.

## CLI

Run inline code:

```bash
GOWORK=off go run ./cmd/cozo --eval 'const db = require("cozodb").open(); db.backend'
```

Run a script file:

```bash
GOWORK=off go run ./cmd/cozo --script ./examples/demo.js
```

Start REPL:

```bash
GOWORK=off go run ./cmd/cozo
```

## JavaScript usage

```js
const cozo = require("cozodb");
const db = cozo.open({ backend: "fake" });

db.exec("?[x] <- [[1]]")
  .then((res) => {
    console.log(res.scalar());
  });
```

Prepared + atomic:

```js
const q1 = db.cq`?[id] <- [[${"u1"}]]`;
const q2 = db.cq`?[id] <- [[${"u2"}]]`;

db.atomic([q1, q2]).then((res) => console.log(res.rows));
```

Relation helper:

```js
const users = db.rel("users");
users.put([{ id: "u1", name: "Ada" }], { returning: true });
```
