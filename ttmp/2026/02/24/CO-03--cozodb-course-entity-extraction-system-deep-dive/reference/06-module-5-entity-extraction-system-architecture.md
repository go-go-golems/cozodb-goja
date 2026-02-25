---
Title: "Module 5 - Entity Extraction System Architecture"
Ticket: CO-03
Status: active
Topics:
    - cozodb
    - course
    - entity-extraction
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/extract_relationships.py:Python extraction pipeline"
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo_demo.py:CozoDB schema + query demo"
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-relationship-js-runner/main.go:Go CLI runner for JS extractors"
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-relationship-js-runner/plugin_loader.go:Plugin descriptor loading protocol"
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/relation_extractor_template.js:JS extractor plugin"
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/cozo-relationship-js-runner/scripts/lib/relationship_extractor_factory.js:Extractor factory with LLM integration"
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/demo_text.txt:Sample narrative text"
    - "/home/manuel/workspaces/2026-02-24/cozodb-goja-init/2026-02-18--cozodb-extraction/manus-fact-extraction/fact-extraction-go/go-extractor/pkg/types/types.go:Go data types for RDF triples"
ExternalSources:
    - https://docs.cozodb.org/en/latest/
Summary: "Architecture walkthrough of the entity extraction pipeline: from text to CozoDB"
LastUpdated: 2026-02-24T18:45:00-05:00
WhatFor: "Understand how the system extracts entities from text and stores them in CozoDB"
WhenToUse: "When you need to understand the full pipeline architecture"
---

# Module 5: Entity Extraction System Architecture

## Learning Objectives

By the end of this module you will:
- Understand the end-to-end extraction pipeline
- Know the four entity types and their data model
- Understand how LLMs are used for structured extraction
- Understand the RDF triple model vs the entity-relation model
- Be able to trace data flow from raw text to CozoDB queries
- Understand the different subsystems in this repository

## 5.1 The Big Picture

This repository implements a **knowledge graph extraction system**: given narrative text (meeting transcripts, news articles, documents), it uses LLMs to extract structured entities and relationships, stores them in CozoDB, and enables complex graph queries over the extracted data.

```
                    +-----------------+
                    |  Narrative Text |
                    |  (transcripts,  |
                    |   documents)    |
                    +--------+--------+
                             |
                    +--------v--------+
                    |   LLM Extraction |
                    |  (GPT-4.1-mini) |
                    |  Structured JSON |
                    +--------+--------+
                             |
                    +--------v--------+
                    |   Parse + Embed  |
                    |  (validate JSON, |
                    |   generate 384d  |
                    |   embeddings)    |
                    +--------+--------+
                             |
                    +--------v--------+
                    |   Store in       |
                    |   CozoDB         |
                    |   (4 relations + |
                    |    HNSW indices)  |
                    +--------+--------+
                             |
              +--------------+--------------+
              |              |              |
     +--------v---+   +-----v------+  +----v-------+
     | Graph       |   | Vector     |  | Temporal   |
     | Queries     |   | Search     |  | Queries    |
     | (multi-hop, |   | (similar   |  | (evolution |
     |  patterns)  |   |  entities)  |  |  tracking) |
     +-------------+   +------------+  +------------+
```

## 5.2 The Four Entity Types

The extraction system models four types of entities. These were designed to capture the richness of human relationships in narrative text.

### Person

A mentioned individual. Simple primary key, rich metadata.

```cozoscript
:create person {
    id: String                    # Primary key: "sarah_martinez"
    =>
    name: String,                 # Display name: "Sarah Martinez"
    description: String,          # Context: "Senior scientist at the research lab"
    first_mentioned: String,      # When: "2023-01"
    embedding: <F32; 384>         # Semantic vector for similarity search
}
```

**Design decisions**:
- `id` is a normalized slug (lowercase, underscores): `"sarah_martinez"`
- `description` captures the LLM's understanding of who this person is *in context*
- `embedding` is computed from `name + description` for semantic search

### Relationship

A connection between two people, tracked over time. **Four-part composite key**.

```cozoscript
:create relationship {
    id: String,                   # Unique relationship ID
    from_person: String,          # Source person ID
    to_person: String,            # Target person ID
    timestamp: String             # When observed: "2023-06"
    =>
    relationship_type: String,    # "friend", "colleague", "mentor", "rival"
    description: String,          # "Sarah and David had a heated argument..."
    sentiment: String,            # "positive", "negative", "neutral"
    strength: Float,              # 0.0 to 1.0
    embedding: <F32; 384>         # Semantic vector
}
```

**Design decisions**:
- `timestamp` is part of the KEY, not a value. This means the same pair of people can have multiple relationship snapshots at different times.
- This enables temporal tracking: Sarah and David go from "colleague" (positive, 0.7) in Jan 2023 to "rival" (negative, 0.3) in Jun 2023, back to "close colleague" (positive, 0.9) in Aug 2023.
- `sentiment` is a simple enum, not a float -- easier for LLMs to produce reliably.
- `strength` is a float for nuance.

### Behavior

A specific action performed by a person.

```cozoscript
:create behavior {
    id: String,                   # Unique behavior ID
    person_id: String,            # Who performed it
    timestamp: String             # When
    =>
    action: String,               # "met", "argued", "collaborated", "presented"
    description: String,          # Detailed context
    location: String?,            # Where (nullable)
    involves_persons: String?,    # JSON-encoded list of other person IDs
    embedding: <F32; 384>
}
```

**Design decisions**:
- `involves_persons` is a JSON-encoded string, not a proper list relation. This is a pragmatic simplification -- a normalized design would use a separate `behavior_participant` relation.
- `location` is nullable because not all behaviors have a place.

### Event

A multi-person occurrence.

```cozoscript
:create event {
    id: String,                   # Unique event ID
    timestamp: String             # When
    =>
    title: String,                # "Conference in Boston"
    description: String,          # Full description
    participants: String,         # JSON-encoded list of person IDs
    location: String?,            # Where
    embedding: <F32; 384>
}
```

## 5.3 How LLM Extraction Works

The extraction is driven by a carefully crafted system prompt that tells the LLM exactly what to extract and how to format it.

### The System Prompt

From `main.go:36-51`:

```
You are an expert at extracting structured relationship data from narrative text.

Extract the following from the text:
1. **Persons**: All individuals mentioned
2. **Relationships**: Connections between people
3. **Behaviors**: Specific actions or behaviors performed by individuals
4. **Events**: Occurrences involving multiple people

Guidelines:
- Use lowercase IDs with underscores (for example, "sarah_martinez")
- Infer timestamps from context (YYYY-MM format)
- For relationships, assess sentiment (positive/negative/neutral) and strength (0.0-1.0)
- Include as much detail as possible in descriptions
- Link behaviors and events to specific persons using their IDs
```

### Structured Output (JSON Schema)

The extraction uses **JSON Schema enforcement** to ensure the LLM produces valid, parseable output. The schema is defined in `relationship_constants.js` and passed to the LLM as a structured output configuration.

The Python version uses **Pydantic models** (`extract_relationships.py:18-60`) which automatically generate the JSON schema:

```python
class ExtractionResult(BaseModel):
    persons: List[Person]
    relationships: List[Relationship]
    behaviors: List[Behavior]
    events: List[Event]
```

The LLM returns structured JSON like:
```json
{
  "persons": [
    {
      "id": "sarah_martinez",
      "name": "Sarah Martinez",
      "description": "Senior scientist at the research lab",
      "first_mentioned": "2023-01"
    }
  ],
  "relationships": [
    {
      "id": "sarah_david_colleague_2023_01",
      "from_person": "sarah_martinez",
      "to_person": "david_chen",
      "timestamp": "2023-01",
      "relationship_type": "colleague",
      "description": "Sarah quickly became friends with David...",
      "sentiment": "positive",
      "strength": 0.7
    }
  ],
  "behaviors": [...],
  "events": [...]
}
```

### Embedding Generation

After extraction, each entity gets a 384-dimensional embedding vector for semantic similarity search:

```python
# From extract_relationships.py:152-171
person_texts = [f"{p.name}: {p.description}" for p in result.persons]
person_embeddings = generate_embeddings(person_texts)

relationship_texts = [
    f"{r.relationship_type} between {r.from_person} and {r.to_person}: {r.description}"
    for r in result.relationships
]
relationship_embeddings = generate_embeddings(relationship_texts)
```

In the demo, embeddings are deterministic random vectors. In production, you'd use a real embedding model (sentence-transformers, OpenAI embeddings).

## 5.4 The Four Subsystems

This repository evolved through several iterations. Understanding all four helps you see the design space:

### Subsystem 1: Python Demo (Simplest)

**Files**: `extract_relationships.py`, `cozo_demo.py`, `cozo_advanced_demo.py`

The simplest pipeline:
1. Read text from file
2. Call OpenAI with Pydantic schema
3. Generate embeddings
4. Save to JSON
5. Load into CozoDB (in-memory)
6. Run demo queries

**Good for**: Learning CozoDB basics, understanding the data model.

### Subsystem 2: Full-Stack Web App

**Directory**: `cozo-relationship-explorer/`

A React + tRPC + Python full-stack app with:
- Web UI for text input and visualization
- SQLite-backed persistent CozoDB
- Real-time extraction status

**Good for**: Building user-facing tools.

### Subsystem 3: Go Fact Extraction + Cayley

**Directory**: `manus-fact-extraction/fact-extraction-go/`

Uses a different data model -- **RDF triples** (Actor, Action, Target):

```go
// From go-extractor/pkg/types/types.go
type RDFTriple struct {
    Actor           string
    Action          string
    Target          string
    ExplicitTopic   string
    ImplicitTopic   string
    Tags            []string
    Timestamp       *string
    Location        *string
    ActorLikelyType *string
}
```

This subsystem stores triples in SQLite and loads them into **Cayley** (a Go graph database) for Gizmo-based graph queries.

**Good for**: Understanding the RDF triple model vs. the entity-relation model.

### Subsystem 4: JS Plugin Runner (Most Sophisticated)

**Directory**: `cozo-relationship-js-runner/`

This is the **flagship system**. It:
1. Uses Goja (Go JS runtime) to run JavaScript extraction plugins
2. Integrates Geppetto middleware for LLM access
3. Records metrics to SQLite
4. Uses a plugin descriptor pattern for extensibility

## 5.5 The JS Plugin Runner in Detail

### Architecture

```
CLI (Cobra) → Parse args → Resolve Pinocchio profile
    ↓
Create Goja VM + Event Loop
    ↓
Register modules:
  - geppetto (LLM middleware)
  - runnerdb / database (SQLite access)
  - console (logging)
    ↓
Set global variables:
  - RELATIONSHIP_TRANSCRIPT
  - RELATIONSHIP_PROMPT
  - RELATIONSHIP_PROFILE
  - RELATIONSHIP_ENGINE_OPTIONS
  - RELATIONSHIP_TIMEOUT_MS
  - RELATIONSHIP_RUN_ID
  - RELATIONSHIP_SCRIPT_ROOT
  - RELATIONSHIP_SCRIPT_DB_DSN
    ↓
Load JS plugin via require()
    ↓
Validate plugin descriptor:
  - apiVersion == "cozo.extractor/v1"
  - kind == "extractor"
  - id, name present
  - create() function exists
    ↓
Call descriptor.create(hostContext)
    ↓
Call instance.run(input, options)
    ↓
Parse return value (JSON string or object)
    ↓
Record metrics + Output JSON
```

### The Plugin Contract

Every extraction plugin must export a descriptor object:

```javascript
// From relation_extractor_template.js
module.exports = defineExtractorPlugin({
  id: "cozo.relationship-extractor.base",
  name: "Cozo Relationship Extractor (Base)",
  create() {
    const extractor = createRelationshipExtractor({
      variantName: "base",
    });
    return {
      run: wrapExtractorRun(
        (input) => extractor.extractRelations(input.transcript, input)
      ),
    };
  },
});
```

The Go host calls this contract as defined in `plugin_loader.go:74-158`:

1. `require(scriptPath)` -- load the JS module
2. Validate metadata: `apiVersion`, `kind`, `id`, `name`
3. Call `descriptor.create(hostContext)` -- get a plugin instance
4. Call `instance.run(input, options)` -- execute extraction
5. Parse the return value as JSON

### The Extractor Factory

The `createRelationshipExtractor` factory (`relationship_extractor_factory.js:100-213`) orchestrates:

1. **Engine creation** -- from config or Pinocchio profile
2. **Builder chain** -- Geppetto builder with middleware
3. **Session creation** -- `builder.buildSession()`
4. **Turn execution** -- create a seed turn with the transcript, run it
5. **Output parsing** -- extract assistant text from blocks, parse as JSON

Key code path:
```javascript
const engine = gp.engines.fromConfig(engineOptions);
let builder = gp.createBuilder(builderOptions)
    .withEngine(engine)
    .useGoMiddleware("systemPrompt", { prompt });
const session = builder.buildSession();

const seed = gp.turns.newTurn({
    blocks: [gp.turns.newUserBlock(buildPrompt(prompt, transcriptText))],
    data: {
        [gp.consts.TurnDataKeys.STRUCTURED_OUTPUT_CONFIG]: structuredOutput,
    },
});

const out = session.run(seed, { timeoutMs, tags });
const assistantText = resolveAssistantText(out);
return parseLLMOutput(assistantText);
```

### Running the Pipeline

```bash
# Basic usage
cozo-relationship-js-runner extract \
    scripts/relation_extractor_template.js \
    /path/to/transcript.txt \
    --ai-api-type openai \
    --ai-engine gpt-4.1-mini

# With recording and streaming
cozo-relationship-js-runner extract \
    scripts/relation_extractor_template.js \
    /path/to/transcript.txt \
    --stream-events \
    --record \
    --record-db .cozo-runner/runs.sqlite \
    --pretty
```

## 5.6 The RDF Triple Model (Subsystem 3)

The `manus-fact-extraction` subsystem uses a simpler, flatter model:

| Entity-Relation Model (Subsystem 1,2,4) | RDF Triple Model (Subsystem 3) |
|---|---|
| 4 entity types (Person, Relationship, Behavior, Event) | 1 triple type (Actor, Action, Target) |
| Rich schema per type | Flat, uniform structure |
| Explicit temporal columns | Optional timestamp field |
| Structured sentiment/strength | Tags and topics |
| Better for complex queries | Better for flexibility |
| CozoDB | Cayley / SQLite |

Example triple: `("Jeffrey Epstein", "attended event with", "Donald Trump")`

The RDF approach is simpler but loses structure. The entity-relation approach is richer but requires more schema design.

## 5.7 From Extraction to CozoDB

After extraction, data is inserted into CozoDB using the `:put` pattern:

```python
# From cozo_demo.py:26-35
for p in data['persons']:
    db.run("""
        ?[id, name, description, first_mentioned, embedding]
            <- [[$id, $name, $description, $first_mentioned, $embedding]]
        :put person {id => name, description, first_mentioned, embedding}
    """, {
        "id": p['id'],
        "name": p['name'],
        "description": p['description'],
        "first_mentioned": p['first_mentioned'],
        "embedding": p['embedding']
    })
```

The pattern is always:
1. Define a constant rule `?[...]` with parameter bindings `$param`
2. Use `:put` to upsert into the stored relation
3. Pass parameters as a dictionary from the host language

### HNSW Index Creation

After loading data, vector indices are created:

```python
# From cozo_demo.py:149-158
db.run("""
    ::hnsw create person:person_embedding_idx {
        dim: 384,
        m: 16,
        dtype: F32,
        fields: [embedding],
        distance: Cosine,
        ef_construction: 200
    }
""")
```

## 5.8 Sample Data Walkthrough

The `demo_text.txt` describes a research lab team over 14 months (Jan 2023 - Feb 2024):

**People**: Sarah Martinez, David Chen, Elena Rodriguez, Michael Thompson, Aisha Patel

**Timeline**:
- **Jan 2023**: Sarah joins, befriends David
- **Mar 2023**: Conference in Boston, meet Elena
- **Jun 2023**: Sarah-David conflict
- **Aug 2023**: Reconciliation (Elena facilitates)
- **Sep 2023**: Michael leaves for startup
- **Oct 2023**: Aisha joins
- **Nov 2023**: Paper accepted, celebration
- **Dec 2023**: Sarah promoted, advocates for David
- **Jan 2024**: Major grant received
- **Feb 2024**: Elena announces retirement

The extraction produces:
- 5 persons
- 9+ relationships (multiple per pair at different timestamps)
- 24+ behaviors
- 4+ events

---

## Exercises

### Exercise 5.1: Read the Demo Text

Read `demo_text.txt` and manually identify:
1. All 5 persons with their roles
2. At least 5 relationships with sentiment
3. At least 3 events

Compare your extraction with what the LLM produces (see `extracted_data.json`).

### Exercise 5.2: Trace the Data Flow

Starting from `extract_relationships.py`, trace the flow:
1. Where is the system prompt defined?
2. How is structured output enforced?
3. How are embeddings generated?
4. How is data inserted into CozoDB?

### Exercise 5.3: Design a Schema

Design CozoDB schemas for a **company org chart** extraction system:
- Employees (with roles, departments, seniority)
- Reporting relationships (who reports to whom)
- Projects (with team members and timelines)
- Meetings (with attendees and outcomes)

Write the `:create` statements.

<details>
<summary>Solution</summary>

```cozoscript
:create employee {
    id: String
    =>
    name: String,
    role: String,
    department: String,
    seniority_level: Int,
    start_date: String,
    embedding: <F32; 384>
}

:create reports_to {
    employee_id: String,
    manager_id: String,
    effective_date: String
    =>
    relationship_type: String,
    notes: String?
}

:create project {
    id: String
    =>
    name: String,
    description: String,
    start_date: String,
    end_date: String?,
    status: String
}

:create project_member {
    project_id: String,
    employee_id: String
    =>
    role_in_project: String,
    joined_date: String
}

:create meeting {
    id: String,
    timestamp: String
    =>
    title: String,
    description: String,
    outcomes: String?,
    location: String?
}

:create meeting_attendee {
    meeting_id: String,
    employee_id: String
    =>
    role_in_meeting: String
}
```
</details>

### Exercise 5.4: Prompt Engineering

Write a system prompt for extracting entities from **news articles** about technology companies. Your prompt should extract:
- Companies (name, industry, valuation)
- People (name, role at company)
- Products (name, category, company)
- Partnerships/Acquisitions (between companies)

<details>
<summary>Solution</summary>

```
You are an expert at extracting structured information from technology news articles.

Extract the following entities:
1. **Companies**: All companies mentioned (tech, finance, etc.)
2. **People**: Individuals mentioned with their roles
3. **Products**: Named products, services, or platforms
4. **Relationships**: Business relationships between companies

Guidelines:
- Use lowercase IDs with underscores (e.g., "openai", "sam_altman")
- For companies: include industry, estimated valuation if mentioned, headquarters
- For people: include their title/role and which company they belong to
- For products: include the category and owning company
- For relationships: classify as "partnership", "acquisition", "investment", "competition", "lawsuit"
- Include timestamps in YYYY-MM format when available
- Assess relationship sentiment and strategic importance (0.0-1.0)
```
</details>

### Exercise 5.5: Plugin Contract

Write a minimal JS plugin descriptor that:
1. Has apiVersion `"cozo.extractor/v1"`
2. Has kind `"extractor"`
3. Has a unique id and name
4. Has a `create()` function that returns an object with a `run()` method
5. The `run()` method should return a simple JSON object with a "message" field

<details>
<summary>Solution</summary>

```javascript
module.exports = {
    apiVersion: "cozo.extractor/v1",
    kind: "extractor",
    id: "my-first-extractor",
    name: "My First Extractor",
    create(hostContext) {
        return {
            run(input, options) {
                return JSON.stringify({
                    message: "Hello from my extractor!",
                    transcriptLength: (input.transcript || "").length,
                    timestamp: new Date().toISOString()
                });
            }
        };
    }
};
```
</details>

### Exercise 5.6: Compare Models

Given this text:
> "Alice founded TechCorp in 2020. Bob, her CTO, led the development of their flagship product DataFlow. In 2023, TechCorp acquired SmallStartup, which was founded by Charlie."

Extract this text using BOTH models:
1. Entity-Relation model (Persons, Relationships, Behaviors, Events)
2. RDF Triple model (Actor, Action, Target triples)

Which model captures more information? Which is simpler?

<details>
<summary>Solution</summary>

**Entity-Relation Model**:
```json
{
  "persons": [
    {"id": "alice", "name": "Alice", "description": "Founder of TechCorp"},
    {"id": "bob", "name": "Bob", "description": "CTO of TechCorp"},
    {"id": "charlie", "name": "Charlie", "description": "Founder of SmallStartup"}
  ],
  "relationships": [
    {"from_person": "alice", "to_person": "bob", "relationship_type": "employer",
     "sentiment": "positive", "strength": 0.9},
    {"from_person": "alice", "to_person": "charlie", "relationship_type": "acquirer",
     "sentiment": "neutral", "strength": 0.6}
  ],
  "events": [
    {"title": "TechCorp Founded", "timestamp": "2020-01", "participants": ["alice"]},
    {"title": "TechCorp acquires SmallStartup", "timestamp": "2023-01",
     "participants": ["alice", "charlie"]}
  ]
}
```

**RDF Triples**:
```
("Alice", "founded", "TechCorp")
("Bob", "is CTO of", "TechCorp")
("Bob", "led development of", "DataFlow")
("TechCorp", "acquired", "SmallStartup")
("Charlie", "founded", "SmallStartup")
```

The entity-relation model captures sentiment and strength. The RDF model is flatter but captures all actions naturally.
</details>

---

## Key Takeaways

1. The extraction pipeline has clear stages: **Text -> LLM -> Parse -> Embed -> Store -> Query**
2. **Four entity types** (Person, Relationship, Behavior, Event) capture relationship networks
3. **Composite keys with timestamps** enable temporal tracking
4. **HNSW vector indices** enable semantic similarity search
5. The **plugin descriptor pattern** makes the extraction system extensible
6. Both **entity-relation** and **RDF triple** models have trade-offs

## Next Module

In **Module 6: JavaScript Integration with Goja**, you'll dive deeper into how the Go runtime embeds JavaScript, how modules work, and how to write your own extraction plugins.
