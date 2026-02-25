---
Title: Models File Hierarchy
Ticket: CO-04
Status: active
Topics:
    - cozodb
    - tui
    - bubbletea
    - go
    - goja
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cozo-relationship-js-runner/main.go
      Note: Existing Go runner - TUI will share module/deps pattern
    - Path: cozo-relationship-js-runner/plugin_loader.go
      Note: Plugin loading code reused by extraction screen
    - Path: cozo-relationship-js-runner/go.mod
      Note: Existing deps (goja, geppetto, glazed, go-go-goja)
ExternalSources:
    - https://github.com/charmbracelet/bubbletea
    - https://github.com/charmbracelet/bubbles
    - https://github.com/charmbracelet/lipgloss
Summary: "Concrete Go file/directory layout with model struct sketches for every file"
LastUpdated: 2026-02-24T20:05:00-05:00
WhatFor: "Blueprint for creating the TUI project files"
WhenToUse: "When starting implementation, create files in this order and with these contents"
---

# Models File Hierarchy

## Full Tree

```
cozo-tui/
├── main.go
├── go.mod
│
├── internal/
│   ├── app/
│   │   ├── model.go
│   │   ├── keys.go
│   │   ├── navigation.go
│   │   └── styles.go
│   │
│   ├── domain/
│   │   ├── person.go
│   │   ├── relationship.go
│   │   ├── behavior.go
│   │   ├── event.go
│   │   └── extraction.go
│   │
│   ├── db/
│   │   ├── cozo.go
│   │   ├── queries.go
│   │   └── params.go
│   │
│   ├── embedding/
│   │   └── geppetto.go
│   │
│   ├── screens/
│   │   ├── screen.go
│   │   │
│   │   ├── dashboard/
│   │   │   ├── model.go
│   │   │   ├── commands.go
│   │   │   └── view.go
│   │   │
│   │   ├── people/
│   │   │   ├── model.go
│   │   │   ├── commands.go
│   │   │   └── view.go
│   │   │
│   │   ├── relationships/
│   │   │   ├── model.go
│   │   │   ├── commands.go
│   │   │   └── view.go
│   │   │
│   │   ├── evolution/
│   │   │   ├── model.go
│   │   │   ├── commands.go
│   │   │   ├── chart.go
│   │   │   └── view.go
│   │   │
│   │   ├── network/
│   │   │   ├── model.go
│   │   │   ├── commands.go
│   │   │   ├── layout.go
│   │   │   ├── render.go
│   │   │   └── view.go
│   │   │
│   │   ├── timeline/
│   │   │   ├── model.go
│   │   │   ├── commands.go
│   │   │   └── view.go
│   │   │
│   │   ├── query/
│   │   │   ├── model.go
│   │   │   ├── commands.go
│   │   │   ├── history.go
│   │   │   └── view.go
│   │   │
│   │   ├── extraction/
│   │   │   ├── model.go
│   │   │   ├── commands.go
│   │   │   ├── plugins.go
│   │   │   └── view.go
│   │   │
│   │   └── vector/
│   │       ├── model.go
│   │       ├── commands.go
│   │       └── view.go
│   │
│   └── components/
│       ├── statusbar/
│       │   └── model.go
│       ├── help/
│       │   └── model.go
│       ├── datatable/
│       │   └── model.go
│       ├── preview/
│       │   └── model.go
│       └── filterprompt/
│           └── model.go
│
└── scripts/
    └── (extractor plugins live here at runtime)
```

---

## File-by-File Model Sketches

### `main.go`

Entry point. Parses flags, opens CozoDB, creates the app model, runs `tea.NewProgram`.

```yaml
responsibilities:
  - parse --db flag (path to cozo.db)
  - parse --scripts flag (path to extractor plugins dir)
  - open CozoDB connection via db.Open()
  - create app.New(db, scriptsDir)
  - run tea.NewProgram(app, tea.WithAltScreen())
  - defer db.Close()

flags:
  --db: string       # default: "./cozo.db"
  --scripts: string  # default: "./scripts"
  --log-level: string
```

### `go.mod`

```yaml
module: github.com/wesen/cozo-tui

deps:
  # TUI
  - github.com/charmbracelet/bubbletea
  - github.com/charmbracelet/bubbles
  - github.com/charmbracelet/lipgloss

  # Existing (shared with js-runner)
  - github.com/dop251/goja
  - github.com/go-go-golems/geppetto
  - github.com/go-go-golems/go-go-goja

  # CozoDB
  - github.com/nicholasgasior/gcozo   # or whichever cozo-go binding

  # Logging
  - github.com/rs/zerolog
```

---

## `internal/app/` — Application Shell

### `app/model.go`

The root bubbletea model. Owns all screens, routes messages, manages navigation.

```yaml
struct: App
fields:
  activeScreen: ScreenID          # enum: dashboard|people|relationships|...
  screens:                        # lazy-initialized map
    dashboard: *dashboard.Model
    people: *people.Model
    relationships: *relationships.Model
    evolution: *evolution.Model
    network: *network.Model
    timeline: *timeline.Model
    query: *query.Model
    extraction: *extraction.Model
    vector: *vector.Model
  db: *db.DB
  scriptsDir: string
  statusBar: statusbar.Model
  help: help.Model
  width: int
  height: int
  err: error                      # last global error

methods:
  Init: |
    returns tea.Batch(
      statusBar.Init(),
      activeScreen.Init(),
    )
  Update: |
    1. handle tea.WindowSizeMsg -> resize all
    2. handle tea.KeyMsg for globals (F1-F8, ?, q)
    3. handle NavigateMsg -> switch activeScreen, call Init on new screen
    4. handle GlobalErrorMsg -> set err
    5. delegate to activeScreen.Update(msg)
  View: |
    lipgloss.JoinVertical(
      activeScreen.View(),
      statusBar.View(),
    )
    if help.Visible -> overlay help on top
    if err != nil -> overlay error toast
```

### `app/keys.go`

Global keybindings shared across all screens.

```yaml
keymap: GlobalKeyMap
bindings:
  F1: NavigateMsg{Screen: Dashboard}
  F2: NavigateMsg{Screen: People}
  F3: NavigateMsg{Screen: Relationships}
  F4: NavigateMsg{Screen: Timeline}
  F5: NavigateMsg{Screen: Network}
  F6: NavigateMsg{Screen: Query}
  F7: NavigateMsg{Screen: Extraction}
  F8: NavigateMsg{Screen: Vector}
  "?": ToggleHelpMsg
  q: tea.Quit
  ctrl+c: tea.Quit
```

### `app/navigation.go`

Screen ID enum, NavigateMsg type, and the routing logic.

```yaml
type: ScreenID
values: [dashboard, people, relationships, evolution, network,
         timeline, query, extraction, vector]

msgs:
  NavigateMsg:
    screen: ScreenID
    params: map[string]any       # e.g. {personID: "sarah_chen"}

  NavigateBackMsg: {}            # pop screen stack

state:
  screenStack: []ScreenID        # for back navigation
  # push on navigate, pop on NavigateBackMsg
```

### `app/styles.go`

Lipgloss style constants shared across all views.

```yaml
styles:
  Title:         bold, foreground(#7D56F4)
  Subtitle:      foreground(#ADADAD)
  ActiveTab:     bold, background(#7D56F4), foreground(#FFFFFF)
  InactiveTab:   foreground(#626262)
  TableHeader:   bold, foreground(#7D56F4)
  TableRow:      foreground(#DDDDDD)
  TableSelected: background(#3D3D3D), foreground(#FFFFFF)
  Positive:      foreground(#04B575)      # green
  Neutral:       foreground(#FBBF24)      # yellow
  Negative:      foreground(#FF4672)      # red
  Error:         foreground(#FF4672), bold
  Muted:         foreground(#626262)
  Border:        foreground(#3D3D3D)
  Preview:       border(rounded), padding(1)
```

---

## `internal/domain/` — Data Types

Plain structs. No bubbletea dependency. Used by screens, db, and components.

### `domain/person.go`

```yaml
struct: Person
fields:
  ID: string
  Name: string
  Description: string
  FirstMentioned: string
  RelationCount: int             # computed
  BehaviorCount: int             # computed

struct: PersonDetail
fields:
  Person: Person
  Relationships: []RelationshipSummary
  Behaviors: []BehaviorSummary
```

### `domain/relationship.go`

```yaml
struct: Relationship
fields:
  ID: string
  FromPerson: string
  ToPerson: string
  FromName: string               # joined
  ToName: string                 # joined
  RelationshipType: string
  Description: string
  Sentiment: string              # positive|neutral|negative
  Strength: float64
  Timestamp: string

struct: RelationshipSummary
fields:
  ID: string
  FromName: string
  ToName: string
  Type: string
  Strength: float64

# For the evolution screen
struct: Snapshot
fields:
  Timestamp: string
  RelationshipType: string
  Sentiment: string
  Strength: float64
  Description: string

struct: EvolutionData
fields:
  FromName: string
  ToName: string
  Snapshots: []Snapshot
  TypeChanged: bool
  SentimentChanged: bool
  StrengthDelta: float64          # last - first
```

### `domain/behavior.go`

```yaml
struct: Behavior
fields:
  ID: string
  PersonID: string
  PersonName: string             # joined
  Timestamp: string
  BehaviorType: string
  Description: string
  Intensity: string

struct: BehaviorSummary
fields:
  Type: string
  Count: int
```

### `domain/event.go`

```yaml
struct: Event
fields:
  ID: string
  Timestamp: string
  Description: string
  PersonIDs: []string

struct: TimelineEntry
fields:
  Timestamp: string
  EntryType: string              # EVENT|REL|BEH
  Summary: string
  Delta: string                  # e.g. "▲ +0.07" or "(new)" or ""
  SourceID: string               # for navigation
```

### `domain/extraction.go`

```yaml
struct: ExtractionResult
fields:
  Persons: []Person
  Relationships: []Relationship
  Behaviors: []Behavior
  Events: []Event

struct: ImportPreview
fields:
  PersonCount: int
  RelationshipCount: int
  BehaviorCount: int
  EventCount: int
  Conflicts: []Conflict

struct: Conflict
fields:
  EntityType: string
  ID: string
  Reason: string                 # "already exists", "key mismatch"

struct: PluginDescriptor
fields:
  APIVersion: string             # "cozo.extractor/v1"
  Kind: string                   # "extractor"
  ID: string
  Name: string
  FilePath: string
```

---

## `internal/db/` — CozoDB Access

### `db/cozo.go`

Connection lifecycle and query execution.

```yaml
struct: DB
fields:
  handle: *cozo.DB               # native cozo handle
  path: string

methods:
  Open(path string) -> (*DB, error)
  Close() -> error
  Query(script string, params map[string]any) -> (*Result, error)
  QueryCmd(script string, params map[string]any) -> tea.Cmd
    # wraps Query in a tea.Cmd that returns QueryResultMsg or QueryErrorMsg

struct: Result
fields:
  Columns: []string
  Rows: [][]any
  RowCount: int
  TimeMs: int64

msgs:
  QueryResultMsg:
    Tag: string                  # caller-defined tag to route results
    Result: *Result
  QueryErrorMsg:
    Tag: string
    Err: error
```

### `db/queries.go`

Named CozoScript query templates. All queries in one place for easy review.

```yaml
constants:
  # --- Dashboard ---
  QueryEntityCounts: |
    persons[count(id)] := *person{id}
    rels[count(id)] := *relationship{id}
    behs[count(id)] := *behavior{id}
    evts[count(id)] := *event{id}
    ?[entity, cnt] :=
      persons[cnt], entity = 'person'
    ?[entity, cnt] :=
      rels[cnt], entity = 'relationship'
    ?[entity, cnt] :=
      behs[cnt], entity = 'behavior'
    ?[entity, cnt] :=
      evts[cnt], entity = 'event'

  QueryTopRelationships: |
    ?[from_name, to_name, type, strength] :=
      *relationship{from_person, to_person, relationship_type: type, strength},
      *person{id: from_person, name: from_name},
      *person{id: to_person, name: to_name}
    :order -strength
    :limit $limit

  QuerySentimentDistribution: |
    ?[sentiment, count(id)] := *relationship{id, sentiment}

  # --- People ---
  QueryAllPersons: |
    rel_count[pid, count(rid)] :=
      *relationship{id: rid, from_person: pid}
    rel_count[pid, count(rid)] :=
      *relationship{id: rid, to_person: pid}
    beh_count[pid, count(bid)] :=
      *behavior{id: bid, person_id: pid}
    ?[id, name, first_mentioned, description, rcnt, bcnt] :=
      *person{id, name, first_mentioned, description},
      rel_count[id, rcnt],
      beh_count[id, bcnt]

  QueryPersonDetail: |
    rels[type, cnt] :=
      *relationship{from_person: $pid, relationship_type: type},
      cnt = count(type)
    rels[type, cnt] :=
      *relationship{to_person: $pid, relationship_type: type},
      cnt = count(type)
    ?[type, cnt] := rels[type, cnt]

  # --- Relationships ---
  QueryAllRelationships: |
    ?[id, from_name, to_name, relationship_type, sentiment,
      strength, timestamp] :=
      *relationship{id, from_person, to_person, relationship_type,
                    sentiment, strength, timestamp},
      *person{id: from_person, name: from_name},
      *person{id: to_person, name: to_name}
    :order from_name, to_name, timestamp

  QueryRelationshipsByType: |
    ?[id, from_name, to_name, relationship_type, sentiment,
      strength, timestamp] :=
      *relationship{id, from_person, to_person,
                    relationship_type: $type,
                    sentiment, strength, timestamp},
      *person{id: from_person, name: from_name},
      *person{id: to_person, name: to_name}
    :order from_name, to_name, timestamp

  QueryRelationshipsByPerson: |
    ?[id, from_name, to_name, relationship_type, sentiment,
      strength, timestamp] :=
      *relationship{id, from_person, to_person, relationship_type,
                    sentiment, strength, timestamp},
      *person{id: from_person, name: from_name},
      *person{id: to_person, name: to_name},
      or(from_person == $pid, to_person == $pid)
    :order timestamp

  # --- Evolution ---
  QuerySnapshots: |
    ?[timestamp, relationship_type, sentiment, strength, description] :=
      *relationship{from_person: $from, to_person: $to,
                    timestamp, relationship_type, sentiment,
                    strength, description}
    :order timestamp

  # --- Timeline ---
  QueryTimeline: |
    events[timestamp, 'EVENT', description, '', id] :=
      *event{id, timestamp, description}
    rels[timestamp, 'REL', summary, delta, id] :=
      *relationship{id, from_person, to_person, timestamp,
                    relationship_type, strength},
      *person{id: from_person, name: fn},
      *person{id: to_person, name: tn},
      summary = concat(fn, ' → ', tn, '  ', relationship_type),
      delta = format('%.2f', strength)
    behs[timestamp, 'BEH', summary, '', id] :=
      *behavior{id, person_id, timestamp, behavior_type},
      *person{id: person_id, name},
      summary = concat(name, ': ', behavior_type)
    ?[timestamp, type, summary, delta, source_id] :=
      events[timestamp, type, summary, delta, source_id]
    ?[timestamp, type, summary, delta, source_id] :=
      rels[timestamp, type, summary, delta, source_id]
    ?[timestamp, type, summary, delta, source_id] :=
      behs[timestamp, type, summary, delta, source_id]
    :order -timestamp

  # --- Network ---
  QueryGraphEdges: |
    ?[from_id, to_id, from_name, to_name, strength, sentiment] :=
      *relationship{from_person: from_id, to_person: to_id,
                    strength, sentiment},
      *person{id: from_id, name: from_name},
      *person{id: to_id, name: to_name}

  QueryCommunities: |
    edges[f, t] := *relationship{from_person: f, to_person: t}
    ?[node, community] <~ CommunityDetectionLouvain(edges[from, to])

  # --- Vector Search ---
  QueryVectorPerson: |
    ?[name, score, description] :=
      ~person:person_embedding_idx{id | query: $vec, k: $k,
                                   ef: 200, score},
      *person{id, name, description}
    :order -score

  QueryVectorRelationship: |
    ?[description, score, sentiment, strength] :=
      ~relationship:rel_embedding_idx{id | query: $vec, k: $k,
                                      ef: 200, score},
      *relationship{id, description, sentiment, strength}
    :order -score

  QueryVectorBehavior: |
    ?[person_name, behavior_type, score] :=
      ~behavior:beh_embedding_idx{id | query: $vec, k: $k,
                                  ef: 200, score},
      *behavior{id, person_id, behavior_type},
      *person{id: person_id, name: person_name}
    :order -score

  QueryVectorEvent: |
    ?[description, score, timestamp] :=
      ~event:event_embedding_idx{id | query: $vec, k: $k,
                                 ef: 200, score},
      *event{id, description, timestamp}
    :order -score

  # --- Import ---
  MutationPutPerson: |
    ?[id, name, description, first_mentioned, embedding]
      <- [[$id, $name, $description, $first_mentioned, $embedding]]
    :put person {id => name, description, first_mentioned, embedding}

  MutationPutRelationship: |
    ?[id, from_person, to_person, timestamp, relationship_type,
      description, sentiment, strength, embedding]
      <- [[$id, $from_person, $to_person, $timestamp,
           $relationship_type, $description, $sentiment,
           $strength, $embedding]]
    :put relationship {id, from_person, to_person, timestamp =>
      relationship_type, description, sentiment, strength, embedding}

  MutationPutBehavior: |
    ?[id, person_id, timestamp, behavior_type, description,
      intensity, embedding]
      <- [[$id, $person_id, $timestamp, $behavior_type,
           $description, $intensity, $embedding]]
    :put behavior {id, person_id, timestamp =>
      behavior_type, description, intensity, embedding}

  MutationPutEvent: |
    ?[id, timestamp, description, person_ids, embedding]
      <- [[$id, $timestamp, $description, $person_ids, $embedding]]
    :put event {id => timestamp, description, person_ids, embedding}
```

### `db/params.go`

Helper to build CozoDB parameter maps from domain structs.

```yaml
functions:
  PersonParams(p domain.Person) -> map[string]any
  RelationshipParams(r domain.Relationship) -> map[string]any
  BehaviorParams(b domain.Behavior) -> map[string]any
  EventParams(e domain.Event) -> map[string]any
  VectorSearchParams(vec []float32, k int) -> map[string]any
```

---

## `internal/embedding/` — Vector Embedding

### `embedding/geppetto.go`

Wraps Geppetto's embedding API to produce 384-dim F32 vectors from text.

```yaml
struct: Embedder
fields:
  engine: geppetto.EmbeddingEngine   # configured for text-embedding-3-small

methods:
  New(configPath string) -> (*Embedder, error)
  Embed(ctx context.Context, text string) -> ([]float32, error)
  EmbedCmd(text string) -> tea.Cmd
    # returns EmbeddingResultMsg or EmbeddingErrorMsg

msgs:
  EmbeddingResultMsg:
    Text: string
    Vector: []float32
  EmbeddingErrorMsg:
    Err: error
```

---

## `internal/screens/` — Screen Models

### `screens/screen.go`

Interface that all screens implement. This is how `app/model.go` talks to screens generically.

```yaml
interface: Screen
methods:
  Init() -> tea.Cmd
  Update(msg tea.Msg) -> (Screen, tea.Cmd)
  View() -> string
  SetSize(width, height int)     # called on resize
  ShortHelp() -> []key.Binding   # for status bar hints
  FullHelp() -> [][]key.Binding  # for help overlay
```

### `screens/dashboard/model.go`

```yaml
struct: Model
fields:
  db: *db.DB
  width: int
  height: int
  counts: map[string]int         # person->12, relationship->34, ...
  topRels: []domain.RelationshipSummary
  recentEvents: []domain.Event
  sentiment: map[string]int      # positive->65, neutral->25, negative->10
  avgStrength: float64
  focusedPanel: int              # 0=counts, 1=topRels, 2=events, 3=sentiment
  cursor: int                    # within focused list panel
  loading: bool
  spinner: spinner.Model

lifecycle:
  Init: |
    tea.Batch(
      fetchCounts(db),
      fetchTopRelationships(db, 10),
      fetchRecentEvents(db, 8),
      fetchSentiment(db),
    )
  Update: |
    handle QueryResultMsg by tag (counts|topRels|events|sentiment)
    handle KeyMsg: tab=cycle panel, j/k=cursor, enter=navigate, r=refresh
```

### `screens/dashboard/commands.go`

```yaml
functions:
  fetchCounts(db) -> tea.Cmd          # tag: "counts"
  fetchTopRelationships(db, limit) -> tea.Cmd  # tag: "topRels"
  fetchRecentEvents(db, limit) -> tea.Cmd      # tag: "events"
  fetchSentiment(db) -> tea.Cmd       # tag: "sentiment"
```

### `screens/dashboard/view.go`

```yaml
functions:
  View(m Model) -> string
    # composes 4 panels using lipgloss.JoinHorizontal/Vertical
    # renders bar charts for strength/sentiment using block chars
```

---

### `screens/people/model.go`

```yaml
struct: Model
fields:
  db: *db.DB
  width: int
  height: int
  table: datatable.Model         # reusable component
  preview: preview.Model         # bottom pane
  search: textinput.Model        # bubbles/textinput
  searchActive: bool
  sortField: string              # name|first_mentioned|relation_count
  sortAsc: bool
  filter: string                 # all|has_relationships|isolated
  persons: []domain.Person       # full dataset
  filtered: []domain.Person      # after search/filter
  previewDebounce: time.Duration

lifecycle:
  Init: fetchPersons(db)
  Update: |
    handle QueryResultMsg tag "persons" -> populate table
    handle QueryResultMsg tag "preview" -> populate preview pane
    handle KeyMsg: j/k, /, s=sort, f=filter, enter=navigate(person_detail)
    on cursor change -> debounced fetchPreview
```

### `screens/people/commands.go`

```yaml
functions:
  fetchPersons(db, sort, filter) -> tea.Cmd       # tag: "persons"
  fetchPreview(db, personID) -> tea.Cmd            # tag: "preview"
```

---

### `screens/relationships/model.go`

```yaml
struct: Model
fields:
  db: *db.DB
  width: int
  height: int
  table: datatable.Model
  detail: preview.Model          # bottom detail pane
  typeFilter: string             # "all" or specific type
  personFilter: string           # "" or person ID
  relationships: []domain.Relationship

lifecycle:
  Init: fetchRelationships(db, "all", "")
  Update: |
    handle data load -> populate table
    handle KeyMsg: j/k, t=type filter, p=person filter,
                   enter=navigate(evolution, {from, to})
```

---

### `screens/evolution/model.go`

```yaml
struct: Model
fields:
  db: *db.DB
  width: int
  height: int
  fromPerson: string             # param from navigation
  toPerson: string               # param from navigation
  fromName: string
  toName: string
  snapshots: []domain.Snapshot
  chart: Chart                   # ASCII strength-over-time
  table: datatable.Model         # snapshot list
  summary: domain.EvolutionData  # computed
  diffMode: bool
  diffA: int                     # selected snapshot index
  diffB: int

lifecycle:
  Init: fetchSnapshots(db, fromPerson, toPerson)
  Update: |
    handle data load -> populate chart + table + compute summary
    handle KeyMsg: j/k, d=toggle diff, escape=back
```

### `screens/evolution/chart.go`

```yaml
struct: Chart
fields:
  points: []ChartPoint           # {x: timestamp, y: strength}
  width: int
  height: int                    # typically 8 rows

methods:
  Render() -> string
    # ASCII line chart using braille/block characters
    # X axis: timestamps, Y axis: 0.0-1.0
```

---

### `screens/network/model.go`

```yaml
struct: Model
fields:
  db: *db.DB
  width: int
  height: int
  nodes: []GraphNode
  edges: []GraphEdge
  positions: map[string]Point    # computed by layout
  layoutType: string             # force|circular|hierarchical
  depth: int                     # hop limit from focus
  focusNode: string              # "" = show all
  communities: map[string]int    # node -> community ID
  communitiesActive: bool
  viewport: viewport.Model       # bubbles/viewport for panning
  stats: GraphStats

lifecycle:
  Init: fetchGraphData(db)
  Update: |
    handle GraphDataLoaded -> run computeLayout
    handle LayoutComputed -> set positions
    handle KeyMsg: hjkl=pan, enter=focus, +/-=depth, c=communities
```

### `screens/network/layout.go`

```yaml
struct: GraphNode
fields:
  ID: string
  Name: string

struct: GraphEdge
fields:
  From: string
  To: string
  Strength: float64
  Sentiment: string

struct: Point
fields:
  X: int
  Y: int

functions:
  ForceDirectedLayout(nodes, edges, width, height, iterations) -> map[string]Point
  CircularLayout(nodes, width, height) -> map[string]Point
  HierarchicalLayout(nodes, edges, width, height) -> map[string]Point
```

### `screens/network/render.go`

```yaml
functions:
  RenderGraph(nodes, edges, positions, width, height, focus, communities) -> string
    # draws nodes as [Name] boxes
    # draws edges as lines: ── positive, ╌╌ neutral, ░░ negative
    # labels edges with strength
    # colors nodes by community if active
```

---

### `screens/timeline/model.go`

```yaml
struct: Model
fields:
  db: *db.DB
  width: int
  height: int
  entries: []domain.TimelineEntry
  groups: []TimelineGroup        # grouped by year-month
  list: list.Model               # bubbles/list
  typeFilter: string             # all|event|rel|behavior
  personFilter: string

struct: TimelineGroup
fields:
  YearMonth: string              # "2024-02"
  Entries: []domain.TimelineEntry

lifecycle:
  Init: fetchTimeline(db, "all", "")
  Update: |
    handle data load -> group by year-month, populate list
    handle KeyMsg: j/k, f=type filter, p=person filter,
                   enter=navigate to detail based on EntryType
```

---

### `screens/query/model.go`

```yaml
struct: Model
fields:
  db: *db.DB
  width: int
  height: int
  editor: textarea.Model         # bubbles/textarea
  results: datatable.Model       # dynamic columns
  resultColumns: []string
  resultRows: [][]any
  executionTimeMs: int64
  rowCount: int
  queryError: string
  errorLine: int
  errorCol: int
  history: History               # see history.go
  historyOverlay: bool
  savedQueries: map[string]string
  paneSplit: float64             # 0.0-1.0, fraction for editor
  focus: int                     # 0=editor, 1=results

lifecycle:
  Init: load history from disk, set editor placeholder
  Update: |
    handle QueryResultMsg tag "query" -> populate results table
    handle QueryErrorMsg tag "query" -> show error
    handle KeyMsg: ctrl+e=run, ctrl+h=history, ctrl+s=save,
                   ctrl+up/down=resize, tab=switch focus
```

### `screens/query/history.go`

```yaml
struct: History
fields:
  entries: []HistoryEntry
  index: int                     # current position (-1 = new)
  maxEntries: int                # 50
  filePath: string               # ~/.cozo-tui/history.json

struct: HistoryEntry
fields:
  Query: string
  Timestamp: time.Time
  RowCount: int
  TimeMs: int64

methods:
  Add(entry HistoryEntry)
  Previous() -> *HistoryEntry
  Next() -> *HistoryEntry
  Load() -> error
  Save() -> error
```

---

### `screens/extraction/model.go`

```yaml
struct: Model
fields:
  db: *db.DB
  scriptsDir: string
  width: int
  height: int
  sourceFile: string
  pluginID: string
  pluginName: string
  status: string                 # idle|running|completed|error
  elapsedMs: int64
  err: string
  results: *domain.ExtractionResult
  preview: *domain.ImportPreview
  plugins: []domain.PluginDescriptor
  pluginBrowser: list.Model      # overlay
  pluginBrowserActive: bool
  spinner: spinner.Model
  focus: int                     # 0=status, 1=results, 2=import

lifecycle:
  Init: discoverPlugins(scriptsDir)
  Update: |
    handle PluginsLoaded -> populate plugin list
    handle ExtractionCompleted -> show results, compute preview
    handle ImportCompleted -> show success toast
    handle KeyMsg: n=new, r=rerun, i=import, e=export, p=plugins
```

### `screens/extraction/plugins.go`

```yaml
functions:
  discoverPlugins(scriptsDir string) -> tea.Cmd
    # walks scriptsDir, loads each .js file header
    # validates apiVersion/kind/id/name fields
    # returns PluginsLoaded msg

  runExtraction(scriptsDir, pluginID, sourceFile string) -> tea.Cmd
    # creates Goja runtime (same as cozo-relationship-js-runner)
    # loads plugin, calls create(), then run({transcript: text})
    # returns ExtractionCompleted or ExtractionError

  importResults(db *db.DB, results *domain.ExtractionResult) -> tea.Cmd
    # runs :put queries for each entity type
    # returns ImportCompleted with counts
```

---

### `screens/vector/model.go`

```yaml
struct: Model
fields:
  db: *db.DB
  embedder: *embedding.Embedder
  width: int
  height: int
  queryInput: textinput.Model
  index: string                  # person|relationship|behavior|event
  k: int                        # default 10
  results: datatable.Model       # dynamic columns per index
  searchTimeMs: int64
  searching: bool
  spinner: spinner.Model

lifecycle:
  Init: focus queryInput
  Update: |
    handle EmbeddingResultMsg -> run HNSW query
    handle QueryResultMsg tag "vector" -> populate results
    handle KeyMsg: enter=search, tab=cycle index, K=set k,
                   j/k=browse results
```

---

## `internal/components/` — Reusable Widgets

### `components/datatable/model.go`

A generic sortable, scrollable table with configurable columns.

```yaml
struct: Column
fields:
  Key: string
  Label: string
  Width: int                     # 0 = flex
  Format: string                 # go format string, e.g. "%.2f"
  Style: lipgloss.Style          # optional per-column styling
  Truncate: bool

struct: Model
fields:
  columns: []Column
  rows: []Row                    # Row = map[string]any
  cursor: int
  offset: int                    # scroll offset
  width: int
  height: int
  sortCol: int
  sortAsc: bool

methods:
  SetRows(rows []Row)
  SelectedRow() -> Row
  SetSize(w, h int)
  Init/Update/View                # standard bubbletea
```

### `components/preview/model.go`

A bordered text box for showing entity detail below a table.

```yaml
struct: Model
fields:
  title: string
  content: string                # rendered text
  width: int
  height: int
  viewport: viewport.Model       # scrollable if content is long
  loading: bool
  spinner: spinner.Model

methods:
  SetContent(title, content string)
  SetSize(w, h int)
```

### `components/statusbar/model.go`

```yaml
struct: Model
fields:
  tabs: []Tab
  activeTab: int
  dbPath: string
  width: int

struct: Tab
fields:
  Key: string                    # "F1"
  Label: string                  # "Dashboard"
  Screen: ScreenID
```

### `components/help/model.go`

```yaml
struct: Model
fields:
  visible: bool
  width: int
  height: int
  globalBindings: []Binding
  screenBindings: []Binding      # set by active screen

struct: Binding
fields:
  Key: string
  Description: string
```

### `components/filterprompt/model.go`

A small overlay prompt for entering filter text (used by person filter, search).

```yaml
struct: Model
fields:
  active: bool
  title: string                  # "Filter by person:"
  input: textinput.Model
  suggestions: []string          # autocomplete from person names
  selected: string

methods:
  Open(title string, suggestions []string)
  Close()
  Value() -> string
```

---

## Dependency Graph

```
main.go
  └── internal/app
        ├── internal/screens/*       (each screen package)
        │     ├── internal/domain    (data structs)
        │     ├── internal/db        (queries)
        │     ├── internal/embedding (vector search only)
        │     └── internal/components/* (shared widgets)
        ├── internal/components/statusbar
        ├── internal/components/help
        └── internal/db
```

No circular dependencies. Screens depend on domain, db, components. App depends on screens. Domain depends on nothing.

---

## Build Order

Suggested implementation order, building up from foundations:

```
Phase 1 — Foundation
  1. domain/          (plain structs, zero deps)
  2. db/cozo.go       (connection wrapper)
  3. db/queries.go    (all CozoScript templates)
  4. db/params.go     (struct -> param map)

Phase 2 — Components
  5. components/datatable/
  6. components/preview/
  7. components/statusbar/
  8. components/help/
  9. components/filterprompt/

Phase 3 — App Shell
  10. app/styles.go
  11. app/keys.go
  12. app/navigation.go
  13. app/model.go
  14. main.go          (minimal, just dashboard)

Phase 4 — Screens (simplest first)
  15. screens/dashboard/
  16. screens/people/
  17. screens/relationships/
  18. screens/timeline/
  19. screens/evolution/

Phase 5 — Advanced Screens
  20. screens/query/
  21. screens/vector/ + embedding/geppetto.go
  22. screens/extraction/
  23. screens/network/    (hardest — ASCII graph layout)
```
