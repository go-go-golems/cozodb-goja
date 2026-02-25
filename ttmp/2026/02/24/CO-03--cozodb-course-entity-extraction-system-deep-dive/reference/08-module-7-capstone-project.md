---
Title: "Module 7 - Capstone Project"
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
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/extract_relationships.py:Python extraction pipeline to model after"
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo_demo.py:CozoDB loading and querying to model after"
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo_advanced_demo.py:Advanced query patterns"
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/relation_extractor_template.js:JS plugin to extend"
ExternalSources:
    - https://docs.cozodb.org/en/latest/
Summary: "Capstone project: build a complete entity extraction + CozoDB pipeline from scratch"
LastUpdated: 2026-02-24T19:15:00-05:00
WhatFor: "Apply all course concepts in a guided end-to-end project"
WhenToUse: "After completing Modules 1-6"
---

# Module 7: Capstone Project -- Build a Knowledge Graph

## Overview

In this capstone, you will build a complete knowledge graph extraction system from scratch. You'll apply everything from Modules 1-6:

- **Module 1**: CozoScript fundamentals for querying
- **Module 2**: Schema design for your entities
- **Module 3**: Advanced queries and graph algorithms
- **Module 4**: Temporal tracking and triggers
- **Module 5**: Extraction pipeline architecture
- **Module 6**: JavaScript plugin system

## The Project

You will build a **Technology News Knowledge Graph** that:
1. Extracts entities from tech news articles
2. Stores them in CozoDB with vector embeddings
3. Supports complex graph queries
4. Tracks entity evolution over time

## Phase 1: Schema Design

### Task 1.1: Define Your Entity Types

Design CozoDB schemas for these five entity types:

| Entity | Key Fields | Value Fields |
|---|---|---|
| **Company** | `id` | name, industry, hq_city, founded_year, description, embedding |
| **Person** | `id` | name, title, company_id, description, embedding |
| **Product** | `id` | name, company_id, category, launch_date, description, embedding |
| **Partnership** | `company_a`, `company_b`, `timestamp` | type, description, value_usd, sentiment, embedding |
| **Funding** | `company_id`, `round_date` | round_type, amount_usd, lead_investor, description, embedding |

Write the `:create` statements.

<details>
<summary>Guided Solution</summary>

```cozoscript
# Core entities
:create company {
    id: String
    =>
    name: String,
    industry: String,
    hq_city: String?,
    founded_year: Int?,
    description: String,
    embedding: <F32; 384>
}

:create person {
    id: String
    =>
    name: String,
    title: String?,
    company_id: String?,
    description: String,
    embedding: <F32; 384>
}

:create product {
    id: String
    =>
    name: String,
    company_id: String,
    category: String,
    launch_date: String?,
    description: String,
    embedding: <F32; 384>
}

# Relationship entities (temporal keys)
:create partnership {
    company_a: String,
    company_b: String,
    timestamp: String
    =>
    partnership_type: String,
    description: String,
    value_usd: Float?,
    sentiment: String,
    embedding: <F32; 384>
}

:create funding {
    company_id: String,
    round_date: String
    =>
    round_type: String,
    amount_usd: Float?,
    lead_investor: String?,
    description: String,
    embedding: <F32; 384>
}
```

**Design notes**:
- `partnership` has a three-part composite key so the same pair of companies can have multiple partnerships over time
- `funding` uses `(company_id, round_date)` as key -- each company can only have one funding round per date
- All entities have 384-dim embeddings for semantic search
- Nullable fields use `?` suffix for optional data
</details>

### Task 1.2: Create Vector Indices

Write HNSW index creation statements for all five entity types.

<details>
<summary>Solution</summary>

```cozoscript
::hnsw create company:company_embedding_idx {
    dim: 384, m: 16, dtype: F32,
    fields: [embedding], distance: Cosine,
    ef_construction: 200
}

::hnsw create person:person_embedding_idx {
    dim: 384, m: 16, dtype: F32,
    fields: [embedding], distance: Cosine,
    ef_construction: 200
}

::hnsw create product:product_embedding_idx {
    dim: 384, m: 16, dtype: F32,
    fields: [embedding], distance: Cosine,
    ef_construction: 200
}

::hnsw create partnership:partnership_embedding_idx {
    dim: 384, m: 16, dtype: F32,
    fields: [embedding], distance: Cosine,
    ef_construction: 200
}

::hnsw create funding:funding_embedding_idx {
    dim: 384, m: 16, dtype: F32,
    fields: [embedding], distance: Cosine,
    ef_construction: 200
}
```
</details>

### Task 1.3: Create Standard Indices

Create useful standard indices for common query patterns:

<details>
<summary>Solution</summary>

```cozoscript
# Find people by company
::index create person:by_company {company_id, id}

# Find products by company
::index create product:by_company {company_id, id}

# Find funding by amount (for ranking)
::index create funding:by_date {round_date, company_id}
```
</details>

## Phase 2: Sample Data

### Task 2.1: Create Test Data

Create constant-rule queries to populate your schema with sample data. Here's a sample tech news article to extract from:

```
In January 2024, OpenAI announced a major partnership with Microsoft, extending their
cloud computing collaboration. Sam Altman, OpenAI's CEO, said the partnership would
focus on enterprise AI solutions. Satya Nadella, Microsoft's CEO, confirmed a $10 billion
investment.

Meanwhile, Google launched Gemini Ultra in February 2024, their most powerful AI model.
Sundar Pichai described it as a breakthrough in multimodal reasoning. Google also
partnered with Samsung to integrate Gemini into Galaxy devices.

In March 2024, Anthropic raised $2.75 billion in a Series D round led by Menlo Ventures,
with participation from Google. Dario Amodei, Anthropic's CEO, announced the launch of
Claude 3, competing directly with GPT-4.

Apple hired several ex-Google AI researchers in April 2024, signaling their entrance
into the generative AI race. Tim Cook mentioned AI integration across all Apple products
at the Spring event, where they unveiled Apple Intelligence for iOS 18.
```

Write insertion queries for at least:
- 5 companies
- 6 people
- 4 products
- 3 partnerships
- 2 funding events

Use dummy embeddings (384-dimensional zero vectors or random vectors).

<details>
<summary>Guided Solution (Partial)</summary>

```cozoscript
# Helper: generate a dummy 384-dim embedding
# In real code, use a proper embedding model

# Insert companies
?[id, name, industry, hq_city, founded_year, description, embedding] <- [
    ['openai', 'OpenAI', 'Artificial Intelligence', 'San Francisco', 2015,
     'AI research company known for GPT models and ChatGPT',
     vec(range(384))],
    ['microsoft', 'Microsoft', 'Technology', 'Redmond', 1975,
     'Multinational tech corporation, major OpenAI investor',
     vec(range(384))],
    ['google', 'Google', 'Technology', 'Mountain View', 1998,
     'Search and AI company, developer of Gemini models',
     vec(range(384))],
    ['anthropic', 'Anthropic', 'Artificial Intelligence', 'San Francisco', 2021,
     'AI safety company, creator of Claude models',
     vec(range(384))],
    ['apple', 'Apple', 'Technology', 'Cupertino', 1976,
     'Consumer electronics and software company entering AI',
     vec(range(384))]
]
:put company {id => name, industry, hq_city, founded_year, description, embedding}
```

```cozoscript
# Insert people
?[id, name, title, company_id, description, embedding] <- [
    ['sam_altman', 'Sam Altman', 'CEO', 'openai',
     'CEO of OpenAI, leading enterprise AI expansion',
     vec(range(384))],
    ['satya_nadella', 'Satya Nadella', 'CEO', 'microsoft',
     'CEO of Microsoft, orchestrated OpenAI investment',
     vec(range(384))],
    ['sundar_pichai', 'Sundar Pichai', 'CEO', 'google',
     'CEO of Google, overseeing Gemini AI development',
     vec(range(384))],
    ['dario_amodei', 'Dario Amodei', 'CEO', 'anthropic',
     'CEO and co-founder of Anthropic, AI safety researcher',
     vec(range(384))],
    ['tim_cook', 'Tim Cook', 'CEO', 'apple',
     'CEO of Apple, leading AI integration initiative',
     vec(range(384))]
]
:put person {id => name, title, company_id, description, embedding}
```

Continue with products, partnerships, and funding...
</details>

## Phase 3: Query Development

### Task 3.1: Basic Queries

Write queries for:
1. List all companies in the AI industry
2. Find all people at a specific company
3. Find all products launched in 2024

<details>
<summary>Solution</summary>

```cozoscript
# 1. AI companies
?[name, description] := *company{industry: 'Artificial Intelligence', name, description}

# 2. People at Google
?[name, title] := *person{company_id: 'google', name, title}

# 3. Products launched in 2024
?[name, company_id, launch_date] := *product{name, company_id, launch_date},
    starts_with(launch_date, '2024')
:order launch_date
```
</details>

### Task 3.2: Join Queries

Write queries that join multiple relations:
1. List all people with their company names
2. Find all products and their company's industry
3. Find all partnerships involving AI companies

<details>
<summary>Solution</summary>

```cozoscript
# 1. People with company names
?[person_name, person_title, company_name] :=
    *person{name: person_name, title: person_title, company_id},
    *company{id: company_id, name: company_name}
:order company_name, person_name

# 2. Products with company industry
?[product_name, category, company_name, industry] :=
    *product{name: product_name, category, company_id},
    *company{id: company_id, name: company_name, industry}
:order industry, product_name

# 3. Partnerships involving AI companies
?[company_a_name, company_b_name, partnership_type, timestamp] :=
    *partnership{company_a, company_b, partnership_type, timestamp},
    *company{id: company_a, name: company_a_name},
    *company{id: company_b, name: company_b_name},
    (
        *company{id: company_a, industry: 'Artificial Intelligence'}
        or *company{id: company_b, industry: 'Artificial Intelligence'}
    )
:order timestamp
```
</details>

### Task 3.3: Aggregation Queries

1. Count partnerships per company
2. Total funding raised per company
3. Average partnership sentiment by company pair

<details>
<summary>Solution</summary>

```cozoscript
# 1. Partnerships per company (counting both sides)
partner_count[company, count(other)] :=
    *partnership{company_a: company, company_b: other}
partner_count[company, count(other)] :=
    *partnership{company_a: other, company_b: company}
?[company, total_partnerships] := partner_count[company, total_partnerships]
:order -total_partnerships

# 2. Total funding per company
?[company_id, sum(amount_usd), count(round_date)] :=
    *funding{company_id, round_date, amount_usd},
    amount_usd != null
:order -sum(amount_usd)

# 3. Average sentiment score (map to numbers first)
# This requires mapping sentiment to a numeric value
sentiment_scores[company_a, company_b, score] :=
    *partnership{company_a, company_b, sentiment},
    sentiment == 'positive', score = 1.0
sentiment_scores[company_a, company_b, score] :=
    *partnership{company_a, company_b, sentiment},
    sentiment == 'neutral', score = 0.0
sentiment_scores[company_a, company_b, score] :=
    *partnership{company_a, company_b, sentiment},
    sentiment == 'negative', score = -1.0
?[company_a, company_b, mean(score)] := sentiment_scores[company_a, company_b, score]
```
</details>

### Task 3.4: Graph Queries

1. Find all companies connected to OpenAI through partnerships (1-hop)
2. Find companies reachable from OpenAI within 2 hops
3. Use PageRank to find the most influential company

<details>
<summary>Solution</summary>

```cozoscript
# 1. Direct partners of OpenAI
?[partner_name] :=
    *partnership{company_a: 'openai', company_b: partner_id},
    *company{id: partner_id, name: partner_name}
?[partner_name] :=
    *partnership{company_a: partner_id, company_b: 'openai'},
    *company{id: partner_id, name: partner_name}

# 2. Two-hop reachability
edges[a, b] := *partnership{company_a: a, company_b: b}
edges[a, b] := *partnership{company_a: b, company_b: a}

reachable[company] := edges['openai', company]
reachable[company] := reachable[intermediate], edges[intermediate, company], company != 'openai'

?[company_name] :=
    reachable[company_id],
    *company{id: company_id, name: company_name}

# 3. PageRank
edges[from, to] := *partnership{company_a: from, company_b: to}
edges[from, to] := *partnership{company_a: to, company_b: from}

?[company_name, score] <~ PageRank(edges[], damping_factor: 0.85)
company_name_lookup[id, name] := *company{id, name}
?[name, score] := ?[id, score], company_name_lookup[id, name]
:order -score
```
</details>

### Task 3.5: Vector Search

Find companies semantically similar to a given company using HNSW:

<details>
<summary>Solution</summary>

```cozoscript
# Find companies similar to Anthropic
?[dist, id, name, description] :=
    *company{id: 'anthropic', embedding: query_vec},
    ~company:company_embedding_idx{id, name, description |
        query: query_vec,
        k: 5,
        ef: 50,
        bind_distance: dist
    }
:order dist
```
</details>

## Phase 4: Temporal Tracking

### Task 4.1: Model Partnership Evolution

Insert multiple snapshots of the OpenAI-Microsoft partnership at different timestamps showing how it evolved:

```cozoscript
?[company_a, company_b, timestamp, partnership_type, description, value_usd, sentiment, embedding] <- [
    ['openai', 'microsoft', '2019-07', 'investment', 'Initial $1B investment', 1000000000.0, 'positive', vec(range(384))],
    ['openai', 'microsoft', '2023-01', 'investment', 'Extended partnership with $10B investment', 10000000000.0, 'positive', vec(range(384))],
    ['openai', 'microsoft', '2024-01', 'strategic', 'Enterprise AI solutions partnership', null, 'positive', vec(range(384))]
]
:put partnership {company_a, company_b, timestamp => partnership_type, description, value_usd, sentiment, embedding}
```

### Task 4.2: Query Temporal Evolution

Write a query showing the OpenAI-Microsoft partnership evolution:

```cozoscript
?[timestamp, partnership_type, description, value_usd] :=
    *partnership{
        company_a: 'openai', company_b: 'microsoft',
        timestamp, partnership_type, description, value_usd
    }
:order timestamp
```

### Task 4.3: Create a Trigger

Create a trigger on the `partnership` relation that logs all changes to a `partnership_log` relation:

<details>
<summary>Solution</summary>

```cozoscript
:create partnership_log {
    company_a: String,
    company_b: String,
    timestamp: String,
    logged_at: String
    =>
    old_type: String?,
    new_type: String?,
    action: String
}

::set_triggers partnership

on put {
    ?[company_a, company_b, timestamp, logged_at, old_type, new_type, action] :=
        _new[company_a, company_b, timestamp, new_type, _, _, _, _],
        logged_at = now(),
        old_type = null,
        action = 'upsert'
    :put partnership_log {
        company_a, company_b, timestamp, logged_at
        => old_type, new_type, action
    }
}

on rm {
    ?[company_a, company_b, timestamp, logged_at, old_type, new_type, action] :=
        _old[company_a, company_b, timestamp, old_type, _, _, _, _],
        logged_at = now(),
        new_type = null,
        action = 'remove'
    :put partnership_log {
        company_a, company_b, timestamp, logged_at
        => old_type, new_type, action
    }
}
```
</details>

## Phase 5: Write an Extraction Plugin

### Task 5.1: Simple Extractor

Write a JS extraction plugin (using the plugin descriptor contract) that:
1. Takes tech news text as input
2. Uses simple regex/heuristic extraction (no LLM)
3. Finds company names (capitalized multi-word sequences)
4. Finds dollar amounts ($X billion/million)
5. Returns structured JSON

Save it as `scripts/capstone_heuristic_extractor.js` in the ticket folder.

<details>
<summary>Solution</summary>

```javascript
module.exports = {
    apiVersion: "cozo.extractor/v1",
    kind: "extractor",
    id: "capstone.heuristic-extractor",
    name: "Capstone Heuristic Extractor",
    create: function(ctx) {
        return {
            run: function(input) {
                var text = input.transcript || "";

                // Extract company-like names (sequences of capitalized words)
                var companyPattern = /([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)/g;
                var companies = {};
                var match;
                while ((match = companyPattern.exec(text)) !== null) {
                    var name = match[1];
                    if (name.split(/\s+/).length <= 3) {
                        var id = name.toLowerCase().replace(/\s+/g, '_');
                        if (!companies[id]) {
                            companies[id] = { id: id, name: name, mentions: 0 };
                        }
                        companies[id].mentions++;
                    }
                }

                // Extract dollar amounts
                var moneyPattern = /\$([0-9,.]+)\s*(billion|million|trillion)/gi;
                var amounts = [];
                while ((match = moneyPattern.exec(text)) !== null) {
                    var value = parseFloat(match[1].replace(/,/g, ''));
                    var unit = match[2].toLowerCase();
                    var multiplier = unit === 'trillion' ? 1e12 :
                                     unit === 'billion' ? 1e9 : 1e6;
                    amounts.push({
                        raw: match[0],
                        value_usd: value * multiplier,
                        context: text.substring(
                            Math.max(0, match.index - 50),
                            Math.min(text.length, match.index + match[0].length + 50)
                        )
                    });
                }

                // Extract dates
                var datePattern = /(?:January|February|March|April|May|June|July|August|September|October|November|December)\s+(\d{4})/g;
                var dates = [];
                while ((match = datePattern.exec(text)) !== null) {
                    dates.push(match[0]);
                }

                var companyList = [];
                for (var id in companies) {
                    if (companies[id].mentions >= 2) {
                        companyList.push(companies[id]);
                    }
                }
                companyList.sort(function(a, b) { return b.mentions - a.mentions; });

                return JSON.stringify({
                    companies: companyList,
                    financial_amounts: amounts,
                    dates_mentioned: dates,
                    stats: {
                        total_words: text.split(/\s+/).length,
                        companies_found: companyList.length,
                        amounts_found: amounts.length
                    }
                }, null, 2);
            }
        };
    }
};
```
</details>

### Task 5.2: LLM-Powered Extractor

Write a JS plugin that uses the Geppetto module to call an LLM for extraction. Use the structured output config to enforce your tech news schema.

<details>
<summary>Solution Sketch</summary>

```javascript
var gp = require("geppetto");
var { defineExtractorPlugin, wrapExtractorRun } = require("geppetto/plugins");

var TECH_NEWS_SCHEMA = {
    type: "object",
    properties: {
        companies: {
            type: "array",
            items: {
                type: "object",
                properties: {
                    id: { type: "string" },
                    name: { type: "string" },
                    industry: { type: "string" },
                    description: { type: "string" }
                },
                required: ["id", "name", "industry", "description"]
            }
        },
        people: {
            type: "array",
            items: {
                type: "object",
                properties: {
                    id: { type: "string" },
                    name: { type: "string" },
                    title: { type: "string" },
                    company_id: { type: "string" }
                },
                required: ["id", "name"]
            }
        },
        partnerships: {
            type: "array",
            items: {
                type: "object",
                properties: {
                    company_a: { type: "string" },
                    company_b: { type: "string" },
                    timestamp: { type: "string" },
                    type: { type: "string" },
                    description: { type: "string" },
                    sentiment: { type: "string", enum: ["positive", "negative", "neutral"] }
                },
                required: ["company_a", "company_b", "type"]
            }
        }
    },
    required: ["companies", "people", "partnerships"]
};

module.exports = defineExtractorPlugin({
    id: "capstone.tech-news-extractor",
    name: "Capstone Tech News Extractor",
    create: function(ctx) {
        return {
            run: wrapExtractorRun(function(input) {
                var engine = input.engineOptions
                    ? gp.engines.fromConfig(input.engineOptions)
                    : gp.engines.fromProfile(input.profile || "", {
                        timeoutMs: input.timeoutMs
                    });

                var session = gp.createBuilder()
                    .withEngine(engine)
                    .useGoMiddleware("systemPrompt", {
                        prompt: "You are an expert at extracting structured data from technology news articles. Extract all companies, people, products, partnerships, and funding events."
                    })
                    .buildSession();

                var seed = gp.turns.newTurn({
                    blocks: [gp.turns.newUserBlock("Extract from:\n\n" + input.transcript)],
                    data: {
                        [gp.consts.TurnDataKeys.STRUCTURED_OUTPUT_CONFIG]: {
                            mode: "json_schema",
                            name: "tech_news_extraction",
                            schema: TECH_NEWS_SCHEMA,
                            strict: true,
                            require_valid: true
                        }
                    }
                });

                var result = session.run(seed, { timeoutMs: input.timeoutMs });
                var blocks = (result && result.blocks) ? result.blocks : [];
                var text = blocks
                    .filter(function(b) { return b && b.payload && b.payload.text; })
                    .map(function(b) { return b.payload.text; })
                    .join("\n").trim();

                return JSON.parse(text);
            })
        };
    }
});
```
</details>

## Phase 6: Putting It All Together

### Task 6.1: Full Pipeline Script

Write a Python script that:
1. Reads a text file
2. Calls the extraction (simulated with hardcoded JSON for this exercise)
3. Creates the CozoDB schema
4. Inserts extracted data
5. Creates HNSW indices
6. Runs 10 analytical queries

This is modeled after `cozo_demo.py` but for your tech news domain.

### Task 6.2: Advanced Analytics

Write queries to answer:
1. Which company has the most partnerships?
2. Who are the most connected CEOs (via their companies' partnerships)?
3. What is the total investment flowing into AI companies?
4. Find the shortest path between any two companies through partnerships
5. Which companies form tight clusters (community detection)?
6. Find companies that are both competitors and partners
7. Build a timeline of all events ordered chronologically
8. Find similar companies using vector search
9. Detect companies that raised funding AND announced partnerships in the same month
10. Compute a "momentum score" for each company (partnerships + funding + products in last 6 months)

### Task 6.3: Final Challenge

Extend your system to handle **multiple articles over time**. Each article has a publication date. Your system should:
1. Track when information was first extracted (transaction time)
2. Track when events actually happened (valid time)
3. Show how the knowledge graph grows as more articles are processed
4. Detect contradictions between articles

---

## Evaluation Checklist

When you've completed the capstone, verify:

- [ ] Schema has 5+ relations with appropriate keys and values
- [ ] HNSW indices exist for all embedding columns
- [ ] At least one standard index for a common query pattern
- [ ] Sample data covers all entity types
- [ ] Basic queries (filter, join, sort) all work
- [ ] Aggregation queries produce meaningful results
- [ ] At least one recursive graph traversal query
- [ ] At least one graph algorithm (PageRank, community detection, etc.)
- [ ] Temporal data with multiple timestamps per entity pair
- [ ] At least one trigger for change tracking
- [ ] At least one JS extraction plugin (heuristic or LLM-powered)
- [ ] Vector search query using HNSW index

## Congratulations!

If you've completed all 7 modules, you now understand:

1. **CozoDB's Datalog query language** -- rules, atoms, unification, negation
2. **Schema design** -- keys, values, types, indices (standard, HNSW, FTS)
3. **Advanced queries** -- aggregation, recursion, graph algorithms (PageRank, BFS, etc.)
4. **Temporal modeling** -- time travel, triggers, application-level timestamps
5. **Entity extraction** -- LLM prompting, structured output, RDF vs entity-relation models
6. **JavaScript/Go integration** -- Goja runtime, module system, plugin descriptors
7. **End-to-end pipelines** -- from raw text to queryable knowledge graphs

### Where to Go Next

- **CozoDB docs**: https://docs.cozodb.org/en/latest/
- **Goja repo**: https://github.com/dop251/goja
- **Geppetto framework**: https://github.com/go-go-golems/geppetto
- **This project's repo**: Explore the subsystems you haven't studied yet
- **Challenge**: Add a web frontend using the `cozo-relationship-explorer` as a model
