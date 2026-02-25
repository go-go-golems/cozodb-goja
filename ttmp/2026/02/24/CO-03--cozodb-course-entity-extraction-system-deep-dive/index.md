---
Title: CozoDB Course - Entity Extraction System Deep Dive
Ticket: CO-03
Status: active
Topics:
    - cozodb
    - course
    - entity-extraction
    - datalog
    - javascript
    - goja
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cozo-relationship-js-runner/main.go
      Note: Go CLI runner for JS extractors
    - Path: cozo-relationship-js-runner/main.go:Go CLI runner
    - Path: cozo-relationship-js-runner/plugin_loader.go
      Note: Plugin descriptor loading protocol
    - Path: cozo_advanced_demo.py
      Note: Advanced CozoDB analytics queries
    - Path: cozo_advanced_demo.py:Advanced CozoDB queries
    - Path: cozo_demo.py
      Note: Primary CozoDB demo - schema creation + 10 queries
    - Path: cozo_demo.py:CozoDB demo with schema + 10 queries
    - Path: demo_text.txt
      Note: Sample narrative text for extraction
    - Path: demo_text.txt:Sample narrative text
    - Path: extract_relationships.py
      Note: Python LLM extraction pipeline
    - Path: extract_relationships.py:Python extraction pipeline
ExternalSources:
    - https://docs.cozodb.org/en/latest/
Summary: A 7-module course covering CozoDB fundamentals through entity extraction system architecture
LastUpdated: 2026-02-24T19:30:00-05:00
WhatFor: Learn CozoDB from scratch and understand the entity extraction pipeline in this repo
WhenToUse: Start with Module 1 if new to CozoDB. Jump to Module 5+ if already familiar with Datalog.
---


# CozoDB Course -- Entity Extraction System Deep Dive

## Course Overview

A comprehensive 7-module course covering CozoDB from absolute beginner to building a complete entity extraction pipeline. Each module includes theory, real examples from this repository, and hands-on exercises.

**Prerequisites**: Basic programming knowledge (Python or Go). No database experience required.

**Official CozoDB docs**: https://docs.cozodb.org/en/latest/

## Course Map

| Module | Title | Key Topics | Exercises |
|---|---|---|---|
| **1** | [CozoDB Fundamentals](reference/02-module-1-cozodb-fundamentals.md) | Datalog, rules, atoms, unification, `*` queries | 8 |
| **2** | [Data Modeling](reference/03-module-2-data-modeling-in-cozodb.md) | Schemas, types, mutations, indices (HNSW, FTS) | 10 |
| **3** | [Advanced Queries](reference/04-module-3-advanced-queries-and-graph-algorithms.md) | Aggregation, recursion, PageRank, BFS, community detection | 10 |
| **4** | [Time Travel & Temporal Data](reference/05-module-4-time-travel-and-temporal-data.md) | Validity type, triggers, transactions, temporal patterns | 10 |
| **5** | [Entity Extraction Architecture](reference/06-module-5-entity-extraction-system-architecture.md) | Pipeline flow, 4 entity types, LLM prompting, RDF vs entity-relation | 6 |
| **6** | [JavaScript/Goja Integration](reference/07-module-6-javascript-integration-with-goja.md) | Goja VM, module system, Geppetto LLM, plugin descriptors | 5 |
| **7** | [Capstone Project](reference/08-module-7-capstone-project.md) | Build a tech news knowledge graph end-to-end | 15+ |

## Learning Path

```
Module 1 (Fundamentals)
    |
    v
Module 2 (Data Modeling)
    |
    v
Module 3 (Advanced Queries) -----> Module 5 (Extraction Architecture)
    |                                    |
    v                                    v
Module 4 (Time Travel)            Module 6 (JS/Goja Integration)
    |                                    |
    +------>--------+-------<------------+
                    |
                    v
            Module 7 (Capstone)
```

## Exercise Scripts

Runnable exercise files are in the `scripts/` directory:

| Script | Module | Description |
|---|---|---|
| `01-module1-exercises.cozo` | 1 | Basic CozoScript queries |
| `02-module2-schema-setup.cozo` | 2 | Full schema creation |
| `03-module3-graph-queries.cozo` | 3 | Aggregation + graph algorithms |
| `04-module4-temporal.cozo` | 4 | Temporal queries + triggers |
| `05-capstone-heuristic-extractor.js` | 7 | Heuristic extraction plugin |
| `06-capstone-tech-news-schema.cozo` | 7 | Capstone CozoDB schema |
| `07-capstone-sample-article.txt` | 7 | Sample tech news for extraction |

## Supporting Documents

| Document | Purpose |
|---|---|
| [Implementation Diary](reference/01-implementation-diary.md) | Chronological record of course creation |

## Repository Subsystems Covered

The course walks through all four subsystems in this repository:

1. **Python Demo** (`extract_relationships.py`, `cozo_demo.py`) -- Modules 1-4
2. **Full-Stack Explorer** (`cozo-relationship-explorer/`) -- Referenced in Module 5
3. **Go Fact Extraction** (`manus-fact-extraction/`) -- RDF triple model in Module 5
4. **JS Plugin Runner** (`cozo-relationship-js-runner/`) -- Modules 5-6

## Quick Start

1. Read Module 1 to learn CozoScript basics
2. Try the exercises in `scripts/01-module1-exercises.cozo`
3. Progress through modules sequentially, or jump to Module 5 if you already know Datalog
