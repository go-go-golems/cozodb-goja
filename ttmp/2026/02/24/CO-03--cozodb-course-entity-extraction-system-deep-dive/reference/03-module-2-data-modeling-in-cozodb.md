---
Title: "Module 2 - Data Modeling in CozoDB"
Ticket: CO-03
Status: active
Topics:
    - cozodb
    - course
    - entity-extraction
    - datalog
    - schema-design
    - vector-search
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo_demo.py:Schema creation and data insertion patterns (lines 20-138)"
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/schema_design.md:Full schema design document for the entity extraction system"
ExternalSources:
    - https://docs.cozodb.org/en/latest/stored.html
    - https://docs.cozodb.org/en/latest/datatypes.html
    - https://docs.cozodb.org/en/latest/vector.html
    - https://docs.cozodb.org/en/latest/nonscript.html
Summary: "Data modeling in CozoDB: stored relations, schema design, data types, mutations, indexing (standard, HNSW vector, FTS, MinHash-LSH), and real-world schema design from the entity extraction project"
LastUpdated: 2026-02-24T19:00:00-05:00
WhatFor: "Second module of the CozoDB course - covers everything about persisting and structuring data"
WhenToUse: "After completing Module 1, when you need to understand how to design schemas, insert/update data, and create indices in CozoDB"
---

# Module 2: Data Modeling in CozoDB

## Learning Objectives

By the end of this module you will:
- Understand stored relations and their schema specification syntax
- Know all CozoDB data types and when to use each one
- Create, modify, and drop relations
- Insert, update, and delete data using the full range of mutation operations
- Design real-world schemas with thoughtful key vs value decisions
- Create standard indices for accelerating queries
- Create HNSW vector indices for semantic similarity search
- Create full-text search (FTS) indices for text queries
- Understand MinHash-LSH indices for near-duplicate detection

---

## 2.1 Stored Relations

In Module 1, we worked with constant rules (`<-`) and inline rules (`:=`). Those are ephemeral -- the data only exists for the duration of the query. **Stored relations** are the persistent, on-disk data structures in CozoDB. They survive across queries, sessions, and restarts (when using SQLite or RocksDB backends).

A stored relation is like a table in SQL, but with an important difference: it is an **ordered key-value store**. Every stored relation has a schema that divides its columns into **key columns** and **value columns**.

### Schema Specification Syntax

The schema uses the `{key_cols => value_cols}` syntax:

```cozoscript
{
    key_col1: Type1,
    key_col2: Type2
    =>
    value_col1: Type3,
    value_col2: Type4
}
```

- Columns **before** the `=>` are **key columns**
- Columns **after** the `=>` are **value columns**
- If **all columns are keys** (no value columns), the `=>` can be omitted

### Keys vs Values

This distinction is fundamental:

| Aspect | Key Columns | Value Columns |
|---|---|---|
| Uniqueness | Combination of all key columns must be unique | No uniqueness constraint |
| Storage order | Rows sorted lexicographically by key columns | N/A |
| Lookup | Efficient prefix-key lookup | Requires scanning |
| Upsert behavior | Used to identify the row | Replaced on `:put` |

**Example**: In a `person` relation with `{id => name, description}`:
- `id` is the key -- every person has a unique `id`
- `name` and `description` are values -- they "belong to" the row identified by that `id`
- If you `:put` a row with an existing `id`, the `name` and `description` are replaced

### Storage Order

Rows in a stored relation are physically stored in **lexicographic order** by their key columns. This means:

- Queries that filter on key prefixes are fast (range scans)
- Composite keys are sorted left-to-right (first by `key_col1`, then `key_col2`, etc.)
- String keys are sorted alphabetically; integer keys numerically

```cozoscript
# A relation with composite key (department, employee_id)
# Rows are stored sorted by department first, then employee_id
:create staff {
    department: String,
    employee_id: Int
    =>
    name: String
}
```

This means looking up all employees in a specific department is efficient -- they're stored contiguously.

---

## 2.2 Data Types

CozoDB supports the following data types:

### Primitive Types

| Type | Description | Example Values |
|---|---|---|
| `String` | UTF-8 text | `'hello'`, `'sarah_martinez'` |
| `Int` | 64-bit signed integer | `42`, `-1`, `0` |
| `Float` | 64-bit floating point (IEEE 754) | `3.14`, `-0.5`, `1.0e10` |
| `Bool` | Boolean | `true`, `false` |
| `Bytes` | Byte array | (from host language) |
| `Null` | The null value | `null` |

### Special Types

| Type | Description | Example |
|---|---|---|
| `Any` | Dynamic type -- accepts any value | Useful for flexible schemas |
| `Json` | JSON values | Stored as structured JSON data |
| `Validity` | Temporal validity marker | Used for time-travel queries (Module 5) |

### Vector Type

```
<F32; N>
```

Fixed-dimension float32 vectors, where `N` is the number of dimensions. This is critical for embedding-based search.

```cozoscript
# A 384-dimensional embedding (common for sentence-transformers like all-MiniLM-L6-v2)
embedding: <F32; 384>

# A smaller 128-dimensional vector
small_vec: <F32; 128>
```

Vectors are stored efficiently and can be indexed with HNSW indices for approximate nearest-neighbor search.

### Compound Types

```
List[T]
```

A typed list where all elements must be of type `T`:

```cozoscript
# A list of strings
tags: List[String]

# A list of integers
scores: List[Int]
```

### Nullable Types

Append `?` to any type to make it nullable:

```cozoscript
# Required -- must always have a value
name: String

# Nullable -- can be null
location: String?
description: Float?
```

When a column is non-nullable, attempting to insert a `null` value will produce an error.

### Type Coercion Rules

CozoDB performs some implicit type coercion:

- `Int` can be coerced to `Float` (e.g., `42` becomes `42.0`)
- `Null` is accepted for any nullable type (`T?`)
- Vectors must match their declared dimension exactly -- `<F32; 384>` requires exactly 384 elements
- `Any` accepts all types without coercion
- No implicit coercion between `String` and numeric types -- use `to_float()`, `to_int()`, or `to_string()` explicitly

---

## 2.3 Creating Relations

### `:create` -- Create New Relation

Creates a new stored relation. **Errors if the relation already exists.**

```cozoscript
:create person {
    id: String
    =>
    name: String,
    description: String
}
```

This is the "safe" creation command -- it won't accidentally overwrite an existing relation.

### `:replace` -- Create or Replace Relation

Creates a new relation, or **fully replaces** an existing one (drops all data and redefines the schema).

```cozoscript
:replace person {
    id: String
    =>
    name: String,
    description: String,
    age: Int
}
```

**Warning**: `:replace` destroys all existing data in the relation if it already exists. Use this when you want to start fresh or change a schema.

### Default Values

Columns can have default values using the `default` keyword:

```cozoscript
:create article {
    id: String
    =>
    title: String,
    status: String default 'draft',
    view_count: Int default 0,
    created_at: Float default now()
}
```

When inserting data, if a column with a default value is not provided, the default expression is evaluated.

### Explicit Column-to-Binding Correspondence

When creating a relation as part of a query, you specify which query columns map to which relation columns:

```cozoscript
?[article_id, article_title] <- [
    ['a1', 'First Post'],
    ['a2', 'Second Post']
]

:create article {
    id: String = article_id
    =>
    title: String = article_title,
    status: String default 'draft'
}
```

Here `id` gets its value from the `article_id` column, and `title` from `article_title`. The `status` column uses its default value since it has no explicit binding.

---

## 2.4 Mutating Data

All mutation operations follow the same pattern: a query produces rows, and the system command tells CozoDB what to do with them.

### The Mutation Pattern

```cozoscript
?[col1, col2, col3] <- [[$val1, $val2, $val3]]
:put relation_name {col1 => col2, col3}
```

The query head (`?[col1, col2, col3]`) must contain columns that match the relation schema. The system command (`:put`, `:rm`, etc.) specifies the target relation and its schema mapping.

### `:put` -- Upsert (Insert or Replace)

`:put` inserts a new row, or replaces the value columns if the key already exists.

```cozoscript
# Insert a person (or update if id already exists)
?[id, name, description] <- [['p1', 'Alice', 'Software engineer']]
:put person {id => name, description}
```

This is the most commonly used mutation. It's idempotent -- running it twice with the same data produces the same result.

**Example from the entity extraction project** (from `cozo_demo.py`):

```cozoscript
?[id, name, description, first_mentioned, embedding] <- [[
    $id, $name, $description, $first_mentioned, $embedding
]]
:put person {id => name, description, first_mentioned, embedding}
```

### `:rm` -- Remove by Key

`:rm` removes rows matching the given keys. It silently does nothing if the key doesn't exist.

```cozoscript
# Remove person with id 'p1'
?[id] <- [['p1']]
:rm person {id}
```

Only key columns are needed -- value columns are ignored.

### `:insert` -- Strict Insert

`:insert` inserts a new row. **Errors if the key already exists.**

```cozoscript
# Only succeeds if id 'p2' doesn't exist yet
?[id, name, description] <- [['p2', 'Bob', 'Data scientist']]
:insert person {id => name, description}
```

Use this when you want to enforce that you're creating something new and not accidentally overwriting.

### `:update` -- Strict Update

`:update` updates value columns for an existing row. **Errors if the key doesn't exist.**

```cozoscript
# Only succeeds if id 'p1' already exists
?[id, description] <- [['p1', 'Senior software engineer']]
:update person {id => description}
```

Note: You only need to provide the columns you want to update -- other value columns are left unchanged.

### `:delete` -- Strict Delete

`:delete` removes a row. **Errors if the key doesn't exist.**

```cozoscript
?[id] <- [['p1']]
:delete person {id}
```

Use this when the absence of the row indicates a logic error in your program.

### `:ensure` and `:ensure_not` -- Assertions

These don't mutate data -- they check conditions and error if the assertion fails.

```cozoscript
# Error if person 'p1' does NOT exist
?[id] <- [['p1']]
:ensure person {id}

# Error if person 'p999' DOES exist
?[id] <- [['p999']]
:ensure_not person {id}
```

These are useful for enforcing preconditions before a sequence of operations.

### `:returning` -- Get Back Mutated Rows

Add `:returning` to any mutation to see what was actually changed:

```cozoscript
?[id, name, description] <- [
    ['p1', 'Alice', 'Engineer'],
    ['p2', 'Bob', 'Designer']
]
:put person {id => name, description}
:returning
```

This returns the rows that were actually inserted or updated, which is useful for confirming mutations and for debugging.

### Summary of Mutation Operations

| Command | If Key Exists | If Key Doesn't Exist |
|---|---|---|
| `:put` | Replace values | Insert new row |
| `:rm` | Remove row | Do nothing (silent) |
| `:insert` | **Error** | Insert new row |
| `:update` | Update values | **Error** |
| `:delete` | Remove row | **Error** |
| `:ensure` | OK (no-op) | **Error** |
| `:ensure_not` | **Error** | OK (no-op) |

---

## 2.5 Schema Management

CozoDB provides system commands (prefixed with `::`) for inspecting and managing schema.

### `::relations` -- List All Relations

```cozoscript
::relations
```

Returns a list of all stored relations in the database along with metadata.

### `::columns relation_name` -- Show Schema

```cozoscript
::columns person
```

Returns the column definitions for a relation: column names, types, whether they're keys, and default values.

### `::remove relation_name` -- Drop Relation

```cozoscript
::remove person
```

Permanently deletes the relation and all its data. This cannot be undone.

You can remove multiple relations at once:

```cozoscript
::remove person, relationship, behavior
```

### `::rename old new` -- Rename Relation

```cozoscript
::rename person people
```

Renames a relation without affecting its data or indices.

---

## 2.6 Real-World Schema Design: The Entity Extraction Model

Now let's apply everything we've learned to a real schema. The entity extraction project tracks people, their relationships, behaviors, and events extracted from narrative text. Here's how each relation is designed and why.

> **Reference**: The full schema is defined in `schema_design.md` and implemented in `cozo_demo.py` (lines 20-138).

### Person Relation

```cozoscript
:create person {
    id: String
    =>
    name: String,
    description: String,
    first_mentioned: String,
    embedding: <F32; 384>
}
```

**Design decisions:**

- **`id` is the sole key**: Each person is uniquely identified by a single string ID (e.g., `'sarah_martinez'`). This is a normalized, slug-style identifier.
- **`name` is a value, not a key**: A person's display name (`'Sarah Martinez'`) can change without creating a new entity. It's descriptive, not identifying.
- **`description` is a value**: LLM-generated summary. May be updated as more text is processed.
- **`first_mentioned` is a value**: Tracks when this person first appeared in the corpus. It's a String rather than a timestamp because the extraction system works with text-based dates.
- **`embedding` is a value** of type `<F32; 384>`: A 384-dimensional vector from a sentence-transformer model (`all-MiniLM-L6-v2`). Enables semantic search for "find people similar to this description."

### Relationship Relation

```cozoscript
:create relationship {
    id: String,
    from_person: String,
    to_person: String,
    timestamp: String
    =>
    relationship_type: String,
    description: String,
    sentiment: String,
    strength: Float,
    embedding: <F32; 384>
}
```

**Design decisions:**

- **Four key columns** form a composite key: `(id, from_person, to_person, timestamp)`. This is the most interesting design choice in the schema.
- **Why is `timestamp` a key?** Because the same relationship between two people can be observed at different points in time with different characteristics. Sarah and David might be "colleagues" in January and "close friends" by June. Each observation is a distinct row.
- **Why is `id` a key?** It provides a unique identifier for each relationship instance, allowing multiple relationship observations in the same timestamp.
- **`from_person` and `to_person` are keys**: Together with `id` and `timestamp`, they enable efficient lookup of all relationships between two specific people (prefix-key scan on `id, from_person, to_person`).
- **`relationship_type` is a value**: Categories like `'friend'`, `'colleague'`, `'mentor'`. It's a value because the same key-identified relationship observation has exactly one type.
- **`sentiment` and `strength` are values**: Descriptive attributes of the observation.
- **`embedding` is a value**: Enables "find relationships similar to this one."

**Key ordering implications**: Because keys are stored in lexicographic order `(id, from_person, to_person, timestamp)`, these lookups are efficient:
- All relationships for a specific `id`
- All relationships from a specific person (prefix scan on `id, from_person`)
- All relationships between two specific people (prefix scan on `id, from_person, to_person`)

### Behavior Relation

```cozoscript
:create behavior {
    id: String,
    person_id: String,
    timestamp: String
    =>
    action: String,
    description: String,
    location: String?,
    involves_persons: String?,
    embedding: <F32; 384>
}
```

**Design decisions:**

- **`(id, person_id, timestamp)` composite key**: Uniquely identifies a behavior observation. The combination allows tracking the same person's actions over time.
- **`location` and `involves_persons` are nullable** (`String?`): Not every behavior has a known location or involves other people. The `?` prevents errors when inserting rows without these values.
- **`involves_persons` is `String?` instead of `List[String]?`**: In the actual implementation (`cozo_demo.py` line 92), the list is serialized to JSON via `json.dumps()`. This is a pragmatic choice -- CozoDB stores it as a string, and the application deserializes it when needed.

### Event Relation

```cozoscript
:create event {
    id: String,
    timestamp: String
    =>
    title: String,
    description: String,
    participants: String,
    location: String?,
    embedding: <F32; 384>
}
```

**Design decisions:**

- **`(id, timestamp)` composite key**: An event is identified by its ID and when it occurred.
- **`participants` is `String` (not `List[String]`)**: Similar to `involves_persons` above, the participant list is serialized to JSON at the application level (`cozo_demo.py` line 135).
- **`location` is nullable**: Not all events have a known location.

### Schema Design Principles Observed

1. **Key columns define identity and sort order.** Put columns you'll filter on most into the key.
2. **Temporal columns in keys enable time-series tracking.** Including `timestamp` in the key allows multiple observations of the same entity over time.
3. **Nullable types for optional data.** Use `String?`, `Float?`, etc. when data might be absent.
4. **Embeddings as values.** Vector embeddings are descriptive (computed from the row's content), not identifying.
5. **Pragmatic serialization.** Lists of IDs are often stored as JSON strings rather than native `List[T]` for simpler interop with host languages.

---

## 2.7 Indexing

By default, CozoDB can efficiently look up rows by key prefix. But what if you need to query by a value column, or by a non-prefix subset of key columns? That's where indices come in.

### Creating Standard Indices

```cozoscript
::index create relation_name:index_name {col_a, col_b}
```

This creates an index named `index_name` on `relation_name`, covering columns `col_a` and `col_b`.

**Example**: Index the `relationship` relation by `relationship_type`:

```cozoscript
::index create relationship:rel_type_idx {relationship_type}
```

Now you can quickly find all relationships of a given type without scanning the entire relation.

### How Indices Work

An index in CozoDB is essentially a **column reordering**. It creates a secondary copy of the data sorted by the indexed columns. This means:

- Index lookups are fast (prefix scans on the reordered columns)
- Indices cost extra storage space
- Writes are slower (both the main relation and the index must be updated)

### Querying Indices Directly

You can query an index directly using the `*relation:index_name{...}` syntax:

```cozoscript
# Query the rel_type_idx index directly
?[from_person, to_person, description] :=
    *relationship:rel_type_idx{
        relationship_type: 'mentor',
        from_person,
        to_person,
        description
    }
```

### Automatic Index Usage

CozoDB's query planner can automatically use an index when it determines that scanning the index is more efficient than scanning the main relation. However, you can also force index usage by querying the index directly as shown above.

### Dropping Indices

```cozoscript
::index drop relationship:rel_type_idx
```

This removes the index but leaves the underlying relation intact.

---

## 2.8 HNSW Vector Indices

HNSW (Hierarchical Navigable Small World) indices enable **approximate nearest-neighbor search** over vector columns. This is what powers semantic similarity search in the entity extraction project.

### Creating an HNSW Index

```cozoscript
::hnsw create relation_name:index_name {
    dim: <dimension>,
    m: <max_connections>,
    dtype: F32,
    fields: [<vector_column>],
    distance: <metric>,
    ef_construction: <build_quality>
}
```

**Parameters explained:**

| Parameter | Description | Typical Value |
|---|---|---|
| `dim` | Vector dimension (must match the column's `<F32; N>`) | `384` |
| `m` | Max connections per node in the graph. Higher = better recall, more memory | `16` |
| `dtype` | Data type of vector elements | `F32` |
| `fields` | Which column(s) contain the vectors | `[embedding]` |
| `distance` | Distance metric | `Cosine`, `L2`, `IP` |
| `ef_construction` | Build-time beam width. Higher = better index quality, slower build | `200` |

**Distance metrics:**

| Metric | Description | Use When |
|---|---|---|
| `Cosine` | Cosine distance (1 - cosine similarity) | Text embeddings, normalized vectors |
| `L2` | Euclidean distance | Spatial data, unnormalized vectors |
| `IP` | Inner product (negative, so smaller = more similar) | When vectors are pre-normalized |

### Example from the Entity Extraction Project

The project creates four HNSW indices, one for each relation with embeddings (from `cozo_demo.py` lines 142-198):

```cozoscript
# Person embedding index
::hnsw create person:person_embedding_idx {
    dim: 384,
    m: 16,
    dtype: F32,
    fields: [embedding],
    distance: Cosine,
    ef_construction: 200
}

# Relationship embedding index
::hnsw create relationship:relationship_embedding_idx {
    dim: 384,
    m: 16,
    dtype: F32,
    fields: [embedding],
    distance: Cosine,
    ef_construction: 200
}

# Behavior embedding index
::hnsw create behavior:behavior_embedding_idx {
    dim: 384,
    m: 16,
    dtype: F32,
    fields: [embedding],
    distance: Cosine,
    ef_construction: 200
}

# Event embedding index
::hnsw create event:event_embedding_idx {
    dim: 384,
    m: 16,
    dtype: F32,
    fields: [embedding],
    distance: Cosine,
    ef_construction: 200
}
```

### Vector Search Query Syntax

HNSW indices are queried with the `~` prefix (instead of `*` for stored relations):

```cozoscript
?[dist, name, description] :=
    ~person:person_embedding_idx{
        name, description |
        query: vec($query_embedding),
        k: 5,
        ef: 50,
        bind_distance: dist
    }

:order dist
```

**Query parameters:**

| Parameter | Required | Description |
|---|---|---|
| `query` | Yes | The query vector. Use `vec(list)` to create from a list. |
| `k` | Yes | Number of nearest neighbors to return |
| `ef` | No | Search beam width. Higher = better recall, slower search. Default varies. |
| `bind_distance` | No | Bind the distance to a variable for sorting/filtering |
| `filter` | No | A CozoScript expression to pre-filter candidates |
| `radius` | No | Only return results within this distance |

**The `|` separator**: In the HNSW query syntax, columns before `|` are the output bindings (which columns to return from the relation). Parameters after `|` control the search.

### Filtering During Vector Search

You can filter results during the search itself, which is more efficient than post-filtering:

```cozoscript
# Find people similar to a query, but only those first mentioned in 2023
?[dist, name, description] :=
    ~person:person_embedding_idx{
        name, description, first_mentioned |
        query: vec($query_embedding),
        k: 5,
        ef: 50,
        bind_distance: dist,
        filter: starts_with(first_mentioned, '2023')
    }

:order dist
```

### Radius Search

Instead of returning the top-k nearest neighbors, return all vectors within a given distance:

```cozoscript
?[dist, name, description] :=
    ~person:person_embedding_idx{
        name, description |
        query: vec($query_embedding),
        radius: 0.3,
        ef: 100,
        bind_distance: dist
    }

:order dist
```

### Dropping HNSW Indices

```cozoscript
::hnsw drop person:person_embedding_idx
```

---

## 2.9 Full-Text Search (FTS) Indices

FTS indices allow you to search text content using Boolean queries with tokenization, stemming, and stop-word filtering.

### Creating an FTS Index

```cozoscript
::fts create relation_name:index_name {
    extractor: column_name,
    tokenizer: Tokenizer_Type,
    filters: [Filter1, Filter2, ...]
}
```

### Tokenizers

| Tokenizer | Description | Example Input -> Tokens |
|---|---|---|
| `Raw` | No tokenization -- entire string is one token | `'hello world'` -> `['hello world']` |
| `Simple` | Split on non-alphanumeric characters | `'hello-world'` -> `['hello', 'world']` |
| `Whitespace` | Split on whitespace | `'hello world'` -> `['hello', 'world']` |
| `Ngram` | Character n-grams | Configured with `min_gram` and `max_gram` |

### Filters

Filters are applied in order to the token stream:

| Filter | Description |
|---|---|
| `Lowercase` | Convert tokens to lowercase |
| `Stemmer('English')` | Reduce words to stems (e.g., `'running'` -> `'run'`) |
| `Stopwords('English')` | Remove common words (`'the'`, `'is'`, `'at'`, etc.) |

### Example: FTS Index on Person Descriptions

```cozoscript
::fts create person:person_desc_fts {
    extractor: description,
    tokenizer: Simple,
    filters: [Lowercase, Stemmer('English'), Stopwords('English')]
}
```

### Searching with FTS

FTS indices are queried using the `~` prefix, similar to HNSW:

```cozoscript
?[score, id, name, description] :=
    ~person:person_desc_fts{
        id, name, description |
        query: 'engineer AND software',
        bind_score: score
    }

:order -score
```

### Boolean Query Syntax

FTS queries support Boolean operators:

| Operator | Meaning | Example |
|---|---|---|
| `AND` | Both terms must appear | `'engineer AND software'` |
| `OR` | Either term must appear | `'engineer OR scientist'` |
| `NOT` | Exclude a term | `'engineer NOT junior'` |
| `"..."` | Exact phrase | `'"software engineer"'` |
| `(...)` | Grouping | `'(engineer OR scientist) AND senior'` |

### Dropping FTS Indices

```cozoscript
::fts drop person:person_desc_fts
```

---

## 2.10 MinHash-LSH Indices

MinHash-LSH (Locality-Sensitive Hashing) indices are designed for **near-duplicate detection**. They estimate the Jaccard similarity between sets of tokens, which is useful for finding documents or descriptions that are "almost the same."

### When to Use MinHash-LSH

- Finding near-duplicate descriptions across entities
- Deduplicating extracted text
- Grouping similar records that may have slight wording differences

### Creating a MinHash-LSH Index

```cozoscript
::lsh create relation_name:index_name {
    extractor: column_name,
    tokenizer: Simple,
    filters: [Lowercase],
    n_perm: 200,
    target_threshold: 0.5,
    n_gram: 3,
    false_positive_weight: 1.0,
    false_negative_weight: 1.0
}
```

**Parameters:**

| Parameter | Description |
|---|---|
| `extractor` | The text column to index |
| `tokenizer` | How to tokenize the text |
| `filters` | Token filters (same as FTS) |
| `n_perm` | Number of permutation functions (higher = more accurate, more memory) |
| `target_threshold` | Jaccard similarity threshold for candidates |
| `n_gram` | N-gram size for shingling |
| `false_positive_weight` | Weight for false positives in threshold tuning |
| `false_negative_weight` | Weight for false negatives in threshold tuning |

### Querying MinHash-LSH

```cozoscript
?[id, name, description] :=
    ~person:person_lsh_idx{
        id, name, description |
        query: 'experienced software engineer with expertise in AI',
        k: 5
    }
```

### Dropping MinHash-LSH Indices

```cozoscript
::lsh drop person:person_lsh_idx
```

---

## Exercises

### Exercise 2.1: Create a Simple Relation and Insert Data

Create a `book` relation with `isbn` as the key and `title`, `author`, and `year` as values. Insert three books, then query to see all of them.

<details>
<summary>Solution</summary>

```cozoscript
# Step 1: Create the relation
:create book {
    isbn: String
    =>
    title: String,
    author: String,
    year: Int
}
```

```cozoscript
# Step 2: Insert data
?[isbn, title, author, year] <- [
    ['978-0-06-112008-4', 'To Kill a Mockingbird', 'Harper Lee', 1960],
    ['978-0-452-28423-4', '1984', 'George Orwell', 1949],
    ['978-0-7432-7356-5', 'The Great Gatsby', 'F. Scott Fitzgerald', 1925]
]
:put book {isbn => title, author, year}
```

```cozoscript
# Step 3: Query
?[isbn, title, author, year] := *book{isbn, title, author, year}
:order year
```
</details>

### Exercise 2.2: Design a Library Schema

Design a schema for a library system with three relations: `book`, `patron` (library member), and `checkout` (borrowing records). Think carefully about which columns should be keys vs values.

Requirements:
- Books have ISBN, title, author, genre, and publication year
- Patrons have an ID, name, email, and membership date
- Checkouts track which patron borrowed which book and when, plus when it was returned (nullable)

<details>
<summary>Solution</summary>

```cozoscript
:create book {
    isbn: String
    =>
    title: String,
    author: String,
    genre: String,
    pub_year: Int
}

:create patron {
    id: String
    =>
    name: String,
    email: String,
    member_since: String
}

:create checkout {
    patron_id: String,
    isbn: String,
    checkout_date: String
    =>
    return_date: String?
}
```

**Why this key design?**

- `book` keyed on `isbn`: One row per book, globally unique identifier.
- `patron` keyed on `id`: One row per library member.
- `checkout` keyed on `(patron_id, isbn, checkout_date)`: The same patron can borrow the same book multiple times, so we need the date in the key. `return_date` is nullable because the book might still be checked out.

The composite key on `checkout` means:
- "Show all checkouts by patron X" is a fast prefix scan on `patron_id`
- "Show all checkouts of book Y by patron X" is a prefix scan on `(patron_id, isbn)`
</details>

### Exercise 2.3: `:put` vs `:insert` Behavior

Demonstrate the difference between `:put` and `:insert`. First, create a `config` relation. Insert a key-value pair with `:insert`, then try to insert the same key again with `:insert` (observe the error), then use `:put` (observe the success).

<details>
<summary>Solution</summary>

```cozoscript
# Create a config relation
:create config {
    key: String
    =>
    value: String
}
```

```cozoscript
# First insert succeeds
?[key, value] <- [['theme', 'dark']]
:insert config {key => value}
```

```cozoscript
# Second insert with same key FAILS with an error
?[key, value] <- [['theme', 'light']]
:insert config {key => value}
# ERROR: key already exists
```

```cozoscript
# :put with same key succeeds (updates the value)
?[key, value] <- [['theme', 'light']]
:put config {key => value}
```

```cozoscript
# Verify the value was updated
?[key, value] := *config{key, value}
# Returns: theme -> light
```
</details>

### Exercise 2.4: Create a Vector Index and Do a Similarity Search

Create a `product` relation with a name, category, and a small 4-dimensional embedding vector. Insert five products with hand-crafted vectors. Create an HNSW index and search for the product most similar to a query vector.

<details>
<summary>Solution</summary>

```cozoscript
# Create the relation
:create product {
    id: String
    =>
    name: String,
    category: String,
    embedding: <F32; 4>
}
```

```cozoscript
# Insert products with 4-dimensional vectors
?[id, name, category, embedding] <- [
    ['p1', 'Laptop',       'electronics', [0.9, 0.1, 0.0, 0.2]],
    ['p2', 'Smartphone',   'electronics', [0.8, 0.2, 0.0, 0.3]],
    ['p3', 'Running Shoes', 'sports',     [0.1, 0.8, 0.9, 0.0]],
    ['p4', 'Tennis Racket', 'sports',      [0.0, 0.7, 0.8, 0.1]],
    ['p5', 'Headphones',   'electronics', [0.7, 0.3, 0.1, 0.4]]
]
:put product {id => name, category, embedding}
```

```cozoscript
# Create the HNSW index
::hnsw create product:product_emb_idx {
    dim: 4,
    m: 8,
    dtype: F32,
    fields: [embedding],
    distance: Cosine,
    ef_construction: 50
}
```

```cozoscript
# Search for products similar to "electronics-like" vector
?[dist, name, category] :=
    ~product:product_emb_idx{
        name, category |
        query: vec([0.85, 0.15, 0.05, 0.25]),
        k: 3,
        ef: 20,
        bind_distance: dist
    }

:order dist
```

The results should show the three electronics products (Laptop, Smartphone, Headphones) as closest to the query vector.
</details>

### Exercise 2.5: Standard Index and Direct Querying

Using the `checkout` relation from Exercise 2.2, create a standard index on `isbn` (so we can efficiently look up all checkouts for a specific book without knowing the patron). Insert some sample data and query through the index.

<details>
<summary>Solution</summary>

```cozoscript
# Ensure the checkout relation exists (from Exercise 2.2)
:create checkout {
    patron_id: String,
    isbn: String,
    checkout_date: String
    =>
    return_date: String?
}
```

```cozoscript
# Insert sample checkout data
?[patron_id, isbn, checkout_date, return_date] <- [
    ['pat1', '978-0-06-112008-4', '2026-01-15', '2026-02-01'],
    ['pat2', '978-0-06-112008-4', '2026-02-05', null],
    ['pat1', '978-0-452-28423-4', '2026-01-20', '2026-02-10'],
    ['pat3', '978-0-7432-7356-5', '2026-02-01', null]
]
:put checkout {patron_id, isbn, checkout_date => return_date}
```

```cozoscript
# Create index on isbn for book-centric lookups
::index create checkout:by_isbn {isbn, patron_id, checkout_date}
```

```cozoscript
# Query through the index: find all checkouts of a specific book
?[patron_id, checkout_date, return_date] :=
    *checkout:by_isbn{
        isbn: '978-0-06-112008-4',
        patron_id,
        checkout_date,
        return_date
    }

:order checkout_date
```

Without the index, CozoDB would need to scan all checkouts. With the `by_isbn` index, it can jump directly to the rows for the requested ISBN.
</details>

### Exercise 2.6: Inspect Schema with System Commands

Using a database with the entity extraction schema (person, relationship, behavior, event), use `::relations` and `::columns` to inspect the schema. Write down what you observe about the key/value structure.

<details>
<summary>Solution</summary>

```cozoscript
# List all relations
::relations
```

This shows all four relations: `person`, `relationship`, `behavior`, `event`.

```cozoscript
# Inspect person schema
::columns person
```

Expected output columns:
- `id` (String, key)
- `name` (String, value)
- `description` (String, value)
- `first_mentioned` (String, value)
- `embedding` (<F32; 384>, value)

```cozoscript
# Inspect relationship schema
::columns relationship
```

Expected output columns:
- `id` (String, key)
- `from_person` (String, key)
- `to_person` (String, key)
- `timestamp` (String, key)
- `relationship_type` (String, value)
- `description` (String, value)
- `sentiment` (String, value)
- `strength` (Float, value)
- `embedding` (<F32; 384>, value)

```cozoscript
# Inspect behavior schema
::columns behavior
```

Notice that `location` and `involves_persons` are nullable (`String?`).

```cozoscript
# Inspect event schema
::columns event
```

Notice the `(id, timestamp)` composite key.
</details>

### Exercise 2.7: Schema Design Challenge -- Social Network

Design a schema for a simple social network with:
- Users (profile info)
- Posts (text content with timestamps)
- Follows (who follows whom)
- Likes (who liked which post, when)

Think about: What should be keys vs values? Which columns should be nullable? What composite keys make sense for common query patterns?

<details>
<summary>Solution</summary>

```cozoscript
:create user {
    username: String
    =>
    display_name: String,
    bio: String?,
    joined: String
}

:create post {
    id: String
    =>
    author: String,
    content: String,
    created_at: String,
    edited_at: String?
}

:create follows {
    follower: String,
    followed: String
}

:create likes {
    user: String,
    post_id: String
    =>
    liked_at: String
}
```

**Design rationale:**

- **`user`**: Keyed on `username` (unique handle). Display name, bio (nullable -- not everyone fills it in), and join date are descriptive values.

- **`post`**: Keyed on `id` (UUID). The `author` is a value because it describes the post, not identifies it. `edited_at` is nullable because most posts aren't edited.

- **`follows`**: Both columns are keys, no `=>` needed. This is a pure set -- a user either follows someone or doesn't. The composite key `(follower, followed)` means "show everyone that user X follows" is a fast prefix scan. If you also need "show all followers of user Y," you'd add an index: `::index create follows:by_followed {followed, follower}`.

- **`likes`**: Keyed on `(user, post_id)` because a user can only like a post once. `liked_at` is a value -- it describes when the like happened. The key ordering means "show all posts liked by user X" is efficient by default.
</details>

### Exercise 2.8: Using `:returning` to See Mutations

Create a `counter` relation, insert some data, update it, and delete it -- using `:returning` each time to see what was affected.

<details>
<summary>Solution</summary>

```cozoscript
:create counter {
    name: String
    =>
    value: Int
}
```

```cozoscript
# Insert with :returning
?[name, value] <- [
    ['page_views', 0],
    ['api_calls', 0],
    ['errors', 0]
]
:put counter {name => value}
:returning
```

Returns all three inserted rows.

```cozoscript
# Update one counter
?[name, value] <- [['page_views', 42]]
:put counter {name => value}
:returning
```

Returns only the updated row: `page_views -> 42`.

```cozoscript
# Delete a counter
?[name] <- [['errors']]
:rm counter {name}
:returning
```

Returns the deleted row: `errors -> 0`.

```cozoscript
# Verify final state
?[name, value] := *counter{name, value}
:order name
```

Should show `api_calls -> 0` and `page_views -> 42`.
</details>

### Exercise 2.9: FTS Index on the Library Schema

Create an FTS index on the `book` relation's `title` column. Insert several books and search for books with "great" in the title.

<details>
<summary>Solution</summary>

```cozoscript
# Ensure the book relation exists with data
:replace book {
    isbn: String
    =>
    title: String,
    author: String,
    genre: String,
    pub_year: Int
}
```

```cozoscript
?[isbn, title, author, genre, pub_year] <- [
    ['978-0-7432-7356-5', 'The Great Gatsby', 'F. Scott Fitzgerald', 'fiction', 1925],
    ['978-0-06-112008-4', 'To Kill a Mockingbird', 'Harper Lee', 'fiction', 1960],
    ['978-0-14-028329-7', 'Great Expectations', 'Charles Dickens', 'fiction', 1861],
    ['978-0-452-28423-4', '1984', 'George Orwell', 'fiction', 1949],
    ['978-0-06-093546-7', 'To Kill a Mockingbird', 'Harper Lee', 'fiction', 1960],
    ['978-0-316-76948-0', 'The Catcher in the Rye', 'J.D. Salinger', 'fiction', 1951]
]
:put book {isbn => title, author, genre, pub_year}
```

```cozoscript
# Create FTS index on title
::fts create book:title_fts {
    extractor: title,
    tokenizer: Simple,
    filters: [Lowercase]
}
```

```cozoscript
# Search for books with "great" in the title
?[score, isbn, title, author] :=
    ~book:title_fts{
        isbn, title, author |
        query: 'great',
        bind_score: score
    }

:order -score
```

Should return "The Great Gatsby" and "Great Expectations".
</details>

### Exercise 2.10: Complete Entity Pipeline

This exercise ties together everything in the module. Create a simplified version of the entity extraction schema with just `person` and `relationship`, insert sample data with embeddings (use small 4-dimensional vectors for simplicity), create HNSW indices, and then:

1. Query all people
2. Query relationships for a specific person
3. Do a vector similarity search

<details>
<summary>Solution</summary>

```cozoscript
# Create the relations
:create person {
    id: String
    =>
    name: String,
    description: String,
    embedding: <F32; 4>
}

:create relationship {
    id: String,
    from_person: String,
    to_person: String
    =>
    relationship_type: String,
    description: String,
    embedding: <F32; 4>
}
```

```cozoscript
# Insert people
?[id, name, description, embedding] <- [
    ['alice', 'Alice Chen', 'Senior engineer who leads the backend team', [0.9, 0.3, 0.1, 0.5]],
    ['bob', 'Bob Smith', 'Junior developer working on frontend', [0.8, 0.4, 0.2, 0.3]],
    ['carol', 'Carol Davis', 'Product manager coordinating releases', [0.2, 0.9, 0.7, 0.1]],
    ['dave', 'Dave Wilson', 'DevOps engineer maintaining infrastructure', [0.7, 0.2, 0.3, 0.8]]
]
:put person {id => name, description, embedding}
```

```cozoscript
# Insert relationships
?[id, from_person, to_person, relationship_type, description, embedding] <- [
    ['r1', 'alice', 'bob', 'mentor', 'Alice mentors Bob on backend development', [0.8, 0.3, 0.2, 0.4]],
    ['r2', 'carol', 'alice', 'collaborator', 'Carol and Alice plan feature releases together', [0.5, 0.7, 0.6, 0.3]],
    ['r3', 'alice', 'dave', 'colleague', 'Alice and Dave coordinate on deployments', [0.7, 0.2, 0.4, 0.7]],
    ['r4', 'bob', 'carol', 'reports_to', 'Bob provides frontend updates to Carol', [0.3, 0.8, 0.5, 0.2]]
]
:put relationship {id, from_person, to_person => relationship_type, description, embedding}
```

```cozoscript
# Create HNSW indices
::hnsw create person:person_emb_idx {
    dim: 4, m: 8, dtype: F32, fields: [embedding],
    distance: Cosine, ef_construction: 50
}

::hnsw create relationship:rel_emb_idx {
    dim: 4, m: 8, dtype: F32, fields: [embedding],
    distance: Cosine, ef_construction: 50
}
```

```cozoscript
# 1. Query all people
?[id, name, description] := *person{id, name, description}
:order name
```

```cozoscript
# 2. Query relationships for Alice
?[to_person, relationship_type, description] :=
    *relationship{from_person: 'alice', to_person, relationship_type, description}
```

```cozoscript
# 3. Vector similarity search: find people similar to "engineering" vector
?[dist, name, description] :=
    ~person:person_emb_idx{
        name, description |
        query: vec([0.85, 0.25, 0.15, 0.6]),
        k: 3,
        ef: 20,
        bind_distance: dist
    }

:order dist
```

Alice and Dave (the engineering-oriented people) should be closest to the query vector.
</details>

---

## Key Takeaways

1. **Stored relations use `{key_cols => value_cols}` syntax.** Keys define uniqueness and sort order; values are descriptive data.
2. **CozoDB has rich data types** including vectors (`<F32; N>`), nullable types (`T?`), and `Json`.
3. **Six mutation commands** with different strictness: `:put` (upsert), `:rm` (silent remove), `:insert` (strict insert), `:update` (strict update), `:delete` (strict delete), plus `:ensure`/`:ensure_not` for assertions.
4. **All mutations follow the query-then-command pattern**: produce rows with a query, then tell CozoDB what to do with them.
5. **System commands (`::`)** manage schema: `::relations`, `::columns`, `::remove`, `::rename`.
6. **Standard indices** reorder columns for efficient non-key lookups.
7. **HNSW indices** enable approximate nearest-neighbor search over vectors -- the core of semantic search.
8. **FTS indices** enable Boolean text search with tokenization and stemming.
9. **MinHash-LSH indices** find near-duplicates based on set similarity.
10. **Good schema design puts frequently-queried columns in the key** and uses composite keys for temporal/multi-dimensional lookups.

## Next Module

In **Module 3: Advanced Queries and Graph Algorithms**, you'll learn about aggregation, recursion, fixed rules, and CozoDB's built-in graph algorithms like PageRank, shortest path, and community detection.
