---
Title: Imported CozoDB JavaScript API proposal
Ticket: COJS-01-INITIAL-BUILD
Status: active
Topics:
    - cozodb
    - goja
    - javascript
    - api
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: User-provided API proposal imported from /tmp/cozodb-js.md.
LastUpdated: 2026-02-24T17:44:00-05:00
WhatFor: Source snapshot used as evidence for COJS-01 research.
WhenToUse: Use when reviewing upstream API/documentation details referenced in the design doc.
---

Here’s a JS/TS API design I’d use to embed **CozoDB** into a JavaScript sandbox in a way that feels modern, capability-oriented, and “use it for everything” practical—while still mapping cleanly onto Cozo’s real primitives (CozoScript + stored relations + query options + system ops).

This design intentionally supports **both**:

* **Node** via the official native binding (`cozo-node` / `CozoDb`) (async, supports callbacks + multi-transactions + named rules). ([GitHub][1])
* **Browser/WASM** via `cozo-lib-wasm` (sync by default, blocks the main thread unless you run it in a Worker). ([GitHub][2])

---

## Design goals

1. **One async interface everywhere**

   * Even if WASM is sync, you wrap it behind a Worker/port so sandbox code always does `await db.exec(...)`.

2. **CozoScript stays first-class**

   * It’s already the “universal API” for relations + graph + vector + time-travel. Don’t hide it—make it ergonomic.

3. **Safe-by-default parameterization**

   * Template-tag queries compile to CozoScript with `$params` so sandbox code doesn’t build strings manually.

4. **Sandbox-friendly**

   * The object you expose into the sandbox is pure JSON-in/JSON-out + Promises (easy to bridge across isolates).

5. **Policy hooks**

   * Host can enforce timeouts/limits, block system ops, block certain relations, etc.
   * Cozo has access levels (`normal/protected/read_only/hidden`) that are useful guardrails (but not a security boundary against malicious code). ([CozoDB][3])

---

## Public API (TypeScript)

### Core types

```ts
export type CozoValue =
  | null
  | boolean
  | number
  | string
  | Uint8Array
  | CozoValue[]
  | { [k: string]: CozoValue };

export type CozoParams = Record<string, CozoValue>;

export type CozoRow = CozoValue[];
export type CozoHeaders = string[];

export interface CozoResult {
  headers: CozoHeaders;
  rows: CozoRow[];
  took?: number;
  count?: number;
  next?: unknown;

  // ergonomic helpers (no extra allocations unless you call them)
  objects<T extends Record<string, CozoValue> = Record<string, CozoValue>>(): T[];
  firstObject<T extends Record<string, CozoValue> = Record<string, CozoValue>>(): T | undefined;
  scalar<T extends CozoValue = CozoValue>(): T | undefined;
}

export interface QueryOptions {
  // Maps to Cozo query options like :limit/:offset/:timeout
  // (:timeout defaults to 300s in Cozo, 0 disables) :contentReference[oaicite:3]{index=3}
  limit?: number;
  offset?: number;
  timeoutSec?: number;

  // Node-only optimization/safety (CozoDb.run(..., immutable)) :contentReference[oaicite:4]{index=4}
  immutable?: boolean;
}

export interface PreparedQuery {
  script: string;
  params: CozoParams;
  opts?: QueryOptions;
}
```

### The database handle you expose

```ts
export interface CozoDb {
  readonly backend: "node" | "wasm" | "remote";
  readonly capabilities: {
    persistence: boolean;
    backupRestore: boolean;
    callbacks: boolean;
    namedRules: boolean;
  };

  /** Run raw CozoScript (power user escape hatch). */
  exec(script: string, params?: CozoParams, opts?: QueryOptions): Promise<CozoResult>;

  /** Execute a prepared query (produced by cq``). */
  exec(q: PreparedQuery): Promise<CozoResult>;

  /** Tagged-template that compiles parameters safely and executes. */
  q(strings: TemplateStringsArray, ...values: unknown[]): Promise<CozoResult>;
  q(opts: QueryOptions): (strings: TemplateStringsArray, ...values: unknown[]) => Promise<CozoResult>;

  /** Tagged-template that compiles parameters safely without executing. */
  cq(strings: TemplateStringsArray, ...values: unknown[]): PreparedQuery;
  cq(opts: QueryOptions): (strings: TemplateStringsArray, ...values: unknown[]) => PreparedQuery;

  /** Atomic batch: runs multiple queries in one Cozo transaction. */
  atomic(queries: PreparedQuery[], opts?: { write?: boolean }): Promise<CozoResult>;

  /** Higher-level relation helper (still compiles down to CozoScript). */
  rel(name: string): Relation;

  /** Import/export relations (snapshot style). */
  export(relations: string[]): Promise<Record<string, { headers: string[]; rows: CozoRow[] }>>;
  import(data: Record<string, { headers: string[]; rows: CozoRow[] }>): Promise<void>;

  /** Close/free native resources. Node binding requires explicit close(). :contentReference[oaicite:5]{index=5} */
  close(): Promise<void>;
}
```

### Relation helper (CRUD without writing CozoScript every time)

This compiles to documented stored-relation mutation ops like `:create`, `:put`, `:rm`, `:insert`, `:update`, etc. ([CozoDB][4])

```ts
export interface RelationSpec {
  keys: Record<string, string | undefined>;    // e.g. { id: "Uuid" }
  values?: Record<string, string | undefined>; // e.g. { name: "String", meta: "Json?" }
}

export interface Relation {
  readonly name: string;

  create(spec: RelationSpec, opts?: { replace?: boolean }): Promise<void>;

  put(
    rows: Array<Record<string, CozoValue>> | CozoRow[],
    opts?: { returning?: boolean }
  ): Promise<CozoResult>;

  insert(rows: Array<Record<string, CozoValue>> | CozoRow[], opts?: { returning?: boolean }): Promise<CozoResult>;
  update(rows: Array<Record<string, CozoValue>> | CozoRow[], opts?: { returning?: boolean }): Promise<CozoResult>;
  rm(keys: Array<Record<string, CozoValue>> | CozoRow[], opts?: { returning?: boolean }): Promise<CozoResult>;
  del(rows: Array<Record<string, CozoValue>> | CozoRow[], opts?: { returning?: boolean }): Promise<CozoResult>;

  /** Query this relation using *rel{col: $param} style bindings. :contentReference[oaicite:7]{index=7} */
  get(key: Record<string, CozoValue>): Promise<Record<string, CozoValue> | undefined>;

  /** System ops sugar (metadata). */
  columns(): Promise<CozoResult>;
  indices(): Promise<CozoResult>;

  /** Guardrail access levels (not a security boundary). :contentReference[oaicite:8]{index=8} */
  access(level: "normal" | "protected" | "read_only" | "hidden"): Promise<void>;
}
```

---

## Ergonomics: parameter-safe template tags

### Execute immediately

```js
// Safe: ${...} becomes a $param, not string concatenation.
const res = await db.q`
  ?[name, age] := *users{name, age}, age >= ${18}
  :limit ${50}
`;
console.log(res.objects());
```

### Compile (prepare) without running

Useful for batching/transactions:

```js
const q1 = db.cq`?[id, name] := *users{id, name}`;
const q2 = db.cq`?[id] <- [[${"u1"}]] :rm users {id} :returning`;

const final = await db.atomic([q1, q2]);
```

### Why this maps well to Cozo

* Cozo already supports `$param` binding in scripts.
* Query options like `:timeout`, `:limit`, `:offset` are part of the language and can be injected by the wrapper. ([CozoDB][5])
* Multiple `{ ... }` queries in one script run as **one transaction** and return the last result. ([CozoDB][4])

---

## Transactions: `db.atomic([...])` (portable + sandbox-friendly)

Instead of relying on Node-only “multiTransact”, you get a single primitive that works everywhere:

```js
await db.atomic([
  db.cq`?[id, name] <- [[${"u1"}, ${"Ada"}]] :put users {id => name}`,
  db.cq`?[id] <- [[${"u2"}]] :rm users {id}`,
  db.cq`?[id, name] := *users{id, name} :sort +id`
]);
```

Internally it compiles to:

* `{ <query1> } { <query2> } { <query3> }`

Which Cozo documents as a single transaction boundary per script. ([CozoDB][4])

---

## Relation helper examples (CRUD without writing specs manually)

### Create a relation

```js
await db.rel("users").create({
  keys: { id: "String" },
  values: { name: "String", meta: "Json?" }
});
```

This compiles to `:create users { id: String => name: String, meta: Json? }` (Cozo stored relation spec). ([CozoDB][4])

### Put rows + return what changed

```js
const changed = await db.rel("users").put(
  [{ id: "u1", name: "Ada", meta: { team: "db" } }],
  { returning: true }
);
console.log(changed.rows);
```

`:returning` behavior is documented (adds `_kind` + returns affected rows). ([CozoDB][4])

---

## Sandbox embedding: Port-based API (the key idea)

In a real sandbox you usually can’t (or shouldn’t) hand out a native DB handle directly. Instead:

* Host owns the real `CozoDb`
* Sandbox gets a *capability port* with a single `request()` method
* A small client wraps it back into `db.exec/db.q/...`

### The “port” protocol

```ts
export type CozoRequest =
  | { op: "exec"; script: string; params?: CozoParams; opts?: QueryOptions }
  | { op: "export"; relations: string[] }
  | { op: "import"; data: Record<string, { headers: string[]; rows: CozoRow[] }> }
  | { op: "close" };

export type CozoResponse =
  | { ok: true; result?: any }
  | { ok: false; error: { message: string; code?: string; display?: string } };
```

### Host side

```js
function createCozoSandboxPort(db, policy) {
  return {
    async request(req) {
      // policy enforcement lives here:
      // - enforce a max :timeout (Cozo supports :timeout; default is 300s) :contentReference[oaicite:14]{index=14}
      // - optionally block scripts starting with :: (system ops)
      // - optionally block access to certain relations (or set ::access_level hidden/read_only) :contentReference[oaicite:15]{index=15}
      // - clamp :limit/:offset
      return dispatch(db, req, policy);
    }
  };
}
```

### Sandbox side

```js
const db = connectCozo(port);

await db.q`?[x] <- [[${1}]]`;
```

This is also how you’d run WASM in a Worker: Worker implements the same `port.request()` shape, so the sandbox can’t tell whether it’s local Node, Worker WASM, or remote.

---

## Backend adapters (how this maps onto real Cozo APIs)

### Node adapter

* Uses `new CozoDb(engine, path, options)` and `await db.run(script, params, immutable)` (note `immutable`). ([GitHub][6])
* Must call `close()` to release resources. ([GitHub][1])
* Optional advanced features:

  * `registerCallback(relation, cb, capacity)` for change notifications. ([GitHub][6])
  * `multiTransact(write)` if you want Node-only “interactive transactions”. ([GitHub][6])
  * `registerNamedRule(...)` for custom fixed rules. ([GitHub][6])

### WASM adapter

* `await init()` then `CozoDb.new()` then `db.run(script, paramsJsonString)` returning JSON string. ([GitHub][2])
* Because it’s synchronous + blocks, the idiomatic wrapper runs it in a Worker and exposes it via the same port interface. ([GitHub][2])

---

## One important JS gotcha (and how the API should treat it)

Cozo has distinct runtime numeric types (`Int` vs `Float`) and even warns that JS can make key usage confusing because JS numbers are all floating-point. ([CozoDB][7])

Two practical API design choices:

* Encourage **String/Uuid** keys for most sandbox data models.
* Optionally expose helpers like `cozo.uuid(str)` / `cozo.bytes(u8)` / `cozo.json(obj)` so sandbox code is explicit about non-scalars.

Cozo supports `Uuid`, `Bytes`, `Json`, `Vector`, etc. ([CozoDB][7])

---

## What you get from this design

* **Ergonomic “daily driver”**: `db.q\`...``+`res.objects()`
* **Power user mode**: `db.exec(rawCozoScript)`
* **Atomic operations** (portable): `db.atomic([ ... ])` leveraging Cozo’s multi-query transaction scripting ([CozoDB][4])
* **Sandbox-safe**: you can expose only a port, and enforce policy/timeouts centrally using Cozo’s `:timeout` and relation access levels ([CozoDB][5])
* **Backend-flexible**: Node native binding vs WASM Worker both fit

---

If you tell me what your sandbox actually is (Node `vm`, `isolated-vm`, QuickJS, browser iframe, etc.), I can tailor the **port/bridge shape** (structured clone vs JSON, transferable buffers, streaming results, etc.) while keeping the same public API surface.

[1]: https://raw.githubusercontent.com/cozodb/cozo/main/cozo-lib-nodejs/README.md "raw.githubusercontent.com"
[2]: https://raw.githubusercontent.com/cozodb/cozo/main/cozo-lib-wasm/README.md "raw.githubusercontent.com"
[3]: https://docs.cozodb.org/en/latest/sysops.html "System ops - CozoDB v0.7"
[4]: https://docs.cozodb.org/en/latest/stored.html "Stored relations and transactions - CozoDB v0.7"
[5]: https://docs.cozodb.org/en/latest/queries.html "Queries - CozoDB v0.7"
[6]: https://raw.githubusercontent.com/cozodb/cozo/main/cozo-lib-nodejs/index.js "raw.githubusercontent.com"
[7]: https://docs.cozodb.org/en/latest/datatypes.html "Types - CozoDB v0.7"

