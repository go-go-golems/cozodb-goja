---
Title: TUI Screen Designs and Bubbletea Models
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
      Note: Go CLI runner that connects to CozoDB - TUI will share this connection logic
    - Path: cozo_demo.py
      Note: Schema definitions and query examples to replicate in TUI
    - Path: cozo_advanced_demo.py
      Note: Advanced queries (PageRank, community detection) to expose in TUI
ExternalSources:
    - https://github.com/charmbracelet/bubbletea
    - https://github.com/charmbracelet/lipgloss
    - https://github.com/charmbracelet/bubbles
Summary: "ASCII mockups and Bubbletea model sketches for 9 TUI screens"
LastUpdated: 2026-02-24T20:00:00-05:00
WhatFor: "Design reference for building a terminal UI to explore the CozoDB relationship database"
WhenToUse: "Before implementing any TUI screen, consult this doc for layout, model structure, and navigation"
---

# TUI Screen Designs and Bubbletea Models

## Design Principles

- **Data-driven**: Every screen maps directly to CozoDB queries
- **Keyboard-first**: vim-style navigation (j/k, g/G, /search, :command)
- **Composable**: Screens share common components (status bar, help overlay, table widget)
- **Lazy loading**: Data fetched on demand, with loading spinners

## Navigation Map

```
┌──────────┐    tab     ┌──────────┐    tab     ┌──────────┐
│Dashboard │ ────────── │ People   │ ────────── │Relations │
│  (F1)    │            │  (F2)    │            │  (F3)    │
└────┬─────┘            └────┬─────┘            └────┬─────┘
     │                       │ enter                  │ enter
     │                  ┌────┴─────┐            ┌────┴─────┐
     │                  │ Person   │            │ Rel.     │
     │                  │ Detail   │            │ Detail   │
     │                  └──────────┘            └──────────┘
     │
     │  tab     ┌──────────┐    tab     ┌──────────┐
     ├───────── │ Timeline │ ────────── │ Network  │
     │          │  (F4)    │            │  (F5)    │
     │          └──────────┘            └──────────┘
     │
     │  tab     ┌──────────┐    tab     ┌──────────┐
     ├───────── │ Query    │ ────────── │ Extract  │
     │          │  (F6)    │            │  (F7)    │
     │          └──────────┘            └──────────┘
     │
     │  tab     ┌──────────┐
     └───────── │ Vector   │
                │  (F8)    │
                └──────────┘
```

Global: `F1`-`F8` jump to any screen. `?` opens help overlay. `q` quits.

---

## Screen 1: Dashboard

### Mockup

```
┌─ CozoDB Relationship Explorer ─────────────────────────────────────────────┐
│                                                                            │
│  ┌─ Entities ─────────┐  ┌─ Top Relationships ────────────────────────┐   │
│  │ Persons       12    │  │ Sarah Chen → James Liu   mentor   ██████ │   │
│  │ Relationships 34    │  │ Sarah Chen → Maya Patel  collab   █████  │   │
│  │ Behaviors     28    │  │ James Liu → Alex Kim     mentor   ████   │   │
│  │ Events        15    │  │ Maya Patel → Tom Rivera  collab   ████   │   │
│  │                     │  │ Sarah Chen → Alex Kim    friend   ███    │   │
│  │ Last extract:       │  │                                          │   │
│  │   2024-02-15 14:30  │  │                              by strength │   │
│  └─────────────────────┘  └──────────────────────────────────────────┘   │
│                                                                            │
│  ┌─ Recent Events ──────────────────┐  ┌─ Sentiment Distribution ──────┐  │
│  │ 2024-02  Grant awarded to lab    │  │ Positive  ████████████ 65%    │  │
│  │ 2024-01  New team member joined  │  │ Neutral   ██████       25%    │  │
│  │ 2023-11  Paper published         │  │ Negative  ███          10%    │  │
│  │ 2023-09  Conference presentation │  │                               │  │
│  │ 2023-06  Lab reorganization      │  │ Avg strength: 0.72           │  │
│  └──────────────────────────────────┘  └───────────────────────────────┘  │
│                                                                            │
│  [F1]Dashboard [F2]People [F3]Relations [F4]Timeline [F5]Network          │
│  [F6]Query [F7]Extract [F8]Vector                        db:cozo.db ?help │
└────────────────────────────────────────────────────────────────────────────┘
```

### How It Works

The dashboard is the landing screen. It runs 4 CozoDB aggregate queries on startup to populate the entity counts, top relationships (sorted by strength descending), recent events (sorted by timestamp descending), and sentiment distribution. Each panel is an independent component that can show a loading spinner while its query runs. Panels auto-refresh when extraction completes.

The top-relationships panel is interactive: pressing `enter` on a highlighted row navigates to the Relationship Detail view. The recent-events panel works the same way. The entity counts panel is display-only.

### Model

```yaml
screen: dashboard
model:
  counts:
    type: EntityCounts
    fields: [persons, relationships, behaviors, events, last_extract_time]
    query: |
      persons[count(id)] := *person{id}
      relationships[count(id)] := *relationship{id}
      # ... similar for behaviors, events
  top_relationships:
    type: list.Model          # bubbles/list
    items: RelationshipRow[]
    fields: [from_name, to_name, type, strength]
    query: |
      ?[from_name, to_name, type, strength] :=
        *relationship{from_person, to_person, relationship_type: type, strength},
        *person{id: from_person, name: from_name},
        *person{id: to_person, name: to_name}
      :order -strength
      :limit 10
    cursor: true
    on_enter: navigate(relationship_detail, selected.id)
  recent_events:
    type: list.Model
    items: EventRow[]
    fields: [timestamp, description]
    query: |
      ?[timestamp, description] := *event{timestamp, description}
      :order -timestamp
      :limit 8
    cursor: true
    on_enter: navigate(timeline, filter=selected.timestamp)
  sentiment:
    type: SentimentChart
    fields: [positive_pct, neutral_pct, negative_pct, avg_strength]
    query: |
      ?[sentiment, count(id)] := *relationship{id, sentiment}

msgs:
  - DataLoaded { panel: string, data: any }
  - PanelFocusChanged { panel: string }
  - NavigateTo { screen: string, params: map }

cmds:
  - fetchCounts -> DataLoaded
  - fetchTopRelationships -> DataLoaded
  - fetchRecentEvents -> DataLoaded
  - fetchSentiment -> DataLoaded

keybindings:
  tab: cycle_panel_focus
  enter: activate_selected_item
  r: refresh_all_panels
```

---

## Screen 2: People Browser

### Mockup

```
┌─ People ──────────────────────────────────────────────────────── /search ──┐
│                                                                            │
│  Filter: [all▾]  Sort: [name▾]         Showing 12 of 12                   │
│                                                                            │
│  Name              First Seen   Relations   Behaviors   Description        │
│  ─────────────────────────────────────────────────────────────────────────  │
│▸ Sarah Chen        2023-01      8           6           Senior researcher  │
│  James Liu         2023-01      6           4           Lab director and   │
│  Maya Patel        2023-03      5           5           Graduate student   │
│  Alex Kim          2023-06      4           3           Postdoctoral res   │
│  Tom Rivera        2023-01      3           2           Lab technician a   │
│  Lisa Wang         2023-09      2           3           Visiting researc   │
│                                                                            │
│                                                                            │
│                                                                            │
│  ┌─ Preview: Sarah Chen ──────────────────────────────────────────────┐   │
│  │ Senior researcher and team lead. Mentors multiple junior members.  │   │
│  │ Relationships: James Liu (mentor), Maya Patel (collaborator),      │   │
│  │   Alex Kim (mentor), Tom Rivera (supervisor)                       │   │
│  │ Key behaviors: mentoring, publishing, grant writing                │   │
│  └────────────────────────────────────────────────────────────────────┘   │
│                                                                            │
│  j/k:move  enter:detail  /:search  s:sort  f:filter          ?help       │
└────────────────────────────────────────────────────────────────────────────┘
```

### How It Works

A filterable, sortable table of all persons in the database. The table is the primary component — standard j/k to move, `/` to open a search filter that narrows rows by fuzzy-matching against name and description. The preview pane at the bottom shows a summary for the currently highlighted person, including their relationship count and most recent behaviors. This preview is fetched lazily as the cursor moves (debounced 150ms).

Sort cycles through: name, first_mentioned, relationship_count. Filter dropdown offers: all, has_relationships, isolated (no relationships), by_sentiment (only those involved in negative relationships).

Pressing `enter` opens the Person Detail view (a sub-screen that overlays or replaces the browser).

### Model

```yaml
screen: people_browser
model:
  table:
    type: table.Model          # bubbles/table
    columns:
      - name: name
        width: 18
      - name: first_mentioned
        width: 12
        label: "First Seen"
      - name: relation_count
        width: 11
        label: "Relations"
      - name: behavior_count
        width: 11
        label: "Behaviors"
      - name: description
        width: flex
        truncate: true
    query: |
      rel_count[person_id, count(id)] :=
        *relationship{id, from_person: person_id}
      rel_count[person_id, count(id)] :=
        *relationship{id, to_person: person_id}
      ?[id, name, first_mentioned, description, rel_count, beh_count] :=
        *person{id, name, first_mentioned, description},
        rel_count[id, rel_count],
        beh_count[id, beh_count]
    cursor: true
  search:
    type: textinput.Model      # bubbles/textinput
    placeholder: "search people..."
    active: false
  sort_field: "name"
  sort_dir: "asc"
  filter: "all"
  preview:
    type: PersonPreview
    person_id: null             # set from cursor
    relationships: []
    behaviors: []
    debounce_ms: 150

msgs:
  - PeopleLoaded { rows: PersonRow[] }
  - PreviewLoaded { person_id: string, data: PreviewData }
  - SearchChanged { query: string }
  - SortChanged { field: string }
  - FilterChanged { filter: string }

cmds:
  - fetchPeople(sort, filter, search) -> PeopleLoaded
  - fetchPreview(person_id) -> PreviewLoaded

keybindings:
  j: cursor_down
  k: cursor_up
  enter: navigate(person_detail, selected.id)
  /: activate_search
  escape: deactivate_search
  s: cycle_sort
  f: cycle_filter
  g: cursor_top
  G: cursor_bottom
```

---

## Screen 3: Relationship Explorer

### Mockup

```
┌─ Relationships ────────────────────────────────────────── type:[all▾] ─────┐
│                                                                            │
│  From            To              Type         Sent   Str   When            │
│  ─────────────────────────────────────────────────────────────────────────  │
│▸ Sarah Chen      James Liu       mentor       pos    0.9   2023-01        │
│  Sarah Chen      James Liu       mentor       pos    0.85  2023-06        │
│  Sarah Chen      James Liu       mentor       pos    0.92  2024-01        │
│  Sarah Chen      Maya Patel      collaborator pos    0.8   2023-03        │
│  Sarah Chen      Alex Kim        mentor       pos    0.75  2023-06        │
│  James Liu       Maya Patel      supervisor   neu    0.7   2023-03        │
│  Maya Patel      Tom Rivera      collaborator pos    0.6   2023-09        │
│  Alex Kim        Tom Rivera      peer         pos    0.5   2023-09        │
│                                                                            │
│  ┌─ Detail ───────────────────────────────────────────────────────────┐   │
│  │ Sarah Chen → James Liu (mentor)                                    │   │
│  │ "Chen and Liu maintained a strong mentoring relationship           │   │
│  │  throughout the research period, with Chen providing guidance      │   │
│  │  on methodology and Liu offering career mentorship."               │   │
│  │ Sentiment: positive  Strength: 0.90  Since: 2023-01              │   │
│  │ Snapshots: 3 (2023-01, 2023-06, 2024-01)                         │   │
│  └────────────────────────────────────────────────────────────────────┘   │
│                                                                            │
│  j/k:move  enter:evolution  t:type  p:person  /:search            ?help   │
└────────────────────────────────────────────────────────────────────────────┘
```

### How It Works

This is the main relationship browsing screen. It shows all relationship records, including multiple temporal snapshots of the same logical relationship (note the 3 rows for Sarah Chen → James Liu at different timestamps). The composite key `{id, from_person, to_person, timestamp}` means each row is a point-in-time snapshot.

The type filter dropdown at the top right lets you narrow to: mentor, collaborator, supervisor, peer, friend, rival, etc. The person filter (`p`) opens a fuzzy-match prompt to show only relationships involving a specific person.

Sentiment is color-coded in the terminal: green for positive, yellow for neutral, red for negative. Strength shows as a numeric value and also as a visual bar in the detail pane.

Pressing `enter` on a relationship navigates to the Relationship Evolution screen, pre-filtered to show all temporal snapshots of that specific relationship pair.

### Model

```yaml
screen: relationship_explorer
model:
  table:
    type: table.Model
    columns:
      - { name: from_name, width: 16 }
      - { name: to_name, width: 16 }
      - { name: relationship_type, width: 12, label: "Type" }
      - { name: sentiment, width: 6, label: "Sent", style: sentiment_color }
      - { name: strength, width: 5, label: "Str", format: "%.1f" }
      - { name: timestamp, width: 10, label: "When" }
    query: |
      ?[id, from_name, to_name, relationship_type, sentiment, strength, timestamp] :=
        *relationship{id, from_person, to_person, relationship_type,
                      sentiment, strength, timestamp},
        *person{id: from_person, name: from_name},
        *person{id: to_person, name: to_name}
      :order from_name, to_name, timestamp
    cursor: true
  type_filter: "all"
  person_filter: null
  detail:
    type: RelationshipDetail
    fields: [description, sentiment, strength, snapshot_count, snapshot_dates]
    query_on_cursor: true

msgs:
  - RelationshipsLoaded { rows: RelRow[] }
  - DetailLoaded { rel_id: string, data: DetailData }
  - TypeFilterChanged { type: string }
  - PersonFilterChanged { person: string }

cmds:
  - fetchRelationships(type_filter, person_filter) -> RelationshipsLoaded
  - fetchDetail(rel_id) -> DetailLoaded

keybindings:
  j: cursor_down
  k: cursor_up
  enter: navigate(relationship_evolution, { from: selected.from_person, to: selected.to_person })
  t: open_type_filter
  p: open_person_filter
  /: search
```

---

## Screen 4: Relationship Evolution

### Mockup

```
┌─ Evolution: Sarah Chen → James Liu ───────────────────────────────────────┐
│                                                                            │
│  Strength over time                                                        │
│  1.0 ┤                                                          ●         │
│  0.9 ┤  ●                                                                 │
│  0.8 ┤                          ●                                          │
│  0.7 ┤                                                                     │
│  0.6 ┤                                                                     │
│  0.5 ┤                                                                     │
│      └──────┬─────────────────┬──────────────────────────────┬──────      │
│          2023-01           2023-06                         2024-01         │
│                                                                            │
│  Snapshot Details                                                          │
│  ─────────────────────────────────────────────────────────────────────     │
│  2023-01 │ mentor   │ positive │ 0.90 │ Initial mentoring established     │
│▸ 2023-06 │ mentor   │ positive │ 0.85 │ Continued mentoring, slight       │
│          │          │          │      │ distance during sabbatical         │
│  2024-01 │ mentor   │ positive │ 0.92 │ Strengthened after joint paper    │
│                                                                            │
│  Changes:  type: stable  sentiment: stable  strength: +0.02 net          │
│                                                                            │
│  esc:back  j/k:select snapshot  d:diff two snapshots              ?help   │
└────────────────────────────────────────────────────────────────────────────┘
```

### How It Works

This is a drill-down screen reached from the Relationship Explorer. It takes a `{from_person, to_person}` pair and queries all temporal snapshots of that relationship. The top half shows a simple ASCII line chart of strength over time using the `termenv`/lipgloss character drawing. The bottom half is a table of all snapshots with full detail.

The "Changes" summary at the bottom computes: whether relationship_type changed over time, whether sentiment shifted, and the net change in strength from first to last snapshot.

The `d` key enters diff mode: select two snapshots and see a side-by-side comparison of their descriptions, highlighting what changed. This uses CozoDB's temporal composite key directly — each row in the table corresponds to one `{id, from_person, to_person, timestamp}` tuple.

### Model

```yaml
screen: relationship_evolution
params:
  from_person: string          # passed from relationship_explorer
  to_person: string

model:
  chart:
    type: StrengthChart
    points: TimePoint[]        # [{timestamp, strength}]
    width: flex
    height: 8
  snapshots:
    type: table.Model
    columns:
      - { name: timestamp, width: 10 }
      - { name: relationship_type, width: 10, label: "Type" }
      - { name: sentiment, width: 10, style: sentiment_color }
      - { name: strength, width: 6, format: "%.2f" }
      - { name: description, width: flex }
    query: |
      ?[timestamp, relationship_type, sentiment, strength, description] :=
        *relationship{from_person: $from, to_person: $to,
                      timestamp, relationship_type, sentiment,
                      strength, description}
      :order timestamp
    cursor: true
  summary:
    type: ChangeSummary
    fields: [type_changed, sentiment_changed, strength_delta]
    computed_from: snapshots
  diff_mode:
    active: false
    snapshot_a: null
    snapshot_b: null

msgs:
  - SnapshotsLoaded { rows: SnapshotRow[] }
  - DiffModeToggled
  - DiffSnapshotSelected { index: int }

cmds:
  - fetchSnapshots(from_person, to_person) -> SnapshotsLoaded

keybindings:
  j: cursor_down
  k: cursor_up
  d: toggle_diff_mode
  escape: navigate_back
```

---

## Screen 5: Network Graph (ASCII)

### Mockup

```
┌─ Network ─────────────────────────────── depth:[2] layout:[force] ────────┐
│                                                                            │
│                          [James Liu]                                       │
│                          /    |    \                                        │
│                     0.9 /     |     \ 0.7                                  │
│                        /      |      \                                     │
│                [Sarah Chen]   |   [Maya Patel]                             │
│                  / \          |       |                                     │
│             0.75/   \0.6     |       | 0.6                                │
│                /     \       |       |                                     │
│          [Alex Kim] [Tom Rivera]─────+                                     │
│               \        |                                                   │
│            0.5 \       | 0.4                                               │
│                 \      |                                                   │
│               [Lisa Wang]                                                  │
│                                                                            │
│  Legend: ── positive  ╌╌ neutral  ░░ negative     line = strength          │
│                                                                            │
│  Focus: (none)                                                             │
│  ─────────────────────────────────────────────────────────────────────     │
│  Connections: 12 total, 8 positive, 3 neutral, 1 negative                 │
│                                                                            │
│  enter:focus  +/-:depth  l:layout  c:communities                  ?help   │
└────────────────────────────────────────────────────────────────────────────┘
```

### How It Works

An ASCII-rendered network graph showing all persons as nodes and relationships as edges. This is the most visually ambitious screen. The layout algorithm runs client-side (a simple force-directed layout computed in Go, not a full physics simulation — just enough iterations to produce a readable layout).

Edge style encodes sentiment: solid lines for positive, dashed for neutral, dotted/dim for negative. Edge labels show strength values. Node size (bracket width) could optionally encode degree centrality.

The depth control (`+`/`-`) controls how many hops from the focused node to display. When no node is focused, the entire graph is shown. Pressing `enter` on a node focuses it, showing only its N-hop neighborhood. The `l` key cycles layout modes: force-directed, circular, hierarchical.

The `c` key runs CozoDB's `CommunityDetectionLouvain` algorithm and color-codes nodes by community.

This screen is best-effort — ASCII graph layout is inherently limited. For >20 nodes it may need pagination or a focus-based approach.

### Model

```yaml
screen: network_graph
model:
  graph:
    type: AsciiGraph
    nodes: GraphNode[]         # [{id, name, x, y, community}]
    edges: GraphEdge[]         # [{from, to, strength, sentiment}]
    layout: "force"            # force | circular | hierarchical
    depth: 2
    focus_node: null
    viewport:
      width: flex
      height: flex
      offset_x: 0
      offset_y: 0
  stats:
    type: GraphStats
    fields: [total_edges, positive, neutral, negative]
  communities:
    active: false
    query: |
      community[person, grp] <~
        CommunityDetectionLouvain(*relationship[], from_person, to_person)
    colors: [red, green, blue, yellow, magenta, cyan]

msgs:
  - GraphDataLoaded { nodes: [], edges: [] }
  - LayoutComputed { positions: map[string]Point }
  - FocusChanged { node_id: string }
  - DepthChanged { depth: int }
  - CommunitiesLoaded { assignments: map[string]int }

cmds:
  - fetchGraphData -> GraphDataLoaded
  - computeLayout(nodes, edges, layout_type) -> LayoutComputed
  - fetchCommunities -> CommunitiesLoaded

keybindings:
  h/j/k/l: pan_viewport
  enter: focus_node
  escape: unfocus
  "+": increase_depth
  "-": decrease_depth
  l: cycle_layout
  c: toggle_communities
```

---

## Screen 6: Timeline

### Mockup

```
┌─ Timeline ──────────────────────── filter:[all▾] person:[all▾] ───────────┐
│                                                                            │
│  2024-02                                                                   │
│  ├── EVENT  Grant awarded to research lab for AI safety work              │
│  └── REL   Sarah Chen → James Liu  mentor  0.92  (▲ +0.07)              │
│                                                                            │
│  2024-01                                                                   │
│  ├── EVENT  New team member joined from partner university                │
│  ├── BEH   Alex Kim: increased_collaboration                             │
│  └── REL   Maya Patel → Tom Rivera  collaborator  0.65 (▲ +0.05)        │
│                                                                            │
│  2023-11                                                                   │
│  ├── EVENT  Paper published in Nature on collaborative research           │
│  ├── BEH   Sarah Chen: publishing                                        │
│  └── BEH   James Liu: publishing                                         │
│                                                                            │
│  2023-09                                                                   │
│  ├── EVENT  Conference presentation by team                               │
│▸ ├── REL   Alex Kim → Tom Rivera  peer  0.5  (new)                      │
│  └── BEH   Maya Patel: presenting                                        │
│                                                                            │
│  j/k:move  enter:detail  f:filter  p:person  t:type              ?help   │
└────────────────────────────────────────────────────────────────────────────┘
```

### How It Works

A chronological timeline that interleaves events, relationship changes, and behaviors into a single scrollable feed, grouped by year-month. This is the screen that makes CozoDB's temporal composite keys tangible — you see how the same relationship evolves across time alongside the events that may have caused those changes.

Each entry is tagged with its type (EVENT, REL, BEH) and color-coded accordingly. Relationship entries show the delta from the previous snapshot: `▲ +0.07` means strength increased by 0.07 since the last time this pair was recorded. `(new)` means it's the first snapshot.

Filters: type filter narrows to just events, just relationships, or just behaviors. Person filter shows only entries involving a specific person. These compose — you can filter to "only Sarah Chen's relationship changes."

The query uses a UNION of the three relations, normalized into a common format with timestamp as the sort key.

### Model

```yaml
screen: timeline
model:
  entries:
    type: list.Model
    items: TimelineEntry[]     # union type
    fields: [timestamp, entry_type, summary, delta]
    query: |
      events[timestamp, 'EVENT', description, ''] :=
        *event{timestamp, description}
      rels[timestamp, 'REL', summary, delta] :=
        *relationship{from_person, to_person, timestamp,
                      relationship_type, strength, sentiment},
        *person{id: from_person, name: from_name},
        *person{id: to_person, name: to_name},
        summary = concat(from_name, ' → ', to_name, '  ',
                         relationship_type, '  ', strength)
      behaviors[timestamp, 'BEH', summary, ''] :=
        *behavior{person_id, timestamp, behavior_type},
        *person{id: person_id, name},
        summary = concat(name, ': ', behavior_type)
      ?[timestamp, type, summary, delta] :=
        events[timestamp, type, summary, delta]
      ?[timestamp, type, summary, delta] :=
        rels[timestamp, type, summary, delta]
      ?[timestamp, type, summary, delta] :=
        behaviors[timestamp, type, summary, delta]
      :order -timestamp
    cursor: true
    group_by: year_month
  type_filter: "all"           # all | event | rel | behavior
  person_filter: null

msgs:
  - TimelineLoaded { entries: TimelineEntry[] }
  - FilterChanged { type_filter: string, person_filter: string }

cmds:
  - fetchTimeline(type_filter, person_filter) -> TimelineLoaded

keybindings:
  j: cursor_down
  k: cursor_up
  enter: navigate_to_detail   # routes based on entry_type
  f: cycle_type_filter
  p: open_person_filter
  t: same as f
  g: jump_to_top
  G: jump_to_bottom
```

---

## Screen 7: Query Console

### Mockup

```
┌─ Query Console ───────────────────────────────────────────────────────────┐
│                                                                            │
│  ┌─ Editor ───────────────────────────────────────────────────────────┐   │
│  │ 1  ?[name, rel_type, strength] :=                                  │   │
│  │ 2    *person{id: pid, name},                                       │   │
│  │ 3    *relationship{from_person: pid, relationship_type: rel_type,  │   │
│  │ 4                   strength},                                     │   │
│  │ 5    strength > 0.7                                                │   │
│  │ 6  :order -strength                                                │   │
│  │ 7  :limit 20                                                       │   │
│  └────────────────────────────────────────────────────────────────────┘   │
│                                                                            │
│  ┌─ Results (5 rows, 12ms) ───────────────────────────────────────────┐   │
│  │ name           │ rel_type      │ strength                          │   │
│  │ ───────────────┼───────────────┼─────────                          │   │
│  │ Sarah Chen     │ mentor        │ 0.92                              │   │
│  │ Sarah Chen     │ mentor        │ 0.90                              │   │
│  │ Sarah Chen     │ mentor        │ 0.85                              │   │
│  │ Sarah Chen     │ collaborator  │ 0.80                              │   │
│  │ Alex Kim       │ mentor        │ 0.75                              │   │
│  └────────────────────────────────────────────────────────────────────┘   │
│                                                                            │
│  History: [1/5]  ctrl+e:run  ctrl+h:history  ctrl+s:save         ?help   │
└────────────────────────────────────────────────────────────────────────────┘
```

### How It Works

A split-pane query console with a CozoScript editor at the top and results table at the bottom. The editor supports basic multi-line editing (not a full vim — just enough for writing queries: arrow keys, home/end, basic text insertion, backspace/delete, and line wrapping).

`ctrl+e` executes the current query. Results appear in the table below with column headers auto-detected from the query's output variables. Execution time and row count are shown in the results header.

Query history (`ctrl+h`) opens a list of previously executed queries that you can cycle through and re-execute. `ctrl+s` saves the current query to a named slot for quick recall.

Errors from CozoDB are displayed in the results pane with red styling, showing the CozoDB error message and the line/column where the error occurred.

The pane split is adjustable: `ctrl+up`/`ctrl+down` resizes the editor vs results areas.

### Model

```yaml
screen: query_console
model:
  editor:
    type: textarea.Model       # bubbles/textarea
    content: ""
    line_numbers: true
    height: 8                  # adjustable
    syntax: "cozoscript"       # for highlighting keywords
  results:
    type: table.Model
    columns: dynamic           # set from query output
    rows: []
    execution_time_ms: null
    row_count: null
    error: null
  history:
    type: list.Model
    items: HistoryEntry[]      # [{query, timestamp, row_count}]
    max_entries: 50
    active: false              # overlay when ctrl+h
  saved_queries:
    type: map[string]string    # name -> query text
  pane_split: 0.4              # fraction of height for editor

msgs:
  - QueryExecuted { columns: string[], rows: any[][], time_ms: int }
  - QueryError { message: string, line: int, col: int }
  - HistoryNavigated { index: int }
  - PaneSplitChanged { split: float }

cmds:
  - executeQuery(cozoscript: string) -> QueryExecuted | QueryError
  - loadHistory -> HistoryLoaded

keybindings:
  ctrl+e: execute_query
  ctrl+h: toggle_history_overlay
  ctrl+s: save_query_prompt
  ctrl+up: grow_editor
  ctrl+down: shrink_editor
  tab: switch_focus_editor_results
  escape: close_overlay
```

---

## Screen 8: Extraction Monitor

### Mockup

```
┌─ Extraction Monitor ──────────────────────────────────────────────────────┐
│                                                                            │
│  ┌─ Input ────────────────────────────────────────────────────────────┐   │
│  │ Source: demo_text.txt                                              │   │
│  │ Plugin: relation_extractor_v1                                      │   │
│  │ Status: ● completed (3.2s)                                         │   │
│  └────────────────────────────────────────────────────────────────────┘   │
│                                                                            │
│  ┌─ Extracted Entities ───────────────────────────────────────────────┐   │
│  │ PERSONS  (5)  Sarah Chen, James Liu, Maya Patel, Alex Kim,        │   │
│  │               Tom Rivera                                           │   │
│  │ RELS     (12) mentor(4), collaborator(3), supervisor(2),          │   │
│  │               peer(2), friend(1)                                   │   │
│  │ BEHAV    (8)  mentoring(3), publishing(2), presenting(1),         │   │
│  │               grant_writing(1), collaborating(1)                   │   │
│  │ EVENTS   (5)  lab_established, member_joined, paper_published,    │   │
│  │               conference, grant_awarded                            │   │
│  └────────────────────────────────────────────────────────────────────┘   │
│                                                                            │
│  ┌─ Import Preview ──────────────────────────────────────────────────┐   │
│  │ Will insert: 5 persons, 12 relationships, 8 behaviors, 5 events   │   │
│  │ Conflicts: 0 (no existing data)                                    │   │
│  │                                                                    │   │
│  │  [Import to CozoDB]   [Export JSON]   [Discard]                   │   │
│  └────────────────────────────────────────────────────────────────────┘   │
│                                                                            │
│  n:new extraction  r:re-run  i:import  e:export  p:plugin         ?help  │
└────────────────────────────────────────────────────────────────────────────┘
```

### How It Works

This screen drives the JS extraction pipeline from within the TUI. It integrates with the Goja-based runner — the TUI embeds the same Go code that `cozo-relationship-js-runner/main.go` uses.

The workflow: press `n` for a new extraction, which prompts for an input file and plugin selection. The extraction runs asynchronously, with the status indicator showing progress (● running → ● completed). The extracted entities panel shows a summary of what was found, grouped by type.

The import preview shows what would happen if you imported the results into CozoDB: how many entities per type, and whether any conflicts exist with existing data (based on key matching).

`i` imports the results into CozoDB (runs the `:put` queries). `e` exports raw JSON. `p` opens a plugin browser showing available extractor plugins found in the scripts directory.

The plugin system follows the contract from `plugin_loader.go`: `apiVersion`, `kind`, `id`, `name`, `create()` returning `{run()}`.

### Model

```yaml
screen: extraction_monitor
model:
  input:
    type: ExtractionInput
    source_file: null
    plugin_id: null
    plugin_name: null
    status: "idle"             # idle | running | completed | error
    elapsed_ms: null
    error: null
  results:
    type: ExtractionResults
    persons: EntitySummary[]
    relationships: EntitySummary[]
    behaviors: EntitySummary[]
    events: EntitySummary[]
  import_preview:
    type: ImportPreview
    counts: map[string]int
    conflicts: ConflictInfo[]
  plugin_browser:
    type: list.Model
    items: PluginDescriptor[]  # from plugin_loader.go discovery
    active: false              # overlay

msgs:
  - ExtractionStarted { source: string, plugin: string }
  - ExtractionProgress { phase: string, pct: float }
  - ExtractionCompleted { results: RawResults, elapsed_ms: int }
  - ExtractionError { message: string }
  - ImportCompleted { inserted: map[string]int }
  - PluginsLoaded { plugins: PluginDescriptor[] }

cmds:
  - startExtraction(source_file, plugin_id) -> ExtractionCompleted | ExtractionError
  - previewImport(results) -> ImportPreview
  - executeImport(results) -> ImportCompleted
  - discoverPlugins(scripts_dir) -> PluginsLoaded

keybindings:
  n: new_extraction_prompt
  r: rerun_last_extraction
  i: import_to_cozodb
  e: export_json
  p: toggle_plugin_browser
```

---

## Screen 9: Vector Search

### Mockup

```
┌─ Vector Search ─────────────────────────── index:[person▾] k:[10] ────────┐
│                                                                            │
│  Query: [mentoring and guidance in academic research setting            ]  │
│                                                                            │
│  Results (384-dim cosine similarity, 4ms)                                  │
│  ─────────────────────────────────────────────────────────────────────     │
│  #  Score   Entity          Description                                    │
│  1  0.923   Sarah Chen      Senior researcher and team lead. Mentors      │
│                              multiple junior members.                      │
│  2  0.891   James Liu       Lab director and experienced mentor.          │
│                              Provides career guidance.                     │
│  3  0.834   Alex Kim        Postdoctoral researcher receiving             │
│                              mentorship from senior staff.                 │
│  4  0.756   Maya Patel      Graduate student working closely with         │
│                              senior researchers.                           │
│  5  0.612   Tom Rivera      Lab technician assisting research teams.      │
│                                                                            │
│                                                                            │
│  Embedding: text-embedding-3-small (OpenAI) via Geppetto                  │
│                                                                            │
│  enter:search  tab:index  k:set-k  j/k:browse results            ?help   │
└────────────────────────────────────────────────────────────────────────────┘
```

### How It Works

Semantic search over the HNSW vector indices. The user types a natural language query, which gets embedded via the same embedding model used at extraction time (text-embedding-3-small through Geppetto), then searched against the selected HNSW index.

The index dropdown selects which entity type to search: person, relationship, behavior, or event — each has its own HNSW index with 384-dim F32 vectors and Cosine distance.

The `k` parameter controls how many nearest neighbors to return (default 10). Results show cosine similarity scores alongside entity details.

The CozoDB query uses the `~` HNSW search syntax:

```
?[entity, score] := ~person:embedding_idx{entity | query: $embedding, k: $k, score}
```

The embedding model is called once per search query. Results are near-instant since HNSW is approximate nearest neighbor.

Pressing `enter` on a result navigates to that entity's detail view (person detail, relationship detail, etc., depending on the active index).

### Model

```yaml
screen: vector_search
model:
  query_input:
    type: textinput.Model
    placeholder: "describe what you're looking for..."
    value: ""
  index: "person"              # person | relationship | behavior | event
  k: 10
  results:
    type: table.Model
    columns:
      - { name: rank, width: 3, label: "#" }
      - { name: score, width: 7, format: "%.3f" }
      - { name: name, width: 16, label: "Entity" }
      - { name: description, width: flex }
    rows: []
    cursor: true
  search_time_ms: null
  embedding_model: "text-embedding-3-small"

  # The HNSW query for each index type
  queries:
    person: |
      ?[name, score, description] :=
        ~person:person_embedding_idx{id | query: $vec, k: $k, ef: 200, score},
        *person{id, name, description}
      :order -score
    relationship: |
      ?[desc, score, sentiment, strength] :=
        ~relationship:rel_embedding_idx{id | query: $vec, k: $k, ef: 200, score},
        *relationship{id, description: desc, sentiment, strength}
      :order -score
    behavior: |
      ?[person_name, type, score] :=
        ~behavior:beh_embedding_idx{id | query: $vec, k: $k, ef: 200, score},
        *behavior{id, person_id, behavior_type: type},
        *person{id: person_id, name: person_name}
      :order -score
    event: |
      ?[desc, score, timestamp] :=
        ~event:event_embedding_idx{id | query: $vec, k: $k, ef: 200, score},
        *event{id, description: desc, timestamp}
      :order -score

msgs:
  - SearchCompleted { results: SearchResult[], time_ms: int }
  - SearchError { message: string }
  - IndexChanged { index: string }
  - KChanged { k: int }

cmds:
  - embedAndSearch(query_text, index, k) -> SearchCompleted | SearchError
    # 1. Call Geppetto to embed query_text
    # 2. Run HNSW query with resulting vector
    # 3. Return results

keybindings:
  enter: execute_search        # when input focused
  tab: cycle_index
  K: increase_k                # shift+k to avoid j/k conflict
  j: cursor_down               # in results
  k: cursor_up                 # in results
```

---

## Shared Components

These components are reused across multiple screens.

### Status Bar

```
[F1]Dashboard [F2]People [F3]Relations [F4]Timeline [F5]Network [F6]Query [F7]Extract [F8]Vector    db:cozo.db ?help
```

```yaml
component: status_bar
model:
  tabs: Tab[]                  # [{key: "F1", label: "Dashboard", screen: "dashboard"}]
  active_tab: string
  db_path: string
  help_visible: false

msgs:
  - TabSelected { screen: string }
  - HelpToggled
```

### Help Overlay

```
┌─ Keyboard Shortcuts ─────────────────────────┐
│                                                │
│  Navigation                                    │
│    F1-F8      Jump to screen                   │
│    tab        Next screen / cycle focus         │
│    escape     Back / close overlay              │
│    ?          Toggle this help                  │
│                                                │
│  Lists & Tables                                │
│    j/k        Move up/down                     │
│    g/G        Jump to top/bottom               │
│    enter      Open detail / activate            │
│    /          Search / filter                   │
│                                                │
│  Screen-specific keys shown at bottom          │
│                                                │
│                              [esc] close       │
└────────────────────────────────────────────────┘
```

```yaml
component: help_overlay
model:
  type: overlay
  visible: false
  sections:
    - title: "Navigation"
      bindings: [{key: "F1-F8", action: "Jump to screen"}, ...]
    - title: "Lists & Tables"
      bindings: [{key: "j/k", action: "Move up/down"}, ...]
    - title: "Screen-specific"
      dynamic: true            # populated by active screen
```

### Loading Spinner

```yaml
component: spinner
model:
  type: spinner.Model          # bubbles/spinner
  style: "dot"                 # dot | line | minidot | jump
  message: string
  active: bool
```

---

## Application-Level Model

The top-level bubbletea model that owns all screens and manages navigation.

```yaml
app:
  model:
    active_screen: "dashboard"
    screens:
      dashboard: DashboardModel
      people: PeopleBrowserModel
      relationships: RelationshipExplorerModel
      evolution: RelationshipEvolutionModel
      network: NetworkGraphModel
      timeline: TimelineModel
      query: QueryConsoleModel
      extraction: ExtractionMonitorModel
      vector: VectorSearchModel
    db: CozoDBConnection
    help: HelpOverlayModel
    status_bar: StatusBarModel
    window_size: { width: int, height: int }

  msgs:
    - tea.KeyMsg                # all keyboard input
    - tea.WindowSizeMsg         # terminal resize
    - NavigateTo { screen: string, params: map }
    - GlobalError { message: string }

  update_routing: |
    # Global keys handled first (F1-F8, ?, q)
    # Then delegated to active_screen.Update(msg)
    # NavigateTo messages switch active_screen and
    # trigger initial data fetch for the new screen

  cmd_pattern: |
    # All CozoDB queries wrapped in:
    #   func queryCozo(db, script, params) tea.Cmd
    # Returns a tea.Msg with results or error
    # This keeps the UI non-blocking

  layout: |
    # view() composes:
    #   1. Active screen view (fills available space)
    #   2. Status bar (bottom, 1 line)
    #   3. Help overlay (centered, if visible)
    # Uses lipgloss.Place for centering overlays
```

---

## Implementation Notes

**Package structure** (suggested, not prescriptive):

```
cmd/cozo-tui/
  main.go                      # entry point, DB connection, tea.NewProgram
internal/
  app/
    model.go                   # top-level App model
    navigation.go              # screen routing
  screens/
    dashboard/model.go
    people/model.go
    relationships/model.go
    evolution/model.go
    network/model.go
    timeline/model.go
    query/model.go
    extraction/model.go
    vector/model.go
  components/
    statusbar/model.go
    help/model.go
    spinner/model.go
  db/
    cozo.go                    # CozoDB connection wrapper
    queries.go                 # named query templates
  embedding/
    geppetto.go                # Geppetto embedding integration
```

**Dependencies**: bubbletea, bubbles, lipgloss, glamour (for markdown rendering in help), and the existing cozo-go and goja dependencies already in the repo.

**CozoDB connection**: The TUI opens a single CozoDB connection at startup (same as the JS runner does). All queries go through a shared `db.Query(script, params)` wrapper that returns `tea.Cmd` functions for async execution.
