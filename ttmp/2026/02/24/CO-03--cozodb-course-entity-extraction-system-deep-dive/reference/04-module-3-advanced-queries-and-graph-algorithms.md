---
Title: Module 3 - Advanced Queries and Graph Algorithms
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
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo_advanced_demo.py:Advanced CozoDB demo showing aggregation, recursion, and graph algorithms"
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo_demo.py:Basic CozoDB demo showing schema creation and queries"
ExternalSources:
    - https://docs.cozodb.org/en/latest/queries.html
    - https://docs.cozodb.org/en/latest/algorithms.html
    - https://docs.cozodb.org/en/latest/aggregations.html
Summary: "Advanced CozoDB techniques: aggregation, semi-lattice aggregations, recursion, multi-hop graph traversal, built-in graph algorithms (PageRank, BFS, Dijkstra, community detection), pattern matching, and query optimization"
LastUpdated: 2026-02-24T18:41:06.092846205-05:00
WhatFor: "Third module of the CozoDB course - covers the analytical and graph power of Datalog"
WhenToUse: "After completing Modules 1 and 2, when you need aggregation, recursion, graph algorithms, or advanced pattern matching"
---

# Module 3: Advanced Queries and Graph Algorithms

## Learning Objectives

By the end of this module you will:
- Use aggregation operators to summarize data (count, sum, mean, min, max, collect, unique)
- Understand grouping variables and bag semantics in aggregation
- Write recursive rules for graph traversal and transitive closure
- Construct multi-hop queries to explore relationships at depth
- Invoke built-in graph algorithms (PageRank, BFS, Dijkstra, community detection, and more)
- Apply pattern matching techniques to find structural motifs in graphs
- Optimize queries using `::explain`, selective atom ordering, and indices

---

## 3.1 Aggregation

Aggregation is how you summarize data in CozoDB. Instead of SQL's `GROUP BY` clause, CozoDB uses **aggregation operators in the rule head**.

### Aggregation Operators

| Operator | Description | Example |
|---|---|---|
| `count(x)` | Count the number of values | `count(id)` |
| `sum(x)` | Sum of numeric values | `sum(amount)` |
| `mean(x)` | Arithmetic mean | `mean(score)` |
| `min(x)` | Minimum value | `min(timestamp)` |
| `max(x)` | Maximum value | `max(strength)` |
| `collect(x)` | Collect all values into a list | `collect(name)` |
| `unique(x)` | Collect distinct values into a list | `unique(category)` |

### The Grouping Rule

The key insight: **variables in the rule head that do NOT have an aggregation operator are grouping variables.** This is equivalent to SQL's `GROUP BY` -- but you never write `GROUP BY` explicitly. The non-aggregated columns *are* the groups.

```cozoscript
# Count relationships by type
# 'relationship_type' is a grouping variable
# 'count(id)' is the aggregation
?[relationship_type, count(id)] :=
    *relationship{id, relationship_type}
```

This is analogous to:
```sql
SELECT relationship_type, COUNT(id) FROM relationship GROUP BY relationship_type
```

### Multiple Aggregations

You can combine multiple aggregation operators in a single query:

```cozoscript
# Stats per relationship type
?[relationship_type, mean(strength), count(id), min(strength), max(strength)] :=
    *relationship{relationship_type, strength, id}
```

This produces one row per `relationship_type`, with the mean, count, min, and max of `strength` for each.

### No Grouping Variables = One Row

If every column in the rule head uses an aggregation operator, the result has exactly one row -- a global aggregate across all data:

```cozoscript
# Total number of relationships and overall mean strength
?[count(id), mean(strength)] :=
    *relationship{id, strength}
```

### `collect` and `unique`

These aggregation operators are particularly useful for gathering related data into lists:

```cozoscript
# For each person, collect all their relationship types into a list
?[from_person, collect(relationship_type)] :=
    *relationship{from_person, relationship_type}
```

Result:
| from_person | collect(relationship_type) |
|---|---|
| sarah_martinez | ['mentor', 'friend', 'colleague', 'mentor'] |
| david_chen | ['mentee', 'collaborator'] |

Note the duplicate 'mentor' in the collected list above. That leads us to an important point:

### Bag Semantics for Aggregation

**Aggregation in CozoDB uses bag (multiset) semantics, not set semantics.** This means duplicate values are counted. If Sarah has two 'mentor' relationships, `count(relationship_type)` counts both, and `collect(relationship_type)` includes both.

This is important because regular rule evaluation uses set semantics (deduplication), but aggregation deliberately preserves duplicates so that `count`, `sum`, and `mean` are correct.

If you want only distinct values in a collected list, use `unique`:

```cozoscript
# Distinct relationship types per person
?[from_person, unique(relationship_type)] :=
    *relationship{from_person, relationship_type}
```

### Combining Aggregation with Joins

Aggregation works naturally with multi-relation queries:

```cozoscript
# Count behaviors per person, showing person name
?[name, count(action)] :=
    *behavior{person_id, action},
    *person{id: person_id, name}
```

---

## 3.2 Semi-Lattice Aggregations

Some aggregation operators are special: they can be used inside **recursive rules**. These are called **semi-lattice aggregations** because they satisfy mathematical properties (associativity, commutativity, idempotency) that make them safe for iterative fixed-point computation.

### The Semi-Lattice Aggregation Operators

| Operator | Description |
|---|---|
| `min(x)` | Minimum value (converges downward) |
| `max(x)` | Maximum value (converges upward) |
| `shortest(x)` | Shortest list/path (by length) |
| `choice(x)` | Pick any one value (non-deterministic) |
| `union(x)` | Set union (grows monotonically) |

### Placement Rule

Semi-lattice aggregations **must come at the END of the rule head**, after all grouping variables and regular columns.

```cozoscript
# CORRECT: min_cost is last
shortest_path[node, min(cost)] := ...

# WRONG: min_cost is not last
# shortest_path[min(cost), node] := ...  -- ERROR
```

### Why These Are Special

Regular aggregation operators like `count` and `mean` cannot be used in recursive rules because their values can change unpredictably as new tuples are added during iteration. Semi-lattice aggregations, by contrast, only move in one direction (smaller, larger, shorter, or more inclusive), guaranteeing convergence.

### Example: Shortest Path Length via Recursion

```cozoscript
# Find shortest distance (hop count) from sarah_martinez to every reachable person
shortest_dist[person, min(dist)] :=
    *relationship{from_person: 'sarah_martinez', to_person: person},
    dist = 1

shortest_dist[person, min(dist)] :=
    shortest_dist[intermediate, d],
    *relationship{from_person: intermediate, to_person: person},
    dist = d + 1

?[person, dist] := shortest_dist[person, dist]
:order dist
```

Here `min(dist)` is a semi-lattice aggregation. As the recursion discovers new paths, it keeps only the minimum distance for each person. This is safe because `min` can only decrease or stay the same -- it never increases.

---

## 3.3 Recursion

Recursion is where Datalog truly shines compared to SQL. In CozoDB, a rule can reference itself, enabling natural expression of graph traversal, transitive closure, and reachability queries.

### Basic Recursion: Transitive Closure (Reachability)

The classic recursive query asks: "Who can be reached from a starting node?"

```cozoscript
# Base case: direct connections from Sarah
reachable[person] :=
    *relationship{from_person: 'sarah_martinez', to_person: person}

# Recursive case: if 'intermediate' is reachable, and intermediate connects to person,
# then person is also reachable
reachable[person] :=
    reachable[intermediate],
    *relationship{from_person: intermediate, to_person: person}

# Return all reachable people
?[person] := reachable[person]
```

**How it works:**

1. **Iteration 0 (base case)**: Find all direct connections from Sarah. Say this gives us {alice, bob, charlie}.
2. **Iteration 1**: For each person in {alice, bob, charlie}, find their connections. Say alice connects to {david, eve}. Now reachable = {alice, bob, charlie, david, eve}.
3. **Iteration 2**: For david and eve (newly added), find their connections. If no new people are found, we stop.
4. **Fixed point**: Iteration continues until no new tuples are produced. This is called **fixed-point iteration**.

### How CozoDB Handles Termination

CozoDB uses **stratified fixed-point evaluation**:

- The engine repeatedly evaluates the recursive rule until the result set stops changing.
- Because relations use set semantics, adding a tuple that already exists does not count as a change.
- Termination is guaranteed as long as the domain is finite (which it always is for database queries over finite stored data).

### Warning: Infinite Paths vs. Infinite Results

Recursion over a graph with cycles does NOT cause infinite loops. CozoDB computes the set of reachable nodes, and because sets don't grow when you add duplicates, the fixed point is reached. However, if you try to track the full path as a list, you need semi-lattice aggregation (like `shortest`) to keep paths finite.

```cozoscript
# SAFE: just tracking reachable nodes (set semantics handles cycles)
reachable[person] := *relationship{from_person: 'sarah_martinez', to_person: person}
reachable[person] := reachable[intermediate], *relationship{from_person: intermediate, to_person: person}

# CAUTION: tracking paths requires care with cycles
# Use a built-in algorithm (BFS/DFS) for path tracking instead
```

### Recursion with Computation

You can combine recursion with arithmetic using semi-lattice aggregations:

```cozoscript
# Find max relationship strength reachable from Sarah
# (propagate the maximum strength along paths)
max_strength[person, max(s)] :=
    *relationship{from_person: 'sarah_martinez', to_person: person, strength: s}

max_strength[person, max(s)] :=
    max_strength[intermediate, _],
    *relationship{from_person: intermediate, to_person: person, strength: s}

?[person, strength] := max_strength[person, strength]
:order -strength
```

---

## 3.4 Multi-hop Graph Traversal

Before reaching for recursion, simple multi-hop queries can be expressed directly by chaining relation lookups.

### 2-Hop Query

The project's demo includes a 2-hop traversal from Sarah Martinez:

```cozoscript
# Find people 2 hops away from Sarah, with the relationship types along the way
?[intermediate, final_person, rel1, rel2] :=
    *relationship{from_person: 'sarah_martinez', to_person: intermediate, relationship_type: rel1},
    *relationship{from_person: intermediate, to_person: final_person, relationship_type: rel2},
    final_person != 'sarah_martinez'
```

This reads naturally: "Find an intermediate person connected to Sarah via `rel1`, then find a final person connected to intermediate via `rel2`, excluding Sarah herself."

### 3-Hop Query

Just add another relation lookup:

```cozoscript
?[hop1, hop2, hop3, r1, r2, r3] :=
    *relationship{from_person: 'sarah_martinez', to_person: hop1, relationship_type: r1},
    *relationship{from_person: hop1, to_person: hop2, relationship_type: r2},
    *relationship{from_person: hop2, to_person: hop3, relationship_type: r3},
    hop3 != 'sarah_martinez',
    hop2 != 'sarah_martinez'
```

### N-Hop Generalization with Recursion

For arbitrary depth, switch to a recursive rule:

```cozoscript
# All people reachable within N hops, tracking distance
within_n_hops[person, min(hops)] :=
    *relationship{from_person: 'sarah_martinez', to_person: person},
    hops = 1

within_n_hops[person, min(hops)] :=
    within_n_hops[intermediate, h],
    *relationship{from_person: intermediate, to_person: person},
    hops = h + 1

# Filter to max 3 hops
?[person, hops] := within_n_hops[person, hops], hops <= 3
:order hops, person
```

### When to Use Which Approach

| Approach | Use When |
|---|---|
| Explicit N-hop (chained atoms) | You know the exact depth and want to capture each intermediate step |
| Recursive rule | You need variable or unbounded depth |
| Built-in algorithm (BFS/DFS) | You need path tracking, or performance matters on large graphs |

---

## 3.5 Fixed Rules (Built-in Algorithms)

CozoDB ships with a library of built-in graph algorithms implemented as **fixed rules**. These are invoked with the `<~` syntax and run highly optimized native code.

### Syntax

```
output_rule[col1, col2, ...] <~ AlgorithmName(*input_relation[col_a, col_b, ...], param: value)
```

Key points:
- `<~` introduces a fixed rule invocation.
- `*input_relation[...]` passes a stored relation as input. The columns in `[...]` specify which columns are used (and in what role -- typically `[from, to]` for edges or `[from, to, weight]` for weighted edges).
- You can also pass inline rules as input (without the `*`): `my_rule[...]`.
- Named parameters like `start:`, `goal:`, `undirected:` configure the algorithm.

### Preparing Edge Data

Most graph algorithms expect an edge relation with `[from_node, to_node]` or `[from_node, to_node, weight]`. You can use the project's `relationship` relation directly:

```cozoscript
# Unweighted edges
edges[from_p, to_p] := *relationship{from_person: from_p, to_person: to_p}

# Weighted edges (using strength as weight)
weighted_edges[from_p, to_p, weight] :=
    *relationship{from_person: from_p, to_person: to_p, strength: weight}
```

---

### PageRank: Find Important Nodes

PageRank identifies the most "important" or "influential" nodes in a graph based on the link structure. A node is important if many important nodes point to it.

```cozoscript
edges[from_p, to_p] := *relationship{from_person: from_p, to_person: to_p}

?[node, score] <~ PageRank(edges[from_p, to_p])
:order -score
```

The output has two columns: `node` and `score`. Higher scores indicate more important nodes.

**Parameters:**
- `undirected: true` -- treat edges as undirected (default: false)
- `theta: 0.85` -- damping factor (default: 0.85)
- `epsilon: 0.0001` -- convergence threshold
- `iterations: 100` -- max iterations

```cozoscript
edges[from_p, to_p] := *relationship{from_person: from_p, to_person: to_p}

?[node, score] <~ PageRank(edges[from_p, to_p], undirected: true)
:order -score
:limit 5
```

---

### BFS (Breadth-First Search): Shortest Path by Hop Count

BFS explores the graph layer by layer from a starting node. It finds shortest paths in terms of hop count (unweighted).

```cozoscript
edges[from_p, to_p] := *relationship{from_person: from_p, to_person: to_p}

?[node, distance, path] <~ BFS(edges[from_p, to_p], start: 'sarah_martinez')
:order distance
```

Output columns:
- `node` -- the reached node
- `distance` -- hop count from start
- `path` -- the actual path as a list of nodes

**Parameters:**
- `start` (required) -- starting node
- `goal` -- optional target node (stops early when found)
- `limit` -- max depth to explore

```cozoscript
# Find shortest path from Sarah to a specific person
edges[from_p, to_p] := *relationship{from_person: from_p, to_person: to_p}

?[node, distance, path] <~ BFS(edges[from_p, to_p], start: 'sarah_martinez', goal: 'david_chen')
```

---

### DFS (Depth-First Search): Deep Exploration

DFS explores as deep as possible before backtracking. Useful for topological analysis and detecting structure.

```cozoscript
edges[from_p, to_p] := *relationship{from_person: from_p, to_person: to_p}

?[node, pre_order, post_order] <~ DFS(edges[from_p, to_p])
:order pre_order
```

Output columns:
- `node` -- the visited node
- `pre_order` -- order in which node was first discovered
- `post_order` -- order in which node was fully explored

---

### ShortestPathDijkstra: Weighted Shortest Paths

When edges have weights (costs, distances, strengths), Dijkstra's algorithm finds the path with minimum total weight.

```cozoscript
# Invert strength to use as "cost" (higher strength = lower cost)
weighted_edges[from_p, to_p, cost] :=
    *relationship{from_person: from_p, to_person: to_p, strength: s},
    cost = 1.0 - s

?[cost, node, path] <~ ShortestPathDijkstra(
    weighted_edges[from_p, to_p, cost],
    start: 'sarah_martinez',
    goal: 'david_chen'
)
```

Output columns:
- `cost` -- total path cost
- `node` -- the goal node (or intermediate nodes depending on variant)
- `path` -- list of nodes in the shortest path

**Parameters:**
- `start` (required) -- starting node
- `goal` (required) -- target node
- `undirected: true` -- treat edges as undirected

---

### CommunityDetectionLouvain: Find Communities/Clusters

The Louvain algorithm detects communities -- densely connected subgroups within the graph.

```cozoscript
edges[from_p, to_p] := *relationship{from_person: from_p, to_person: to_p}

?[node, community] <~ CommunityDetectionLouvain(edges[from_p, to_p])
:order community, node
```

Each node is assigned a `community` label (integer). Nodes in the same community are more densely connected to each other than to nodes in other communities.

**Parameters:**
- `undirected: true` -- treat edges as undirected (recommended for community detection)
- `max_iter: 10` -- max iterations
- `keep_depth: 1` -- level of hierarchy to return

---

### TopSort: Topological Ordering

Topological sort orders nodes such that for every edge (A -> B), A comes before B. Only works on directed acyclic graphs (DAGs). Useful for dependency ordering.

```cozoscript
# Example: task dependencies
tasks[from_task, to_task] <- [
    ['design', 'implement'],
    ['implement', 'test'],
    ['test', 'deploy'],
    ['design', 'document'],
    ['document', 'deploy']
]

?[task, order] <~ TopSort(tasks[from_task, to_task])
:order order
```

If the graph contains cycles, TopSort will report an error.

---

### ConnectedComponents: Find Connected Subgraphs

Identifies which nodes belong to the same connected component (subgraph where every node can reach every other node, ignoring edge direction).

```cozoscript
edges[from_p, to_p] := *relationship{from_person: from_p, to_person: to_p}

?[node, component] <~ ConnectedComponents(edges[from_p, to_p])
:order component, node
```

Nodes with the same `component` value are connected to each other (directly or transitively).

---

### MinimumSpanningTree: Minimum Cost Tree

Finds a subset of edges that connects all nodes with minimum total weight, without cycles.

```cozoscript
weighted_edges[from_p, to_p, weight] :=
    *relationship{from_person: from_p, to_person: to_p, strength: weight}

?[from_node, to_node, weight] <~ MinimumSpanningTreePrim(
    weighted_edges[from_p, to_p, weight]
)
:order weight
```

---

### Centrality Metrics

Centrality measures identify the most structurally important nodes from different perspectives.

**ClosenessCentrality**: How close a node is to all other nodes (shorter average path = more central).

```cozoscript
edges[from_p, to_p] := *relationship{from_person: from_p, to_person: to_p}

?[node, centrality] <~ ClosenessCentrality(edges[from_p, to_p])
:order -centrality
```

**BetweennessCentrality**: How often a node lies on the shortest path between other pairs of nodes (bridges between communities).

```cozoscript
edges[from_p, to_p] := *relationship{from_person: from_p, to_person: to_p}

?[node, centrality] <~ BetweennessCentrality(edges[from_p, to_p])
:order -centrality
```

### Quick Reference: Algorithm Summary

| Algorithm | Input | Key Output | Use Case |
|---|---|---|---|
| `PageRank` | edges[from, to] | node, score | Find influential nodes |
| `BFS` | edges[from, to] | node, distance, path | Shortest path (unweighted) |
| `DFS` | edges[from, to] | node, pre_order, post_order | Graph exploration |
| `ShortestPathDijkstra` | edges[from, to, weight] | cost, node, path | Shortest path (weighted) |
| `CommunityDetectionLouvain` | edges[from, to] | node, community | Find clusters |
| `TopSort` | edges[from, to] | node, order | Dependency ordering |
| `ConnectedComponents` | edges[from, to] | node, component | Find subgraphs |
| `MinimumSpanningTreePrim` | edges[from, to, weight] | from, to, weight | Minimum cost tree |
| `ClosenessCentrality` | edges[from, to] | node, centrality | Node proximity |
| `BetweennessCentrality` | edges[from, to] | node, centrality | Bridge nodes |

> **Full algorithm reference**: https://docs.cozodb.org/en/latest/algorithms.html

---

## 3.6 Pattern Matching Techniques

Datalog's declarative nature makes it excellent for expressing structural patterns in graphs.

### Finding Triangles (A -> B, B -> C, C -> A)

A triangle is a cycle of length 3. This is a classic graph pattern:

```cozoscript
?[a, b, c] :=
    *relationship{from_person: a, to_person: b},
    *relationship{from_person: b, to_person: c},
    *relationship{from_person: c, to_person: a},
    a < b, b < c  # avoid duplicate triangles (each triangle appears in 3 rotations)
```

The `a < b, b < c` condition ensures each triangle is reported only once, not three times (once per rotation).

### Finding Mutual Connections

Find all pairs of people who are mutually connected (A -> B and B -> A):

```cozoscript
?[person_a, person_b] :=
    *relationship{from_person: person_a, to_person: person_b},
    *relationship{from_person: person_b, to_person: person_a},
    person_a < person_b  # avoid listing both (A,B) and (B,A)
```

### Finding Common Connections

Find all people connected to both person A and person B:

```cozoscript
?[common_connection] :=
    *relationship{from_person: 'sarah_martinez', to_person: common_connection},
    *relationship{from_person: 'david_chen', to_person: common_connection}
```

### Self-Joins for Comparing Entities

Find people who share the same relationship type with a given person:

```cozoscript
# People who both have a 'mentor' relationship with Sarah
?[person1, person2] :=
    *relationship{from_person: 'sarah_martinez', to_person: person1, relationship_type: 'mentor'},
    *relationship{from_person: 'sarah_martinez', to_person: person2, relationship_type: 'mentor'},
    person1 < person2  # avoid self-pairs and duplicates
```

### Finding Mixed Sentiment Relationships

From the project demo -- find people who have both positive and negative relationships (indicating complex social dynamics):

```cozoscript
# People who have at least one positive AND at least one negative outgoing relationship
?[person] :=
    *relationship{from_person: person, sentiment: 'positive'},
    *relationship{from_person: person, sentiment: 'negative'}
```

This works because the two atoms independently scan the `relationship` relation. A person must have at least one row matching each sentiment for the conjunction (AND) to succeed.

### Finding Sentiment Contradictions Between Two People

```cozoscript
# Find pairs where A has positive sentiment toward B, but B has negative toward A
?[person_a, person_b] :=
    *relationship{from_person: person_a, to_person: person_b, sentiment: 'positive'},
    *relationship{from_person: person_b, to_person: person_a, sentiment: 'negative'}
```

---

## 3.7 Combining Queries with Stored Data

CozoDB allows you to build multi-rule analytical queries that combine intermediate results, stored data, and computations.

### Using Query Results to Drive Analysis

```cozoscript
# Step 1: Identify high-strength relationships
strong_rels[from_p, to_p, strength] :=
    *relationship{from_person: from_p, to_person: to_p, strength},
    strength > 0.7

# Step 2: Count strong connections per person
?[person, count(connection)] :=
    strong_rels[person, connection, _]
:order -count(connection)
```

### Multi-Rule Analytical Queries

Complex analysis often requires multiple intermediate rules:

```cozoscript
# Find people whose PageRank exceeds a threshold AND who have negative relationships

# Step 1: Compute edges
edges[from_p, to_p] := *relationship{from_person: from_p, to_person: to_p}

# Step 2: Run PageRank
pr[node, score] <~ PageRank(edges[from_p, to_p])

# Step 3: Find people with negative outgoing relationships
has_negative[person] := *relationship{from_person: person, sentiment: 'negative'}

# Step 4: Combine results
?[person, score] :=
    pr[person, score],
    has_negative[person],
    score > 0.1
:order -score
```

### Building Derived Relations

You can save analytical results back to the database for later use:

```cozoscript
# First, compute the analysis
edges[from_p, to_p] := *relationship{from_person: from_p, to_person: to_p}
pr[node, score] <~ PageRank(edges[from_p, to_p])
?[node, score] := pr[node, score]

# Then store the results (in a separate query)
# :create page_rank_scores { node: String => score: Float }
# :put page_rank_scores { ... }
```

Note that the computation and the storage are typically separate queries. The host language (Python, Go, JavaScript) orchestrates this: run the analytical query, get the results, then insert them.

---

## 3.8 Query Optimization

As queries grow more complex, understanding how CozoDB executes them helps you write efficient code.

### `::explain` for Query Plans

Prefix any query with `::explain` to see the execution plan without running it:

```cozoscript
::explain
?[from_person, to_person, strength] :=
    *relationship{from_person, to_person, strength},
    strength > 0.5
```

This shows you:
- The order in which atoms are evaluated
- Which indices are used
- How joins are performed

### Atom Ordering: Most Selective First

CozoDB evaluates atoms left to right. Put the most **selective** (filtering) atoms first to reduce the number of intermediate results:

```cozoscript
# BETTER: constant binding first, narrows results immediately
?[name, action] :=
    *relationship{from_person: 'sarah_martinez', to_person: pid},
    *person{id: pid, name},
    *behavior{person_id: pid, action}

# WORSE: starts with full scan of behavior, then filters
?[name, action] :=
    *behavior{person_id: pid, action},
    *person{id: pid, name},
    *relationship{from_person: 'sarah_martinez', to_person: pid}
```

In the first version, the constant binding `from_person: 'sarah_martinez'` immediately narrows the result set. In the second version, the engine scans all behaviors before filtering.

### When Indices Help

CozoDB automatically creates indices on key columns. You can also create additional indices:

```cozoscript
# Create an index on from_person for faster lookups
::index create relationship:from_person_idx {from_person}
```

Indices help when:
- You frequently filter on a non-key column
- You join on a non-key column
- You have a very large relation and need fast lookups

### Avoiding Cross Products

A cross product happens when two atoms share no variables:

```cozoscript
# DANGER: cross product! p and b are completely independent
?[p_name, b_action] :=
    *person{name: p_name},
    *behavior{action: b_action}
```

If `person` has 10 rows and `behavior` has 100 rows, this produces 1,000 rows. Always ensure atoms are connected by shared variables unless you specifically want a cross product.

### Summary of Optimization Tips

| Tip | Why It Helps |
|---|---|
| Put constant bindings first | Narrows results immediately |
| Put most selective atoms first | Reduces intermediate result size |
| Avoid cross products | Prevents combinatorial explosion |
| Use named binding `{}` | Only retrieves columns you need |
| Use `::explain` | Reveals the actual execution plan |
| Create indices on frequently filtered non-key columns | Speeds up lookups |

---

## Exercises

### Exercise 3.1: Count Relationships by Type

Write a query that counts the number of relationships for each `relationship_type`. Sort by count in descending order.

*Hint*: Use `count(id)` as an aggregation operator in the rule head.

<details>
<summary>Solution</summary>

```cozoscript
?[relationship_type, count(id)] :=
    *relationship{id, relationship_type}
:order -count(id)
```
</details>

### Exercise 3.2: Most Connected Person

Find the person with the most outgoing connections. Return the person's name and their connection count.

*Hint*: Use `count` with `:order` and `:limit`.

<details>
<summary>Solution</summary>

```cozoscript
?[from_person, count(to_person)] :=
    *relationship{from_person, to_person}
:order -count(to_person)
:limit 1
```
</details>

### Exercise 3.3: Average Strength by Sentiment

Compute the average relationship strength grouped by sentiment (positive, negative, neutral). Also include the count and min/max strength for each group.

<details>
<summary>Solution</summary>

```cozoscript
?[sentiment, mean(strength), count(id), min(strength), max(strength)] :=
    *relationship{id, sentiment, strength}
:order sentiment
```
</details>

### Exercise 3.4: Recursive Reachability

Write a recursive query to find all people reachable from `sarah_martinez` through any chain of relationships. Return the list of reachable people.

<details>
<summary>Solution</summary>

```cozoscript
# Base case: direct connections
reachable[person] :=
    *relationship{from_person: 'sarah_martinez', to_person: person}

# Recursive case: indirect connections
reachable[person] :=
    reachable[intermediate],
    *relationship{from_person: intermediate, to_person: person}

?[person] := reachable[person]
:order person
```
</details>

### Exercise 3.5: PageRank on the Relationship Graph

Run PageRank on the relationship graph. Return the top 5 most influential people with their scores.

<details>
<summary>Solution</summary>

```cozoscript
edges[from_p, to_p] := *relationship{from_person: from_p, to_person: to_p}

?[node, score] <~ PageRank(edges[from_p, to_p])
:order -score
:limit 5
```
</details>

### Exercise 3.6: Find All Triangles

Find all triangles in the relationship network. A triangle is a set of three people (A, B, C) where A -> B, B -> C, and C -> A all exist. Avoid duplicate triangles.

*Hint*: Use ordering constraints (`a < b, b < c`) to deduplicate.

<details>
<summary>Solution</summary>

```cozoscript
?[a, b, c] :=
    *relationship{from_person: a, to_person: b},
    *relationship{from_person: b, to_person: c},
    *relationship{from_person: c, to_person: a},
    a < b, b < c
```

If no strict ordering exists on person IDs, you might also deduplicate with `unique`:

```cozoscript
# Alternative approach using sorted triples
triangle[a, b, c] :=
    *relationship{from_person: x, to_person: y},
    *relationship{from_person: y, to_person: z},
    *relationship{from_person: z, to_person: x},
    a = min(x, min(y, z)),
    c = max(x, max(y, z)),
    b = if(x != a && x != c, x, if(y != a && y != c, y, z))

?[a, b, c] := triangle[a, b, c]
```
</details>

### Exercise 3.7: Shortest Path Between Two People

Use the BFS algorithm to find the shortest path between `sarah_martinez` and another person of your choice. Display the distance and the path.

<details>
<summary>Solution</summary>

```cozoscript
edges[from_p, to_p] := *relationship{from_person: from_p, to_person: to_p}

?[node, distance, path] <~ BFS(edges[from_p, to_p], start: 'sarah_martinez', goal: 'david_chen')
```
</details>

### Exercise 3.8: Community Detection

Use the Louvain algorithm to find communities in the relationship network. Show how many people are in each community.

*Hint*: Run community detection first, then aggregate the results in a second rule.

<details>
<summary>Solution</summary>

```cozoscript
edges[from_p, to_p] := *relationship{from_person: from_p, to_person: to_p}

communities[node, community] <~ CommunityDetectionLouvain(edges[from_p, to_p], undirected: true)

?[community, count(node), collect(node)] := communities[node, community]
:order community
```
</details>

### Exercise 3.9: People Within 3 Hops of Sarah

Using recursion with semi-lattice aggregation, find all people connected to `sarah_martinez` within 3 hops. Show the minimum hop distance for each person.

<details>
<summary>Solution</summary>

```cozoscript
# Recursive rule with min (semi-lattice aggregation)
within_hops[person, min(dist)] :=
    *relationship{from_person: 'sarah_martinez', to_person: person},
    dist = 1

within_hops[person, min(dist)] :=
    within_hops[intermediate, d],
    *relationship{from_person: intermediate, to_person: person},
    dist = d + 1

# Filter to 3 hops
?[person, dist] := within_hops[person, dist], dist <= 3
:order dist, person
```
</details>

### Exercise 3.10: Challenge -- Influence Score

Compute an "influence score" for each person that combines their PageRank score with their connection count. Define influence as: `influence = pagerank_score * 1000 + connection_count`. Return the top 5 most influential people.

*Hint*: Compute PageRank and connection counts in separate rules, then join them.

<details>
<summary>Solution</summary>

```cozoscript
# Step 1: Compute edges
edges[from_p, to_p] := *relationship{from_person: from_p, to_person: to_p}

# Step 2: PageRank
pr[node, score] <~ PageRank(edges[from_p, to_p])

# Step 3: Count connections (both directions)
outgoing[person, count(other)] := *relationship{from_person: person, to_person: other}
incoming[person, count(other)] := *relationship{from_person: other, to_person: person}

# Step 4: Combine (handle people who might have only incoming or outgoing)
connections[person, out_count + in_count] :=
    outgoing[person, out_count],
    incoming[person, in_count]
connections[person, out_count] :=
    outgoing[person, out_count],
    not incoming[person, _]
connections[person, in_count] :=
    incoming[person, in_count],
    not outgoing[person, _]

# Step 5: Compute influence score
?[person, influence, pr_score, conn_count] :=
    pr[person, pr_score],
    connections[person, conn_count],
    influence = pr_score * 1000 + conn_count
:order -influence
:limit 5
```
</details>

---

## Key Takeaways

1. **Aggregation uses operators in the rule head**, not a GROUP BY clause. Non-aggregated variables are automatically the grouping variables.
2. **Bag semantics for aggregation** -- duplicates count. Use `unique()` if you want distinct values.
3. **Semi-lattice aggregations** (`min`, `max`, `shortest`, `choice`, `union`) are the only aggregations safe for recursion. They must appear at the end of the rule head.
4. **Recursive rules** enable transitive closure and reachability. CozoDB guarantees termination via fixed-point iteration over finite domains.
5. **Multi-hop queries** can be expressed as explicit chains (for known depth) or recursive rules (for arbitrary depth).
6. **Built-in algorithms** (`<~` syntax) provide optimized implementations of PageRank, BFS, DFS, Dijkstra, community detection, centrality metrics, and more.
7. **Pattern matching** is natural in Datalog -- triangles, mutual connections, and structural motifs are just conjunctions of atoms.
8. **Query optimization**: use `::explain`, put selective atoms first, avoid cross products, and index frequently filtered columns.

## Next Module

In **Module 4: Time Travel and Temporal Data**, you'll learn how CozoDB's `Validity` column type enables historical queries, how to store and query temporal data, and how to implement "as-of" queries for the entity extraction system.
