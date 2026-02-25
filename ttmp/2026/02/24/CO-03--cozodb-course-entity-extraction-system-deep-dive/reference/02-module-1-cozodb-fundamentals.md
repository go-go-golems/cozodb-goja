---
Title: "Module 1 - CozoDB Fundamentals"
Ticket: CO-03
Status: active
Topics:
    - cozodb
    - course
    - datalog
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo_demo.py:Basic CozoDB demo showing schema creation and queries"
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/demo_text.txt:Sample narrative text used for extraction"
ExternalSources:
    - https://docs.cozodb.org/en/latest/
    - https://docs.cozodb.org/en/latest/queries.html
Summary: "Introduction to CozoDB for beginners: what it is, why Datalog, basic syntax, first queries"
LastUpdated: 2026-02-24T18:20:00-05:00
WhatFor: "First module of the CozoDB course - establishes all foundational concepts"
WhenToUse: "Start here if you have never used CozoDB or Datalog before"
---

# Module 1: CozoDB Fundamentals

## Learning Objectives

By the end of this module you will:
- Understand what CozoDB is and why it uses Datalog
- Write and run your first CozoScript queries
- Understand rules, atoms, variables, and unification
- Query stored relations using both positional and named binding
- Use basic filtering and sorting

## 1.1 What is CozoDB?

CozoDB is a **hybrid relational-graph-vector database** that uses **Datalog** as its query language. Unlike SQL databases (PostgreSQL, MySQL) or property-graph databases (Neo4j), CozoDB combines:

| Capability | How CozoDB Does It |
|---|---|
| Relational queries | Datalog rules over stored relations |
| Graph traversal | Recursive Datalog rules + built-in graph algorithms |
| Vector search | HNSW indices with cosine/L2/IP distance |
| Full-text search | Built-in FTS indices with Boolean queries |
| Time travel | `Validity` typed columns for historical queries |

**Why Datalog instead of SQL?**

Datalog is a declarative logic programming language. Where SQL says *how* to join tables, Datalog says *what* you want and lets the engine figure out the rest. This makes recursive queries (graph traversal, transitive closure) natural rather than awkward.

> **Official docs**: https://docs.cozodb.org/en/latest/

### Key Differences from SQL

| SQL | CozoDB (Datalog) |
|---|---|
| `SELECT ... FROM ... WHERE ...` | `?[columns] := *relation{bindings}, filters` |
| `JOIN` | Just mention multiple relations in the same rule |
| `UNION` | Multiple definitions of the same rule name |
| `WITH RECURSIVE` | Direct self-recursion in rules |
| `CREATE TABLE` | `:create relation_name { schema }` |
| `INSERT INTO` | `:put relation_name { schema }` |

### Storage Backends

CozoDB supports multiple storage engines:
- **`mem`** -- In-memory (great for testing, ephemeral work)
- **`sqlite`** -- SQLite-backed (persistent, single-process)
- **`rocksdb`** -- RocksDB-backed (persistent, higher performance for large datasets)

## 1.2 Your First CozoScript Query

The simplest possible query uses a **constant rule** to return literal data:

```cozoscript
?[name, age] <- [
    ['Alice', 30],
    ['Bob', 25],
    ['Charlie', 35]
]
```

Let's break this down:

- `?` is the **entry rule** -- its result is what the query returns
- `[name, age]` declares the **output columns** (also called the rule *head*)
- `<-` means "gets data from" (constant rule syntax)
- The right side is an array of arrays -- each inner array is a row

**Try it**: This is a complete, runnable query. No tables needed.

### The Entry Rule `?`

Every CozoScript query must define exactly one rule named `?`. This is the "main" function of your query. All other rules exist to feed data into `?`.

```cozoscript
# This rule computes intermediate data
adults[name, age] <- [['Alice', 30], ['Bob', 25], ['Charlie', 35]]

# The entry rule filters and returns results
?[name, age] := adults[name, age], age >= 30
```

## 1.3 Rules in Depth

CozoScript has three types of rules:

### Inline Rules (`:=`)

These define logic using **atoms** (the body of the rule):

```cozoscript
?[name, doubled_age] := adults[name, age], doubled_age = age * 2
```

The body (after `:=`) is a comma-separated list of atoms. The comma means **AND** (conjunction).

### Constant Rules (`<-`)

Embed literal data directly:

```cozoscript
fruits[name, color] <- [
    ['apple', 'red'],
    ['banana', 'yellow'],
    ['grape', 'purple']
]
```

This is actually syntactic sugar for the `Constant` fixed rule:
```cozoscript
fruits[name, color] <~ Constant(data: [['apple', 'red'], ['banana', 'yellow'], ['grape', 'purple']])
```

### Fixed Rules (`<~`)

Invoke built-in algorithms (covered in detail in Module 3):

```cozoscript
?[] <~ PageRank(*edges[])
```

## 1.4 Atoms: The Building Blocks

The body of an inline rule is built from **atoms**:

### 1. Rule Application

Reference another rule by name with positional binding:

```cozoscript
people[name, age] <- [['Alice', 30], ['Bob', 25]]
?[name] := people[name, age], age > 28
```

Here `people[name, age]` is a rule application atom. It **binds** `name` and `age` to each row of the `people` relation.

### 2. Stored Relation Application (the `*` prefix)

Query a stored (on-disk) relation. Two syntaxes:

**Positional binding** (must match column order exactly):
```cozoscript
?[id, name] := *person[id, name, _description, _first_mentioned, _embedding]
```

**Named binding** (preferred -- any order, skip columns you don't need):
```cozoscript
?[id, name] := *person{id, name}
```

The named `{}` syntax is much more practical. You only mention the columns you need, and you can use shorthand: `*person{name}` is the same as `*person{name: name}`.

You can also bind to a constant value in the same expression:
```cozoscript
# Bind from_person to the literal 'sarah_martinez'
?[to_person] := *relationship{from_person: 'sarah_martinez', to_person}
```

### 3. Expression Atoms (Filters)

Any boolean expression acts as a filter:

```cozoscript
?[name, age] := people[name, age], age > 25, age < 35
```

All variables in the expression must already be bound by a prior atom.

### 4. Unification Atoms

Introduce new bindings with `=`:

```cozoscript
?[name, birth_year] := people[name, age], birth_year = 2026 - age
```

**Important**: `=` is **unification** (binds a new variable). `==` is **equality** (both sides must already be bound).

```cozoscript
# = introduces the binding (left side is new variable)
?[label] := people[name, _], label = concat('Person: ', name)

# == checks equality (both sides already bound)
?[name] := people[name, age], name == 'Alice'
```

### 5. The Underscore Variable `_`

Use `_` when you don't care about a column:

```cozoscript
# We only want names, not ages
?[name] := people[name, _]
```

Each `_` is independent -- multiple `_` in the same rule do NOT bind to each other.

## 1.5 Variables and Binding

A variable is **bound** when it appears in a non-negated rule application or on the left side of a unification. Every variable in the rule head must be bound somewhere in the body.

```cozoscript
# 'name' is bound by people[name, _]
# 'greeting' is bound by the unification
?[name, greeting] := people[name, _], greeting = concat('Hello, ', name)
```

**Variables are case-sensitive** and follow standard naming rules. By convention, use `snake_case`.

### Set Semantics

All relations in CozoDB use **set semantics** -- duplicate rows are automatically removed. Even if a rule would produce the same row multiple times, the result contains only one copy.

### Binding with `in` (List Unification)

Bind a variable to multiple values:

```cozoscript
?[x, x_squared] := x in [1, 2, 3, 4, 5], x_squared = x * x
```

This returns 5 rows.

## 1.6 Querying Stored Relations

In this course's project, the database has four stored relations. Here's how to query them:

### List All People

```cozoscript
?[id, name, description] := *person{id, name, description}
```

This is equivalent to SQL's `SELECT id, name, description FROM person`.

### Filter with Named Binding

```cozoscript
?[to_person, relationship_type, sentiment] :=
    *relationship{from_person: 'sarah_martinez', to_person, relationship_type, sentiment}
```

Notice: `from_person: 'sarah_martinez'` is a bind-and-filter in one step. It's like `WHERE from_person = 'sarah_martinez'` in SQL.

### Joining Relations (No JOIN Keyword Needed!)

In Datalog, if two atoms share the same variable name, they're automatically joined:

```cozoscript
?[person_name, action, description] :=
    *behavior{person_id, action, description},
    *person{id: person_id, name: person_name}
```

The shared variable `person_id` naturally joins `behavior.person_id` with `person.id`. No `JOIN ... ON` syntax needed.

### Multi-Table Joins

You can join as many relations as you want -- just keep sharing variable names:

```cozoscript
# Find the behavior descriptions for people in relationships with Sarah
?[person_name, rel_type, action, action_desc] :=
    *relationship{from_person: 'sarah_martinez', to_person: pid, relationship_type: rel_type},
    *person{id: pid, name: person_name},
    *behavior{person_id: pid, action, description: action_desc}
```

## 1.7 Sorting and Limiting Results

### `:order` (Sort)

```cozoscript
?[timestamp, from_person, to_person] :=
    *relationship{timestamp, from_person, to_person}
:order timestamp
```

Descending order with `-`:
```cozoscript
?[timestamp, from_person, to_person] :=
    *relationship{timestamp, from_person, to_person}
:order -timestamp
```

Sort by multiple columns:
```cozoscript
?[from_person, to_person, timestamp] :=
    *relationship{from_person, to_person, timestamp}
:order from_person, -timestamp
```

### `:limit` and `:offset`

```cozoscript
?[name, description] := *person{name, description}
:limit 3
:offset 1
```

### `:timeout`

Abort queries that run too long:
```cozoscript
?[a, b, c] := some_expensive_rule[a, b, c]
:timeout 10   # abort after 10 seconds
```

### `:assert`

Check that a query returns results (or doesn't):
```cozoscript
# Fails with error if no results
?[name] := *person{name: 'nonexistent_person', name}
:assert some

# Fails with error if any results
?[name] := *person{name: 'should_not_exist', name}
:assert none
```

## 1.8 Multiple Rule Definitions (Disjunction / UNION)

Define the same rule name multiple times to get the union of results:

```cozoscript
# All people who are connected to Sarah (either direction)
sarah_connections[other_person] :=
    *relationship{from_person: 'sarah_martinez', to_person: other_person}
sarah_connections[other_person] :=
    *relationship{from_person: other_person, to_person: 'sarah_martinez'}

?[person] := sarah_connections[person]
```

This is like SQL `UNION` but more natural.

You can also use `or` inline:

```cozoscript
?[person] :=
    *relationship{from_person: 'sarah_martinez', to_person: person}
    or *relationship{from_person: person, to_person: 'sarah_martinez'}
```

## 1.9 Negation

Prefix atoms with `not` to exclude matches:

```cozoscript
# People who are NOT in any relationship as from_person
?[name] := *person{id, name}, not *relationship{from_person: id}
```

**Safety rule**: At least one binding used in a negated atom must be bound elsewhere in a non-negated context. You can't introduce new variables through negation.

```cozoscript
# WRONG -- 'other' is only bound in the negated atom
?[name] := *person{name}, not *relationship{from_person: name, to_person: other}

# RIGHT -- we only check existence, not binding new vars
?[name] := *person{id, name}, not *relationship{from_person: id}
```

## 1.10 Comments

CozoDB supports line comments with `#` or `//`:

```cozoscript
# This is a comment
// This is also a comment
?[name] := *person{name}  # inline comment
```

## 1.11 Parameters

Queries can accept parameters from the host language using `$` prefix:

```cozoscript
?[name, description] := *person{id: $person_id, name, description}
```

Parameters are passed as a dictionary/map from Python, Go, or JavaScript:

```python
# Python
db.run("?[name] := *person{id: $pid, name}", {"pid": "sarah_martinez"})
```

```go
// Go
db.Run("?[name] := *person{id: $pid, name}", cozo.Map{"pid": "sarah_martinez"})
```

## 1.12 Common Built-in Functions (Preview)

CozoDB has many built-in functions. Here are some frequently used ones:

| Function | Description | Example |
|---|---|---|
| `concat(a, b, ...)` | String concatenation | `concat('Hello', ' ', name)` |
| `length(s)` | String/list length | `length(name)` |
| `starts_with(s, prefix)` | String prefix check | `starts_with(timestamp, '2023')` |
| `ends_with(s, suffix)` | String suffix check | `ends_with(name, 'ez')` |
| `contains(s, sub)` | Substring check | `contains(description, 'mentor')` |
| `lowercase(s)` | Lowercase string | `lowercase(name)` |
| `to_float(x)` | Convert to float | `to_float('3.14')` |
| `to_int(x)` | Convert to integer | `to_int('42')` |
| `is_null(x)` | Null check | `is_null(location)` |
| `coalesce(a, b)` | First non-null | `coalesce(location, 'unknown')` |
| `now()` | Current timestamp | `now()` |
| `rand_float()` | Random [0,1) | `rand_float()` |
| `vec(list)` | Create vector from list | `vec([1.0, 2.0, 3.0])` |

> **Full function reference**: https://docs.cozodb.org/en/latest/functions.html

---

## Exercises

### Exercise 1.1: Hello CozoDB

Write a constant-rule query that returns the following data:

| city | country | population_millions |
|---|---|---|
| Tokyo | Japan | 37.4 |
| Delhi | India | 32.9 |
| Shanghai | China | 28.5 |
| Sao Paulo | Brazil | 22.4 |

<details>
<summary>Solution</summary>

```cozoscript
?[city, country, population_millions] <- [
    ['Tokyo', 'Japan', 37.4],
    ['Delhi', 'India', 32.9],
    ['Shanghai', 'China', 28.5],
    ['Sao Paulo', 'Brazil', 22.4]
]
```
</details>

### Exercise 1.2: Filtering

Using the cities data from Exercise 1.1, write a query that returns only cities with population over 30 million.

<details>
<summary>Solution</summary>

```cozoscript
cities[city, country, pop] <- [
    ['Tokyo', 'Japan', 37.4],
    ['Delhi', 'India', 32.9],
    ['Shanghai', 'China', 28.5],
    ['Sao Paulo', 'Brazil', 22.4]
]

?[city, pop] := cities[city, _, pop], pop > 30
```
</details>

### Exercise 1.3: Computed Columns

Write a query over the cities data that adds a `label` column with the format `"City (Country)"`.

*Hint*: Use the `concat()` function.

<details>
<summary>Solution</summary>

```cozoscript
cities[city, country, pop] <- [
    ['Tokyo', 'Japan', 37.4],
    ['Delhi', 'India', 32.9],
    ['Shanghai', 'China', 28.5],
    ['Sao Paulo', 'Brazil', 22.4]
]

?[label, pop] := cities[city, country, pop], label = concat(city, ' (', country, ')')
:order -pop
```
</details>

### Exercise 1.4: Natural Joins

Given these two constant rules, write a query that joins them to show employee names with their department names:

```cozoscript
employees[id, name, dept_id] <- [
    [1, 'Alice', 'eng'],
    [2, 'Bob', 'sales'],
    [3, 'Charlie', 'eng'],
    [4, 'Diana', 'hr']
]

departments[id, dept_name] <- [
    ['eng', 'Engineering'],
    ['sales', 'Sales'],
    ['hr', 'Human Resources']
]
```

<details>
<summary>Solution</summary>

```cozoscript
employees[id, name, dept_id] <- [
    [1, 'Alice', 'eng'],
    [2, 'Bob', 'sales'],
    [3, 'Charlie', 'eng'],
    [4, 'Diana', 'hr']
]

departments[id, dept_name] <- [
    ['eng', 'Engineering'],
    ['sales', 'Sales'],
    ['hr', 'Human Resources']
]

?[name, dept_name] := employees[_, name, dept_id], departments[dept_id, dept_name]
:order name
```

The shared variable `dept_id` automatically joins the two relations.
</details>

### Exercise 1.5: Union / Disjunction

Write a query using the employees data above that returns names of people who are in either Engineering OR Sales, using two definitions of the same rule.

<details>
<summary>Solution</summary>

```cozoscript
employees[id, name, dept_id] <- [
    [1, 'Alice', 'eng'],
    [2, 'Bob', 'sales'],
    [3, 'Charlie', 'eng'],
    [4, 'Diana', 'hr']
]

target[name] := employees[_, name, 'eng']
target[name] := employees[_, name, 'sales']

?[name] := target[name]
:order name
```
</details>

### Exercise 1.6: Using `in` for Multiple Values

Rewrite Exercise 1.5 using `in` instead of multiple rule definitions.

<details>
<summary>Solution</summary>

```cozoscript
employees[id, name, dept_id] <- [
    [1, 'Alice', 'eng'],
    [2, 'Bob', 'sales'],
    [3, 'Charlie', 'eng'],
    [4, 'Diana', 'hr']
]

?[name] := employees[_, name, dept_id], dept_id in ['eng', 'sales']
:order name
```
</details>

### Exercise 1.7: Querying the Project Data

Assuming the project's CozoDB instance is loaded with the demo data, write queries to:

1. List all persons (just name and description)
2. Find all relationships where the sentiment is 'negative'
3. Find all behaviors performed by 'david_chen', ordered by timestamp

<details>
<summary>Solution</summary>

```cozoscript
# 1. All persons
?[name, description] := *person{name, description}
```

```cozoscript
# 2. Negative relationships
?[from_person, to_person, description] :=
    *relationship{from_person, to_person, sentiment: 'negative', description}
```

```cozoscript
# 3. David's behaviors
?[timestamp, action, description] :=
    *behavior{person_id: 'david_chen', timestamp, action, description}
:order timestamp
```
</details>

### Exercise 1.8: Negation

Using the employees/departments data, find all employees who are NOT in the Engineering department.

<details>
<summary>Solution</summary>

```cozoscript
employees[id, name, dept_id] <- [
    [1, 'Alice', 'eng'],
    [2, 'Bob', 'sales'],
    [3, 'Charlie', 'eng'],
    [4, 'Diana', 'hr']
]

?[name] := employees[_, name, dept_id], dept_id != 'eng'
:order name
```

Or using `not`:
```cozoscript
eng_employees[name] := employees[_, name, 'eng']

?[name] := employees[_, name, _], not eng_employees[name]
:order name
```
</details>

---

## Key Takeaways

1. **CozoDB uses Datalog**, not SQL. The mindset is "declare what you want" not "describe how to get it."
2. **`?` is the entry rule** -- every query must have one.
3. **Three rule types**: inline (`:=`), constant (`<-`), fixed (`<~`).
4. **Stored relations use `*` prefix** and support named binding with `{}`.
5. **Joins happen naturally** through shared variable names.
6. **`=` is unification** (binding), `==` is equality (testing).
7. **Parameters use `$` prefix** for host-language integration.
8. **Set semantics** -- all results are automatically deduplicated.

## Next Module

In **Module 2: Data Modeling**, you'll learn how to design schemas, understand keys vs values, create/modify/delete relations, and model entities and relationships for the extraction system.
