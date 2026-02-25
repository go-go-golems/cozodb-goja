---
Title: Module 4 - Time Travel and Temporal Data
Ticket: CO-03
Status: active
Topics:
    - cozodb
    - course
    - entity-extraction
    - datalog
    - javascript
    - goja
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo_demo.py:CozoDB demo with temporal relationship queries"
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo_advanced_demo.py:Advanced demo with sentiment evolution and temporal analysis"
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/demo_text.txt:Source narrative showing relationship evolution over time"
ExternalSources:
    - https://docs.cozodb.org/en/latest/timetravel.html
    - https://docs.cozodb.org/en/latest/stored.html
    - https://docs.cozodb.org/en/latest/queries.html
Summary: "Deep dive into temporal data modeling in CozoDB: the built-in Validity type for time travel queries, application-level temporal modeling with timestamp keys, triggers for automated history tracking, and chained transactions for atomic temporal updates"
LastUpdated: 2026-02-24T18:41:06.214257043-05:00
WhatFor: "Fourth module of the CozoDB course - teaches how to model, store, and query data that changes over time"
WhenToUse: "After completing Modules 1-3; when you need to track evolving relationships, audit history, or query historical states"
---

# Module 4: Time Travel and Temporal Data

## Learning Objectives

By the end of this module you will:
- Understand why temporal data matters and the challenges it presents
- Use CozoDB's built-in `Validity` type for automatic time travel queries
- Model temporal data at the application level using timestamp keys
- Write queries that track relationship evolution over time
- Create triggers for automated temporal tracking
- Build atomic multi-query transactions for safe temporal updates
- Choose the right temporal strategy for your use case

---

## 4.1 Why Temporal Data Matters

Most databases treat data as a snapshot: you insert a row, update it, and the old value is gone forever. But real-world data changes over time, and those changes are often the most interesting part of the story.

### The Problem with Snapshots

Consider the relationship between Sarah Martinez and David Chen in our entity extraction project. If we only store the *current* state, we see: "professional partners, positive sentiment, strength 0.95." But that misses the story entirely:

| When | What Actually Happened |
|---|---|
| Jan 2023 | Sarah joins the lab, meets David -- they become colleagues |
| Mar 2023 | They present at a conference together -- bond strengthens |
| Jun 2023 | Conflict arises over experimental approach -- negative sentiment |
| Aug 2023 | Reconciliation after apology -- relationship rebounds, stronger |
| Dec 2023 | Sarah promoted, advocates for David as co-lead -- deep professional bond |

Without temporal data, you lose the narrative. You cannot answer questions like:
- **"What was their relationship like in June 2023?"** -- impossible, the snapshot only has December's data.
- **"When did the conflict start?"** -- gone, overwritten by the reconciliation.
- **"How many relationships changed sentiment over time?"** -- no history to compare.

### Where Temporal Data Appears

Temporal data shows up everywhere:
- **Social networks**: friendships form, evolve, and dissolve
- **Employment**: roles, titles, and reporting structures change
- **Healthcare**: patient conditions evolve, treatments change
- **Finance**: account balances, credit ratings, and risk profiles shift
- **Knowledge graphs**: facts become outdated, new facts emerge

### Two Fundamental Questions

Any temporal system must answer two questions:

1. **"What did the data look like at time T?"** -- point-in-time queries, sometimes called "as of" queries
2. **"How has this data changed over time?"** -- history queries, audit trails, evolution analysis

CozoDB provides tools for both: a built-in `Validity` type for automatic versioning, and schema patterns for application-level temporal modeling.

---

## 4.2 CozoDB's Time Travel Feature

CozoDB has a first-class feature for temporal data: the `Validity` type. When you add a `Validity` column as the **last key column** of a stored relation, CozoDB automatically manages versioned data and supports "time travel" queries.

### Creating a Time-Travel-Enabled Relation

```cozoscript
:create versioned_person {
    id: String,
    valid_at: Validity
    =>
    name: String,
    description: String
}
```

Key points:
- `valid_at: Validity` is a special column type -- it holds a timestamp plus a flag indicating whether the record is an assertion (the record exists at this time) or a retraction (the record was removed at this time)
- It **must** be the **last key column** (after all other key columns, before the `=>`)
- CozoDB uses this column to track when each version of the record was valid

### How Validity Works Under the Hood

When you write to a `Validity`-enabled relation, CozoDB does not overwrite previous versions. Instead, it appends a new version with a new timestamp. The data store accumulates a history:

```
id: "sarah_martinez", valid_at: T1 => name: "Sarah Martinez", description: "Junior scientist"
id: "sarah_martinez", valid_at: T2 => name: "Sarah Martinez", description: "Senior scientist"
id: "sarah_martinez", valid_at: T3 => name: "Sarah Martinez", description: "Lead scientist"
```

All three versions coexist in storage. CozoDB determines which version is "current" by finding the latest assertion that is not followed by a retraction.

### Inserting Versioned Data

Use `:put` as normal, but you can optionally specify the validity timestamp:

```cozoscript
# Insert with the current timestamp (CozoDB uses "now")
?[id, name, description] <- [['sarah', 'Sarah Martinez', 'Junior scientist']]
:put versioned_person {id => name, description}

# Insert with a specific validity timestamp
# The validity value is a pair: [timestamp, is_assertion]
# is_assertion = true means "this record exists as of this time"
?[id, valid_at, name, description] <- [
    ['sarah', [1672531200, true], 'Sarah Martinez', 'Senior scientist']
]
:put versioned_person {id, valid_at => name, description}
```

The `Validity` value is internally a pair: `[unix_timestamp_in_microseconds, is_assertion_boolean]`. When you omit it, CozoDB uses the current time and `true` (assertion).

### Retracting (Deleting) Versioned Data

To mark a record as "deleted" at a point in time without losing history:

```cozoscript
# Retract: mark that sarah's record is no longer valid
?[id, valid_at, name, description] <- [
    ['sarah', [1672617600, false], '', '']
]
:put versioned_person {id, valid_at => name, description}
```

The `false` flag makes this a retraction. The previous versions still exist -- you just recorded that "as of timestamp 1672617600, this record is no longer current."

---

## 4.3 Querying Historical Data

### Default: Latest Version

A normal query against a `Validity`-enabled relation returns only the latest version of each record:

```cozoscript
# Returns only the most recent version of each person
?[id, name, description] := *versioned_person{id, name, description}
```

This is the same syntax you use for any stored relation. CozoDB automatically resolves to the latest version.

### Time Travel with `@`

To query data **as it existed at a specific point in time**, use the `@` operator:

```cozoscript
# What did the data look like at timestamp T?
?[id, name, description] := *versioned_person @ "2023-01-15T00:00:00" {id, name, description}
```

The `@` syntax tells CozoDB: "Find the version of each record that was current at this timestamp." If a record had been retracted before that timestamp, it does not appear.

The timestamp format accepted by `@` is an ISO 8601 string or a Unix timestamp.

### Listing All Versions

To see every version of a record (the full history), you explicitly bind the `valid_at` column:

```cozoscript
# Show all versions of sarah's record
?[id, valid_at, name, description] :=
    *versioned_person{id: 'sarah', valid_at, name, description}
:order valid_at
```

When you bind `valid_at` explicitly, CozoDB does not filter to the latest version -- it returns every version stored.

### Combining Time Travel with Filtering

```cozoscript
# Find all people whose description mentioned "scientist" as of June 2023
?[id, name, description] :=
    *versioned_person @ "2023-06-01T00:00:00" {id, name, description},
    contains(description, 'scientist')
```

---

## 4.4 Temporal Patterns in This Project

The entity extraction project takes a different approach to temporal data. Instead of using CozoDB's `Validity` type, it models time **at the application level** by making the timestamp part of the composite key.

### The Relationship Schema

Recall the relationship relation from `cozo_demo.py`:

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

The key columns are `{id, from_person, to_person, timestamp}`. Because `timestamp` is part of the key, the **same pair of people** can have **multiple relationship records** at different points in time. Each record captures the state of the relationship at that moment.

This is fundamentally different from a typical relational design where you would have `{from_person, to_person}` as the key and just overwrite the row when the relationship changes.

### Walking Through the Demo Data: Sarah and David

The source narrative (`demo_text.txt`) describes a rich, evolving relationship between Sarah Martinez and David Chen. The entity extraction system captures this as multiple relationship records:

```
Timeline: Sarah Martinez <-> David Chen
============================================================

Jan 2023 - "New Colleagues"
  type: colleague
  sentiment: positive
  strength: 0.5
  Sarah joins the lab, meets David who has been there since 2021.
  A new professional connection forms.

Mar 2023 - "Conference Partners"
  type: collaborator
  sentiment: positive
  strength: 0.7
  They present together at a Boston conference.
  Working together strengthens the bond.

Jun 2023 - "The Conflict"
  type: colleague (strained)
  sentiment: negative
  strength: 0.3
  Heated argument about experimental approach.
  Witnessed by Michael Thompson. Tension persists for weeks.

Aug 2023 - "Reconciliation"
  type: collaborator
  sentiment: positive
  strength: 0.8
  Sarah apologizes, they reconcile.
  Elena facilitates a team-building retreat.
  They realize different perspectives are complementary.

Dec 2023 - "Professional Partnership"
  type: professional_partner
  sentiment: positive
  strength: 0.95
  Sarah promoted to lead scientist.
  Advocates for David as co-lead.
  Deep mutual respect and professional bond.
```

This timeline tells a story that a single snapshot could never capture: initial connection, growing trust, conflict, repair, and eventual deep partnership.

### The Behavior Schema: Another Temporal Pattern

Behaviors also use timestamp as part of the key:

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

This allows tracking what a person did and when. Combined with relationship data, you can correlate behaviors with relationship changes -- for example, did the conference presentation (behavior) in March correspond with the strengthening of Sarah and David's bond (relationship)?

---

## 4.5 Temporal Queries on the Project Data

Now let's write real queries against the project's temporal data. These queries work against the schema established in the demo.

### Track Relationship Evolution

The most basic temporal query: show how a specific relationship changed over time.

```cozoscript
?[timestamp, relationship_type, sentiment, strength] :=
    *relationship{
        from_person: 'sarah_martinez',
        to_person: 'david_chen',
        timestamp, relationship_type, sentiment, strength
    }
:order timestamp
```

Expected output:

| timestamp | relationship_type | sentiment | strength |
|---|---|---|---|
| 2023-01 | colleague | positive | 0.5 |
| 2023-03 | collaborator | positive | 0.7 |
| 2023-06 | colleague | negative | 0.3 |
| 2023-08 | collaborator | positive | 0.8 |
| 2023-12 | professional_partner | positive | 0.95 |

This single query reveals the entire arc: formation, growth, rupture, repair, deepening.

### Find Relationships by Time Period

Filter to a specific year:

```cozoscript
?[from_person, to_person, relationship_type, timestamp] :=
    *relationship{from_person, to_person, relationship_type, timestamp},
    starts_with(timestamp, '2023')
:order timestamp
```

Because the timestamps are ISO 8601 strings (or at least lexicographically sortable strings), `starts_with` serves as a convenient prefix filter. All records from 2023 appear; 2024 records do not.

### Early vs. Late Relationships

Compare the first half and second half of the year:

```cozoscript
# Relationships before the conflict (first half of 2023)
?[from_person, to_person, relationship_type, sentiment, timestamp] :=
    *relationship{from_person, to_person, relationship_type, sentiment, timestamp},
    timestamp < '2023-06'
:order timestamp
```

```cozoscript
# Relationships after the reconciliation (second half of 2023)
?[from_person, to_person, relationship_type, sentiment, timestamp] :=
    *relationship{from_person, to_person, relationship_type, sentiment, timestamp},
    timestamp >= '2023-06'
:order timestamp
```

String comparison works here because the timestamps are formatted as `YYYY-MM` or `YYYY-MM-DD`, which sorts lexicographically in chronological order. This is a practical benefit of using ISO 8601 formatted strings.

### Sentiment Distribution by Half-Year

Use aggregation to compare sentiment counts:

```cozoscript
# Count sentiments in the first half of 2023
?[sentiment, count(id)] :=
    *relationship{id, timestamp, sentiment},
    timestamp < '2023-07'
```

```cozoscript
# Count sentiments in the second half of 2023
?[sentiment, count(id)] :=
    *relationship{id, timestamp, sentiment},
    timestamp >= '2023-07'
```

For a combined view using a computed column:

```cozoscript
?[half, sentiment, count(id)] :=
    *relationship{id, timestamp, sentiment},
    starts_with(timestamp, '2023'),
    half = if(timestamp < '2023-07', 'H1', 'H2')
```

### Find the Busiest Month

Which month saw the most new relationships?

```cozoscript
?[month, count(id)] :=
    *relationship{id, timestamp},
    month = timestamp
:order -count(id)
:limit 5
```

If timestamps are full ISO dates, you can extract the month prefix:

```cozoscript
?[month, count(id)] :=
    *relationship{id, timestamp},
    # Use substr or slice to extract YYYY-MM
    month = substr(timestamp, 0, 7)
:order -count(id)
```

### Detect Sentiment Changes

Find pairs of people whose relationship sentiment changed over time:

```cozoscript
# Find relationships that went from positive to negative
?[from_person, to_person, early_ts, late_ts] :=
    *relationship{from_person, to_person, timestamp: early_ts, sentiment: 'positive'},
    *relationship{from_person, to_person, timestamp: late_ts, sentiment: 'negative'},
    early_ts < late_ts
```

```cozoscript
# Find relationships that recovered (negative -> positive)
?[from_person, to_person, conflict_ts, recovery_ts] :=
    *relationship{from_person, to_person, timestamp: conflict_ts, sentiment: 'negative'},
    *relationship{from_person, to_person, timestamp: recovery_ts, sentiment: 'positive'},
    conflict_ts < recovery_ts
```

### Build a Unified Timeline

Combine events from multiple relations into a single chronological view:

```cozoscript
# Relationship events
timeline[timestamp, source, person, detail] :=
    *relationship{from_person: person, to_person: other, timestamp,
                  relationship_type: rtype, sentiment},
    source = 'relationship',
    detail = concat(person, ' -> ', other, ': ', rtype, ' (', sentiment, ')')

# Behavior events
timeline[timestamp, source, person, detail] :=
    *behavior{person_id: person, timestamp, action, description},
    source = 'behavior',
    detail = concat(person, ': ', action, ' - ', description)

# Event events
timeline[timestamp, source, person, detail] :=
    *event{timestamp, title, description},
    source = 'event',
    person = 'all',
    detail = concat(title, ' - ', description)

?[timestamp, source, person, detail] := timeline[timestamp, source, person, detail]
:order timestamp
```

This uses multiple rule definitions (union/disjunction) to merge three temporal streams into one timeline, then sorts chronologically.

---

## 4.6 Application-Level Temporal Modeling

The entity extraction project uses application-level timestamps rather than CozoDB's built-in `Validity` type. Both approaches are valid. Here is how to choose.

### Option A: CozoDB's `Validity` Type

**How it works:**
```cozoscript
:create versioned_relationship {
    from_person: String,
    to_person: String,
    valid_at: Validity
    =>
    relationship_type: String,
    sentiment: String,
    strength: Float
}
```

**Advantages:**
- Automatic version management -- CozoDB handles the timestamps
- Built-in "as of" queries with the `@` operator
- Retraction support -- you can record that a relationship ended without deleting history
- No manual bookkeeping of versions

**Disadvantages:**
- The `Validity` timestamp is the *transaction time* (when the data was written), not necessarily the *valid time* (when the event actually happened in the real world)
- Less flexibility in timestamp format -- you work with Unix microsecond timestamps, not human-readable strings
- Cannot easily rewrite history or insert backdated records
- Adds complexity to the data model

**Best for:** Audit trails, system-managed versioning, scenarios where "when was this written to the database" is the relevant time dimension.

### Option B: Application Timestamps in Keys

**How it works:**
```cozoscript
:create relationship {
    id: String,
    from_person: String,
    to_person: String,
    timestamp: String
    =>
    relationship_type: String,
    sentiment: String,
    strength: Float
}
```

**Advantages:**
- Full control over timestamp format and semantics
- Can insert records for any time, past or future
- Simpler mental model -- timestamps are just another column
- Natural for event sourcing patterns
- Easy to query with string functions (`starts_with`, `<`, `>`)

**Disadvantages:**
- No built-in "as of" queries -- you must write the logic yourself
- No automatic retraction support -- deletion is manual
- Must ensure timestamp format is consistent and sortable
- Must handle "latest version" queries manually (e.g., using `max`)

**Best for:** Event-sourced data, scenarios where the "real-world time" of an event matters more than when it was recorded, flexible temporal modeling.

### Design Pattern: Event Sourcing

The project's approach is essentially **event sourcing**: each row in the `relationship` relation represents an *event* ("at this time, this relationship had these properties"), not a mutable *state*. The current state is derived by looking at the most recent event.

```cozoscript
# Get the "current" (latest) state of each relationship
latest[from_person, to_person, max(timestamp)] :=
    *relationship{from_person, to_person, timestamp}

?[from_person, to_person, relationship_type, sentiment, strength] :=
    latest[from_person, to_person, latest_ts],
    *relationship{from_person, to_person, timestamp: latest_ts,
                  relationship_type, sentiment, strength}
```

### Design Pattern: Slowly Changing Dimensions

In data warehousing, a "slowly changing dimension" (SCD) tracks how an entity's attributes change over time. The project's person relation uses a simpler model (no timestamp in the key), but you could extend it:

```cozoscript
:create person_history {
    id: String,
    effective_from: String
    =>
    name: String,
    description: String,
    role: String
}
```

Each record says "from this date forward, this person had these attributes." To find the attributes at a given time:

```cozoscript
# Find person attributes as of a target date
?[id, name, description, role] :=
    *person_history{id: $target_id, effective_from, name, description, role},
    effective_from <= $target_date,
    # Ensure no later record exists before the target date
    not *person_history{id: $target_id, effective_from: later},
        later > effective_from,
        later <= $target_date
```

### Design Pattern: Bitemporal Data

The most rigorous temporal approach tracks two time dimensions:
- **Valid time**: when the fact was true in the real world
- **Transaction time**: when the fact was recorded in the database

```cozoscript
:create bitemporal_relationship {
    from_person: String,
    to_person: String,
    valid_from: String,
    recorded_at: Validity
    =>
    relationship_type: String,
    sentiment: String,
    strength: Float
}
```

This is powerful but complex. Use it when you need to answer questions like "What did we *think* the relationship was in March, based on what we *knew* at that time?" -- critical for financial auditing, legal systems, and regulatory compliance.

---

## 4.7 Triggers for Temporal Automation

CozoDB supports **triggers** on stored relations. Triggers fire automatically when data is inserted, updated, or deleted, making them ideal for automated temporal tracking.

### Trigger Basics

Set triggers with the `::set_triggers` system command:

```cozoscript
::set_triggers <relation_name>

on put { <query> }
on rm { <query> }
on replace { <query> }
```

- **`on put`**: fires when new rows are inserted or existing rows are updated via `:put`
- **`on rm`**: fires when rows are removed via `:rm`
- **`on replace`**: fires when rows are updated via `:replace` (like `:put` but fails if the row does not already exist)

### The `_new[]` and `_old[]` Implicit Rules

Inside trigger queries, two special implicit rules are available:

- **`_new[col1, col2, ...]`**: contains the rows being written (available in `on put` and `on replace`)
- **`_old[col1, col2, ...]`**: contains the rows that existed before the operation (available in `on put`, `on rm`, and `on replace` -- only populated when the row already existed)

The columns of `_new` and `_old` follow the same order as the relation's schema (keys first, then values).

### Example: Auto-Archive Relationship Changes

First, create a history table:

```cozoscript
:create relationship_history {
    archive_id: String,
    from_person: String,
    to_person: String,
    archived_at: String,
    original_timestamp: String
    =>
    relationship_type: String,
    sentiment: String,
    strength: Float
}
```

Then set a trigger on the relationship relation to automatically archive the old version whenever a relationship is updated:

```cozoscript
::set_triggers relationship

on put {
    ?[archive_id, from_person, to_person, archived_at, original_timestamp,
      relationship_type, sentiment, strength] :=
        _old[id, from_person, to_person, original_timestamp,
             relationship_type, _description, sentiment, strength, _embedding],
        archived_at = now(),
        archive_id = concat(id, '_', original_timestamp)

    :put relationship_history {
        archive_id, from_person, to_person, archived_at, original_timestamp
        => relationship_type, sentiment, strength
    }
}
```

Now, whenever you `:put` a row into `relationship` that replaces an existing row (same key), the old version is automatically copied to `relationship_history` before the update takes effect.

### Important Notes on Triggers

1. **Triggers execute within the same transaction** as the triggering operation. If the trigger query fails, the entire operation rolls back.
2. **`_old` is empty for truly new rows** -- it only contains data when a row with the same key already existed.
3. **Column order in `_old` and `_new` must match the relation schema exactly** -- keys first (in order), then values (in order).
4. **Multiple triggers are allowed** -- you can define multiple `on put` blocks, and they all fire.
5. **Triggers can write to other relations** -- but be careful of infinite loops if you set triggers on the target relation too.

### Example: Audit Log Trigger

Create a general-purpose audit log:

```cozoscript
:create audit_log {
    log_id: String,
    relation_name: String,
    operation: String,
    timestamp: String
    =>
    details: String
}

::set_triggers relationship

on put {
    ?[log_id, relation_name, operation, timestamp, details] :=
        _new[id, from_person, to_person, ts, rtype, desc, sent, str, _emb],
        relation_name = 'relationship',
        operation = 'put',
        timestamp = now(),
        log_id = concat('audit_', id, '_', ts),
        details = concat(from_person, ' -> ', to_person, ': ', rtype, ' (', sent, ')')

    :put audit_log {log_id, relation_name, operation, timestamp => details}
}

on rm {
    ?[log_id, relation_name, operation, timestamp, details] :=
        _old[id, from_person, to_person, ts, rtype, desc, sent, str, _emb],
        relation_name = 'relationship',
        operation = 'rm',
        timestamp = now(),
        log_id = concat('audit_rm_', id, '_', ts),
        details = concat('REMOVED: ', from_person, ' -> ', to_person)

    :put audit_log {log_id, relation_name, operation, timestamp => details}
}
```

---

## 4.8 Chaining Queries in Transactions

CozoDB supports **multi-query transactions** using curly braces `{}`. Each block within the transaction runs sequentially, and the entire transaction is atomic -- either all blocks succeed, or none of them take effect.

### Basic Syntax

```cozoscript
{
    # First query in the transaction
    ?[x] <- [[1]]
}
{
    # Second query in the transaction
    ?[y] <- [[2]]
}
```

Each `{}` block is a separate query, but they share the same transaction context. Changes made by earlier blocks are visible to later blocks.

### Ephemeral Relations with `_` Prefix

Relations whose names start with `_` (underscore) are **ephemeral** -- they exist only within the current transaction and are automatically dropped when the transaction completes.

```cozoscript
{
    # Create a temporary staging area
    ?[id, name] := *person{id, name}
    :replace _temp_persons {id: String => name: String}
}
{
    # Use the temporary data in the next block
    ?[name] := *_temp_persons{name}
}
```

This is useful for multi-step temporal operations where you need to read the current state, compute changes, and apply updates atomically.

### Control Flow

CozoDB transactions support control flow directives:

- **`%if { condition } %then { block } %else { block } %end`** -- conditional execution
- **`%loop { block } %end`** -- repeated execution (use `%break` or `%return` to exit)
- **`%return query`** -- exit the transaction and return results
- **`%break`** -- exit the current loop
- **`%continue`** -- skip to the next loop iteration
- **`%ignore_error`** -- placed before a query to suppress errors (useful for idempotent operations like `:create` that fail if the relation already exists)

### Example: Atomic Temporal Update

Here is a practical pattern for atomically archiving the old version of a relationship and inserting a new version:

```cozoscript
{
    # Step 1: Read the current state and archive it
    ?[id, from_person, to_person, timestamp,
      relationship_type, sentiment, strength] :=
        *relationship{id: $id, from_person, to_person, timestamp,
                      relationship_type, sentiment, strength}

    :put relationship_archive {
        id, from_person, to_person, timestamp
        => relationship_type, sentiment, strength
    }
}
{
    # Step 2: Insert the new version
    ?[id, from_person, to_person, timestamp, relationship_type,
      description, sentiment, strength, embedding] <- [[
        $id, $from_person, $to_person, $new_timestamp,
        $new_type, $new_description, $new_sentiment, $new_strength, $new_embedding
    ]]

    :put relationship {
        id, from_person, to_person, timestamp
        => relationship_type, description, sentiment, strength, embedding
    }
}
```

Because this runs as a single transaction:
- If step 2 fails, step 1 is rolled back -- the archive entry is not created
- No other query can see a state where the archive exists but the new version does not
- The operation is atomic and consistent

### Example: Idempotent Schema Setup

Use `%ignore_error` for operations that might fail harmlessly:

```cozoscript
{
    %ignore_error
    :create relationship_archive {
        id: String,
        from_person: String,
        to_person: String,
        timestamp: String
        =>
        relationship_type: String,
        sentiment: String,
        strength: Float
    }
}
{
    # This block runs even if the :create above failed
    # (because the relation already exists)
    ?[count(id)] := *relationship_archive{id}
}
```

### Example: Conditional Temporal Logic

```cozoscript
{
    # Check if the relationship already has a recent entry
    ?[count(id)] :=
        *relationship{id, from_person: $from, to_person: $to, timestamp},
        timestamp > $threshold_date

    :replace _recent_count {count_val: Int}
}
%if {
    ?[count_val] := *_recent_count{count_val}, count_val > 0
}
%then {
    # Update the existing recent entry
    ?[id, from, to, ts, type, desc, sent, str, emb] :=
        *relationship{id, from_person: from, to_person: to, timestamp: ts,
                      relationship_type: type, description: desc,
                      sentiment: sent, strength: str, embedding: emb},
        from == $from, to == $to, ts > $threshold_date

    # ... apply update logic
    :put relationship {id, from_person, to_person, timestamp
                       => relationship_type, description, sentiment, strength, embedding}
}
%else {
    # Insert a brand new temporal entry
    ?[id, from, to, ts, type, desc, sent, str, emb] <- [[
        $new_id, $from, $to, $new_timestamp, $new_type, $new_desc, $new_sent, $new_str, $new_emb
    ]]
    :put relationship {id, from_person, to_person, timestamp
                       => relationship_type, description, sentiment, strength, embedding}
}
%end
```

---

## Exercises

### Exercise 4.1: Track Sarah-David Relationship Evolution

Write a query that returns all relationship records between Sarah Martinez and David Chen, ordered chronologically. Include the timestamp, relationship type, sentiment, and strength.

<details>
<summary>Solution</summary>

```cozoscript
?[timestamp, relationship_type, sentiment, strength, description] :=
    *relationship{
        from_person: 'sarah_martinez',
        to_person: 'david_chen',
        timestamp, relationship_type, sentiment, strength, description
    }
:order timestamp
```

This shows the full arc: colleague (positive) -> collaborator (positive) -> colleague/strained (negative) -> collaborator (positive) -> professional_partner (positive).
</details>

### Exercise 4.2: Find All Relationships Before June 2023

Write a query that returns all relationships with a timestamp earlier than June 2023. Order results by timestamp.

<details>
<summary>Solution</summary>

```cozoscript
?[from_person, to_person, relationship_type, sentiment, timestamp] :=
    *relationship{from_person, to_person, relationship_type, sentiment, timestamp},
    timestamp < '2023-06'
:order timestamp
```

Because timestamps are stored as lexicographically sortable strings (e.g., `'2023-01'`, `'2023-03'`), a simple string comparison with `<` produces the correct chronological filter.
</details>

### Exercise 4.3: Compare Sentiment Distribution in H1 vs H2 of 2023

Write a query that shows how many positive and negative relationships existed in the first half (H1: Jan-Jun) versus the second half (H2: Jul-Dec) of 2023.

*Hint*: Use the `if()` function to compute a "half" column, and use `count()` for aggregation.

<details>
<summary>Solution</summary>

```cozoscript
?[half, sentiment, count(id)] :=
    *relationship{id, timestamp, sentiment},
    starts_with(timestamp, '2023'),
    half = if(timestamp < '2023-07', 'H1', 'H2')
:order half, sentiment
```

This groups relationships by half-year and sentiment, showing the distribution shift. You would expect more negative sentiment entries in H1 (the conflict period) and more positive entries in H2 (post-reconciliation, new team members joining).
</details>

### Exercise 4.4: Design a Validity-Based Schema

Redesign the `person` relation to use CozoDB's `Validity` type so that changes to a person's name, description, or role are automatically versioned.

Write the `:create` statement, then write a `:put` statement that adds a person, and a query using `@` to read the person's state at a specific point in time.

<details>
<summary>Solution</summary>

```cozoscript
# Create the versioned relation
:create versioned_person {
    id: String,
    valid_at: Validity
    =>
    name: String,
    description: String,
    role: String
}
```

```cozoscript
# Insert initial version
?[id, name, description, role] <- [[
    'sarah_martinez', 'Sarah Martinez', 'Senior scientist at the research lab', 'senior_scientist'
]]
:put versioned_person {id => name, description, role}
```

```cozoscript
# Later, update her role after promotion
?[id, name, description, role] <- [[
    'sarah_martinez', 'Sarah Martinez', 'Lead scientist and team leader', 'lead_scientist'
]]
:put versioned_person {id => name, description, role}
```

```cozoscript
# Time travel: query her profile as it was before the promotion
?[id, name, description, role] :=
    *versioned_person @ "2023-06-01T00:00:00" {id, name, description, role}
```

The `@` operator returns the version that was current at the specified timestamp. The earlier version with `role: 'senior_scientist'` is preserved in storage and accessible via time travel.
</details>

### Exercise 4.5: Detect Sentiment Changes Over Time

Write a query that finds all pairs of people whose relationship sentiment changed from positive to negative at some point, and another query for those that changed from negative to positive (reconciliation).

*Hint*: Self-join the relationship relation on `from_person` and `to_person` but with different timestamps.

<details>
<summary>Solution</summary>

```cozoscript
# Positive -> Negative transitions (conflicts)
?[from_person, to_person, good_time, bad_time] :=
    *relationship{from_person, to_person, timestamp: good_time, sentiment: 'positive'},
    *relationship{from_person, to_person, timestamp: bad_time, sentiment: 'negative'},
    good_time < bad_time,
    # Ensure no intermediate negative record exists between them
    not *relationship{from_person, to_person, timestamp: mid, sentiment: 'negative'},
        mid > good_time,
        mid < bad_time
:order from_person, to_person
```

```cozoscript
# Negative -> Positive transitions (reconciliations)
?[from_person, to_person, conflict_time, recovery_time] :=
    *relationship{from_person, to_person, timestamp: conflict_time, sentiment: 'negative'},
    *relationship{from_person, to_person, timestamp: recovery_time, sentiment: 'positive'},
    conflict_time < recovery_time,
    not *relationship{from_person, to_person, timestamp: mid, sentiment: 'positive'},
        mid > conflict_time,
        mid < recovery_time
:order from_person, to_person
```

A simpler version without the "no intermediate" constraint:

```cozoscript
# Any positive-to-negative transition
?[from_person, to_person, pos_ts, neg_ts] :=
    *relationship{from_person, to_person, timestamp: pos_ts, sentiment: 'positive'},
    *relationship{from_person, to_person, timestamp: neg_ts, sentiment: 'negative'},
    pos_ts < neg_ts
```
</details>

### Exercise 4.6: Create a Trigger for Relationship History

Write the CozoScript to:
1. Create a `relationship_change_log` relation
2. Set a trigger on the `relationship` relation that logs every `:put` operation (both new inserts and updates) into the change log

*Hint*: Use `_new[]` in the trigger body to access the row being written.

<details>
<summary>Solution</summary>

```cozoscript
# Step 1: Create the change log relation
:create relationship_change_log {
    log_id: String,
    logged_at: String
    =>
    operation: String,
    from_person: String,
    to_person: String,
    timestamp: String,
    relationship_type: String,
    sentiment: String,
    strength: Float,
    had_previous: Bool
}
```

```cozoscript
# Step 2: Set the trigger
::set_triggers relationship

on put {
    # Log new entries (no previous version existed)
    ?[log_id, logged_at, operation, from_person, to_person, timestamp,
      relationship_type, sentiment, strength, had_previous] :=
        _new[id, from_person, to_person, timestamp, relationship_type,
             _desc, sentiment, strength, _emb],
        not _old[id, from_person, to_person, timestamp, _, _, _, _, _],
        logged_at = now(),
        log_id = concat('log_new_', id, '_', timestamp),
        operation = 'insert',
        had_previous = false

    :put relationship_change_log {
        log_id, logged_at
        => operation, from_person, to_person, timestamp,
           relationship_type, sentiment, strength, had_previous
    }
}

on put {
    # Log updates (previous version existed)
    ?[log_id, logged_at, operation, from_person, to_person, timestamp,
      relationship_type, sentiment, strength, had_previous] :=
        _new[id, from_person, to_person, timestamp, relationship_type,
             _desc, sentiment, strength, _emb],
        _old[id, from_person, to_person, timestamp, _, _, _, _, _],
        logged_at = now(),
        log_id = concat('log_upd_', id, '_', timestamp),
        operation = 'update',
        had_previous = true

    :put relationship_change_log {
        log_id, logged_at
        => operation, from_person, to_person, timestamp,
           relationship_type, sentiment, strength, had_previous
    }
}
```

Now every `:put` to `relationship` automatically records a change log entry.
</details>

### Exercise 4.7: Atomic Temporal Update Transaction

Write a chained transaction (using `{}` blocks) that:
1. Reads the current relationship between Sarah and David for a given timestamp
2. Archives it to a `relationship_archive` relation
3. Inserts a new version with updated sentiment and strength

Use parameters `$from`, `$to`, `$old_timestamp`, `$new_timestamp`, `$new_sentiment`, `$new_strength`.

<details>
<summary>Solution</summary>

```cozoscript
{
    # Ensure the archive relation exists
    %ignore_error
    :create relationship_archive {
        id: String,
        from_person: String,
        to_person: String,
        timestamp: String,
        archived_at: String
        =>
        relationship_type: String,
        description: String,
        sentiment: String,
        strength: Float
    }
}
{
    # Step 1: Read current state and archive it
    ?[id, from_person, to_person, timestamp, archived_at,
      relationship_type, description, sentiment, strength] :=
        *relationship{id, from_person: $from, to_person: $to,
                      timestamp: $old_timestamp,
                      relationship_type, description, sentiment, strength},
        from_person = $from,
        to_person = $to,
        timestamp = $old_timestamp,
        archived_at = now()

    :put relationship_archive {
        id, from_person, to_person, timestamp, archived_at
        => relationship_type, description, sentiment, strength
    }
}
{
    # Step 2: Insert new temporal version
    # Reuse the id and other fields from the existing record,
    # but with updated timestamp, sentiment, and strength
    ?[id, from_person, to_person, timestamp, relationship_type,
      description, sentiment, strength, embedding] :=
        *relationship{id, from_person: $from, to_person: $to,
                      timestamp: $old_timestamp,
                      relationship_type, description, embedding},
        from_person = $from,
        to_person = $to,
        timestamp = $new_timestamp,
        sentiment = $new_sentiment,
        strength = $new_strength

    :put relationship {
        id, from_person, to_person, timestamp
        => relationship_type, description, sentiment, strength, embedding
    }
}
```

This transaction is atomic: if any step fails, the entire operation rolls back. The old version is safely archived before the new version is written.
</details>

### Exercise 4.8: Find the Busiest Month

Write a query that counts how many new relationships were recorded in each month, and returns the top 3 busiest months.

<details>
<summary>Solution</summary>

```cozoscript
?[month, count(id)] :=
    *relationship{id, timestamp},
    month = substr(timestamp, 0, 7)
:order -count(id)
:limit 3
```

If the timestamps do not use a uniform format with a parseable month prefix, you can fall back to the raw timestamp:

```cozoscript
?[timestamp, count(id)] :=
    *relationship{id, timestamp}
:order -count(id)
:limit 3
```

The result reveals which periods had the most relationship activity -- likely months when major events occurred (conference, conflict, new team member joining).
</details>

### Exercise 4.9 (Challenge): Build a Complete Timeline View

Write a query that merges all temporal data -- relationships, behaviors, and events -- into a single chronological timeline. Each row should have: `timestamp`, `source` (which relation), `actor` (person involved), and `summary` (a human-readable description).

*Hint*: Use three definitions of the same intermediate rule (one for each relation) to create a union, then sort.

<details>
<summary>Solution</summary>

```cozoscript
# Pull temporal data from relationships
timeline[timestamp, source, actor, summary] :=
    *relationship{from_person, to_person, timestamp,
                  relationship_type, sentiment, strength},
    source = 'relationship',
    actor = from_person,
    summary = concat(from_person, ' -> ', to_person, ': ',
                     relationship_type, ' [', sentiment, ', str=',
                     to_string(strength), ']')

# Pull temporal data from behaviors
timeline[timestamp, source, actor, summary] :=
    *behavior{person_id, timestamp, action, description},
    source = 'behavior',
    actor = person_id,
    summary = concat(person_id, ' performed: ', action, ' - ', description)

# Pull temporal data from events
timeline[timestamp, source, actor, summary] :=
    *event{timestamp, title, description},
    source = 'event',
    actor = 'team',
    summary = concat('EVENT: ', title, ' - ', description)

# Merge and sort
?[timestamp, source, actor, summary] := timeline[timestamp, source, actor, summary]
:order timestamp, source
```

This produces a complete narrative of everything that happened, in order. You can filter by actor, source, or time range by adding conditions to the entry rule.
</details>

### Exercise 4.10 (Challenge): Detect Conflict-Reconciliation Patterns

Write a query that finds all pairs of people who experienced a "conflict followed by reconciliation" pattern: a negative-sentiment relationship record followed by a positive-sentiment record within 6 months (i.e., the positive timestamp is at most 6 months after the negative one).

*Hint*: Since timestamps are strings like `'2023-06'`, you can use string comparison for approximate time ranges. A 6-month window from `'2023-06'` would be `'2023-12'`.

<details>
<summary>Solution</summary>

```cozoscript
# Find conflict-then-reconciliation patterns
# Since timestamps are YYYY-MM strings, adding 6 months means
# the reconciliation timestamp should be less than conflict + ~0.06 in string terms
# A simpler approach: just find negative-then-positive sequences

?[from_person, to_person, conflict_ts, reconciliation_ts,
   conflict_type, recovery_type, strength_gain] :=
    *relationship{from_person, to_person,
                  timestamp: conflict_ts, sentiment: 'negative',
                  relationship_type: conflict_type, strength: conflict_str},
    *relationship{from_person, to_person,
                  timestamp: reconciliation_ts, sentiment: 'positive',
                  relationship_type: recovery_type, strength: recovery_str},
    conflict_ts < reconciliation_ts,
    # No other negative record between conflict and reconciliation
    not *relationship{from_person, to_person,
                      timestamp: mid_ts, sentiment: 'negative'},
        mid_ts > conflict_ts,
        mid_ts < reconciliation_ts,
    # Approximate 6-month window check using string comparison
    # This works because YYYY-MM strings are lexicographically ordered
    reconciliation_ts <= concat(substr(conflict_ts, 0, 5),
        to_string(to_int(substr(conflict_ts, 5, 2)) + 6)),
    strength_gain = recovery_str - conflict_str
:order from_person, conflict_ts
```

A simpler version that just finds the pattern without the strict 6-month window:

```cozoscript
?[from_person, to_person, conflict_ts, recovery_ts] :=
    *relationship{from_person, to_person,
                  timestamp: conflict_ts, sentiment: 'negative'},
    *relationship{from_person, to_person,
                  timestamp: recovery_ts, sentiment: 'positive'},
    conflict_ts < recovery_ts,
    not *relationship{from_person, to_person,
                      timestamp: between_ts, sentiment: 'negative'},
        between_ts > conflict_ts,
        between_ts < recovery_ts
```

In the demo data, this should find the Sarah-David pair: negative in June 2023, positive again in August 2023 -- a classic conflict-reconciliation arc.
</details>

---

## Key Takeaways

1. **Real-world data changes over time.** A database that only stores the current state loses valuable history. Temporal data modeling preserves the full narrative.

2. **CozoDB's `Validity` type** provides built-in time travel: automatic versioning, retraction support, and `@`-based "as of" queries. Use it when transaction time is the relevant dimension.

3. **Application-level timestamps in keys** (as used in this project's `relationship` relation) give you full control over temporal semantics. Use them for event-sourced data where you control the timeline.

4. **Temporal queries** leverage the same Datalog primitives you already know: self-joins for detecting changes, aggregation for distribution analysis, string functions for time filtering.

5. **Triggers** (`::set_triggers`) automate temporal bookkeeping. Use `on put` with `_old[]` and `_new[]` to automatically archive previous versions.

6. **Chained transactions** (`{}` blocks) enable atomic multi-step temporal operations. Combined with `%ignore_error` and control flow, they provide robust temporal update patterns.

7. **Choose your temporal strategy** based on your needs: `Validity` for audit trails and system-managed versioning, application timestamps for event sourcing and domain-controlled timelines, bitemporal for the most rigorous requirements.

## Next Module

In **Module 5: Entity Extraction System Architecture**, you will see how all the CozoDB concepts from Modules 1-4 come together in a real system that extracts entities, relationships, and events from unstructured text and stores them in a temporal knowledge graph.
