package network

import (
	"context"
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
)

type Model struct {
	db     *cozoapi.DB
	width  int
	height int

	nodes   []node
	edges   []edge
	loading bool
	err     error

	focusIdx int // -1 = no focus, index into nodes
	depth    int
	cursor   int // for node selection
}

type node struct {
	id   string
	name string
	x, y float64 // layout coordinates [0,1]
}

type edge struct {
	fromIdx   int
	toIdx     int
	strength  float64
	sentiment string
}

// Messages
type graphLoadedMsg struct {
	nodes []node
	edges []edge
}
type graphErrMsg struct{ err error }

func New(db *cozoapi.DB) Model {
	return Model{
		db:       db,
		loading:  true,
		focusIdx: -1,
		depth:    2,
	}
}

func (m Model) Init() tea.Cmd {
	return fetchGraph(m.db)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.nodes)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			if m.cursor < len(m.nodes) {
				if m.focusIdx == m.cursor {
					m.focusIdx = -1 // unfocus
				} else {
					m.focusIdx = m.cursor
				}
			}
		case "escape":
			m.focusIdx = -1
		case "+", "=":
			if m.depth < 5 {
				m.depth++
			}
		case "-":
			if m.depth > 1 {
				m.depth--
			}
		case "r":
			m.loading = true
			return m, fetchGraph(m.db)
		}

	case graphLoadedMsg:
		m.nodes = msg.nodes
		m.edges = msg.edges
		m.loading = false
		m.layoutCircular()

	case graphErrMsg:
		m.err = msg.err
		m.loading = false
	}

	return m, nil
}

func (m *Model) layoutCircular() {
	n := len(m.nodes)
	if n == 0 {
		return
	}
	for i := range m.nodes {
		angle := 2 * math.Pi * float64(i) / float64(n)
		m.nodes[i].x = 0.5 + 0.4*math.Cos(angle)
		m.nodes[i].y = 0.5 + 0.4*math.Sin(angle)
	}
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	focusLabel := "(none)"
	if m.focusIdx >= 0 && m.focusIdx < len(m.nodes) {
		focusLabel = m.nodes[m.focusIdx].name
	}

	title := titleStyle.Render(" Network ")
	info := dimStyle.Render(fmt.Sprintf(" depth:[%d]  focus:[%s]", m.depth, focusLabel))

	if m.loading {
		return lipgloss.JoinVertical(lipgloss.Left, title+info, "  Loading...")
	}
	if m.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, title+info,
			errorStyle.Render(fmt.Sprintf("  Error: %v", m.err)))
	}

	// Determine visible nodes/edges based on focus
	visibleNodes, visibleEdges := m.getVisible()

	// Render ASCII graph
	graphH := m.height - 14
	if graphH < 8 {
		graphH = 8
	}
	graphW := m.width - 4
	graph := m.renderGraph(visibleNodes, visibleEdges, graphW, graphH)

	// Node list
	nodeList := m.renderNodeList(visibleNodes)

	// Stats
	pos, neu, neg := 0, 0, 0
	for _, e := range visibleEdges {
		switch e.sentiment {
		case "positive":
			pos++
		case "neutral":
			neu++
		case "negative":
			neg++
		}
	}
	stats := dimStyle.Render(fmt.Sprintf("  Connections: %d total, %d positive, %d neutral, %d negative",
		len(visibleEdges), pos, neu, neg))

	legend := dimStyle.Render("  Legend: ── positive  ╌╌ neutral  ░░ negative")
	hints := hintStyle.Render("  j/k:select  enter:focus  esc:unfocus  +/-:depth  r:refresh")

	parts := []string{title + info, "", graph, "", nodeList, "", stats, legend, hints}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) getVisible() ([]node, []edge) {
	if m.focusIdx < 0 || m.focusIdx >= len(m.nodes) {
		return m.nodes, m.edges
	}

	// BFS from focus node up to depth
	visible := map[int]bool{m.focusIdx: true}
	frontier := []int{m.focusIdx}

	for d := 0; d < m.depth; d++ {
		var nextFrontier []int
		for _, ni := range frontier {
			for _, e := range m.edges {
				other := -1
				if e.fromIdx == ni {
					other = e.toIdx
				} else if e.toIdx == ni {
					other = e.fromIdx
				}
				if other >= 0 && !visible[other] {
					visible[other] = true
					nextFrontier = append(nextFrontier, other)
				}
			}
		}
		frontier = nextFrontier
	}

	var vNodes []node
	nodeMap := map[int]int{} // old idx -> new idx
	for i, n := range m.nodes {
		if visible[i] {
			nodeMap[i] = len(vNodes)
			vNodes = append(vNodes, n)
		}
	}

	var vEdges []edge
	for _, e := range m.edges {
		if visible[e.fromIdx] && visible[e.toIdx] {
			vEdges = append(vEdges, edge{
				fromIdx:   nodeMap[e.fromIdx],
				toIdx:     nodeMap[e.toIdx],
				strength:  e.strength,
				sentiment: e.sentiment,
			})
		}
	}

	return vNodes, vEdges
}

func (m Model) renderGraph(nodes []node, edges []edge, w, h int) string {
	if len(nodes) == 0 {
		return dimStyle.Render("  (no nodes)")
	}

	// Create grid
	grid := make([][]rune, h)
	colors := make([][]lipgloss.Style, h)
	for i := range grid {
		grid[i] = make([]rune, w)
		colors[i] = make([]lipgloss.Style, w)
		for j := range grid[i] {
			grid[i][j] = ' '
			colors[i][j] = dimStyle
		}
	}

	// Place edges first
	for _, e := range edges {
		if e.fromIdx >= len(nodes) || e.toIdx >= len(nodes) {
			continue
		}
		from := nodes[e.fromIdx]
		to := nodes[e.toIdx]

		x1 := int(from.x * float64(w-1))
		y1 := int(from.y * float64(h-1))
		x2 := int(to.x * float64(w-1))
		y2 := int(to.y * float64(h-1))

		edgeChar := '─'
		style := positiveStyle
		switch e.sentiment {
		case "neutral":
			edgeChar = '╌'
			style = neutralStyle
		case "negative":
			edgeChar = '░'
			style = negativeStyle
		}

		// Bresenham-like line
		drawLine(grid, colors, x1, y1, x2, y2, edgeChar, style)
	}

	// Place nodes (overwrite edges)
	for i, n := range nodes {
		x := int(n.x * float64(w-1))
		y := int(n.y * float64(h-1))

		label := "[" + n.name + "]"
		startX := x - len(label)/2
		style := nodeStyle
		if i == m.cursor {
			style = focusedNodeStyle
		}

		for ci, ch := range label {
			cx := startX + ci
			if cx >= 0 && cx < w && y >= 0 && y < h {
				grid[y][cx] = ch
				colors[y][cx] = style
			}
		}
	}

	// Render
	var lines []string
	for y := 0; y < h; y++ {
		var sb strings.Builder
		for x := 0; x < w; x++ {
			sb.WriteString(colors[y][x].Render(string(grid[y][x])))
		}
		lines = append(lines, "  "+sb.String())
	}

	return strings.Join(lines, "\n")
}

func drawLine(grid [][]rune, colors [][]lipgloss.Style, x1, y1, x2, y2 int, ch rune, style lipgloss.Style) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx, sy := 1, 1
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy

	h := len(grid)
	w := len(grid[0])

	for i := 0; i < dx+dy+1; i++ {
		if x1 >= 0 && x1 < w && y1 >= 0 && y1 < h && grid[y1][x1] == ' ' {
			grid[y1][x1] = ch
			colors[y1][x1] = style
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func (m Model) renderNodeList(nodes []node) string {
	var parts []string
	for i, n := range nodes {
		style := dimStyle
		prefix := "  "
		if i == m.cursor {
			style = selectedStyle
			prefix = "▸ "
		}
		if m.focusIdx >= 0 && m.focusIdx < len(m.nodes) && m.nodes[m.focusIdx].id == n.id {
			style = focusedNodeStyle
		}
		parts = append(parts, style.Render(prefix+n.name))
	}
	return "  " + strings.Join(parts, "  ")
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Commands

func fetchGraph(db *cozoapi.DB) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Nodes
		res, err := db.ExecScript(ctx, `?[id, name] := *person{id, name} :order name`, nil, nil)
		if err != nil {
			return graphErrMsg{err}
		}

		nodeMap := map[string]int{}
		var nodes []node
		for _, row := range res.Rows {
			if len(row) < 2 {
				continue
			}
			n := node{}
			if v, ok := row[0].(string); ok {
				n.id = v
			}
			if v, ok := row[1].(string); ok {
				n.name = v
			}
			nodeMap[n.id] = len(nodes)
			nodes = append(nodes, n)
		}

		// Edges (all snapshots, deduplicate to latest per pair in Go)
		res, err = db.ExecScript(ctx, `
			?[from_p, to_p, timestamp, strength, sentiment] :=
				*relationship{from_person: from_p, to_person: to_p, timestamp, strength, sentiment}
			:order from_p, to_p, -timestamp`, nil, nil)
		if err != nil {
			return graphErrMsg{err}
		}

		type pairKey struct{ from, to string }
		seen := map[pairKey]bool{}
		var edges []edge
		for _, row := range res.Rows {
			if len(row) < 5 {
				continue
			}
			fromID, _ := row[0].(string)
			toID, _ := row[1].(string)
			pk := pairKey{fromID, toID}
			if seen[pk] {
				continue // skip older snapshots
			}
			seen[pk] = true

			fi, ok1 := nodeMap[fromID]
			ti, ok2 := nodeMap[toID]
			if !ok1 || !ok2 {
				continue
			}
			e := edge{fromIdx: fi, toIdx: ti}
			if v, ok := row[3].(float64); ok {
				e.strength = v
			}
			if v, ok := row[4].(string); ok {
				e.sentiment = v
			}
			edges = append(edges, e)
		}

		return graphLoadedMsg{nodes: nodes, edges: edges}
	}
}

// Styles

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	nodeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	focusedNodeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7D56F4"))

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#3D3D3D")).
			Foreground(lipgloss.Color("#FFFFFF"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4672")).
			Bold(true)

	positiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	neutralStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24"))
	negativeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4672"))
)
