package evolution

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

	fromPerson string
	toPerson   string
	fromName   string
	toName     string

	snapshots []snapshot
	cursor    int
	loading   bool
	err       error
}

type snapshot struct {
	timestamp   string
	relType     string
	sentiment   string
	strength    float64
	description string
}

// Messages
type snapshotsLoadedMsg struct {
	fromName  string
	toName    string
	snapshots []snapshot
}
type evolutionErrMsg struct{ err error }

// NavigateMsg is sent to the app to switch to this screen with parameters.
type NavigateMsg struct {
	FromPerson string
	ToPerson   string
}

func New(db *cozoapi.DB) Model {
	return Model{db: db, loading: true}
}

func (m *Model) SetParams(fromPerson, toPerson string) {
	m.fromPerson = fromPerson
	m.toPerson = toPerson
	m.loading = true
	m.snapshots = nil
	m.cursor = 0
}

func (m Model) Init() tea.Cmd {
	if m.fromPerson == "" {
		return nil
	}
	return fetchSnapshots(m.db, m.fromPerson, m.toPerson)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.snapshots)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		}

	case snapshotsLoadedMsg:
		m.fromName = msg.fromName
		m.toName = msg.toName
		m.snapshots = msg.snapshots
		m.loading = false
		m.cursor = 0

	case evolutionErrMsg:
		m.err = msg.err
		m.loading = false
	}

	return m, nil
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if m.fromPerson == "" {
		return dimStyle.Render("  No relationship selected. Press F3 to browse relationships, then enter to view evolution.")
	}

	title := titleStyle.Render(fmt.Sprintf(" Evolution: %s → %s ", m.fromName, m.toName))

	if m.loading {
		return lipgloss.JoinVertical(lipgloss.Left, title, "  Loading...")
	}

	if m.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, title, errorStyle.Render(fmt.Sprintf("  Error: %v", m.err)))
	}

	if len(m.snapshots) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, title, dimStyle.Render("  No snapshots found"))
	}

	// ASCII chart
	chart := m.renderChart()

	// Snapshot table
	table := m.renderTable()

	// Summary
	summary := m.renderSummary()

	hints := hintStyle.Render("  j/k:select snapshot  esc:back  F3:relationships")

	parts := []string{title, "", chart, "", panelTitleStyle.Render("  Snapshot Details"), table, "", summary, hints}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) renderChart() string {
	if len(m.snapshots) == 0 {
		return ""
	}

	chartH := 8
	chartW := m.width - 12

	// Find min/max strength
	minS, maxS := 1.0, 0.0
	for _, s := range m.snapshots {
		if s.strength < minS {
			minS = s.strength
		}
		if s.strength > maxS {
			maxS = s.strength
		}
	}
	// Round to nice range
	minS = math.Floor(minS*10) / 10
	maxS = math.Ceil(maxS*10) / 10
	if maxS-minS < 0.2 {
		minS = maxS - 0.3
		if minS < 0 {
			minS = 0
		}
	}
	rangeS := maxS - minS
	if rangeS == 0 {
		rangeS = 1
	}

	// Build grid
	grid := make([][]rune, chartH)
	for i := range grid {
		grid[i] = make([]rune, chartW)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// Place points
	for i, s := range m.snapshots {
		x := 0
		if len(m.snapshots) > 1 {
			x = i * (chartW - 1) / (len(m.snapshots) - 1)
		}
		y := chartH - 1 - int((s.strength-minS)/rangeS*float64(chartH-1))
		if y < 0 {
			y = 0
		}
		if y >= chartH {
			y = chartH - 1
		}
		if x >= 0 && x < chartW {
			grid[y][x] = '●'
		}

		// Connect to previous point
		if i > 0 {
			prevX := (i - 1) * (chartW - 1) / (len(m.snapshots) - 1)
			prevY := chartH - 1 - int((m.snapshots[i-1].strength-minS)/rangeS*float64(chartH-1))
			if prevY < 0 {
				prevY = 0
			}
			if prevY >= chartH {
				prevY = chartH - 1
			}
			// Simple line between points
			steps := x - prevX
			if steps > 0 {
				for step := 1; step < steps; step++ {
					cx := prevX + step
					t := float64(step) / float64(steps)
					cy := prevY + int(t*float64(y-prevY))
					if cy >= 0 && cy < chartH && cx >= 0 && cx < chartW && grid[cy][cx] == ' ' {
						grid[cy][cx] = '─'
					}
				}
			}
		}
	}

	// Render with Y axis labels
	var lines []string
	lines = append(lines, panelTitleStyle.Render("  Strength over time"))
	for row := 0; row < chartH; row++ {
		val := maxS - float64(row)/float64(chartH-1)*rangeS
		label := fmt.Sprintf("  %4.1f ┤", val)
		lines = append(lines, dimStyle.Render(label)+chartStyle.Render(string(grid[row])))
	}

	// X axis with timestamps
	axisLine := "       └"
	for i := 0; i < chartW; i++ {
		axisLine += "─"
	}
	lines = append(lines, dimStyle.Render(axisLine))

	// Timestamp labels
	if len(m.snapshots) > 0 {
		labelLine := "        "
		for i, s := range m.snapshots {
			x := 0
			if len(m.snapshots) > 1 {
				x = i * (chartW - 1) / (len(m.snapshots) - 1)
			}
			pos := x - len(s.timestamp)/2
			for len(labelLine) < pos+8 {
				labelLine += " "
			}
			labelLine += s.timestamp
		}
		lines = append(lines, dimStyle.Render(labelLine))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderTable() string {
	sep := "  " + strings.Repeat("─", m.width-4)

	header := fmt.Sprintf("  %-10s %-12s %-10s %6s  %s",
		"When", "Type", "Sentiment", "Str", "Description")

	var lines []string
	lines = append(lines, colHeaderStyle.Render(header))
	lines = append(lines, dimStyle.Render(sep))

	for i, s := range m.snapshots {
		cursor := "  "
		style := lipgloss.NewStyle()
		if i == m.cursor {
			cursor = "▸ "
			style = selectedStyle
		}
		sentStyle := sentimentStyle(s.sentiment)
		descW := m.width - 48
		if descW < 10 {
			descW = 10
		}
		line := fmt.Sprintf("%s%-10s %-12s %s %6.2f  %s",
			cursor, s.timestamp, s.relType,
			sentStyle.Render(fmt.Sprintf("%-10s", s.sentiment)),
			s.strength,
			truncate(s.description, descW))
		lines = append(lines, style.Render(line))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderSummary() string {
	if len(m.snapshots) < 2 {
		return ""
	}

	first := m.snapshots[0]
	last := m.snapshots[len(m.snapshots)-1]
	delta := last.strength - first.strength
	deltaStr := fmt.Sprintf("%+.2f", delta)
	if delta > 0 {
		deltaStr = positiveStyle.Render(deltaStr)
	} else if delta < 0 {
		deltaStr = negativeStyle.Render(deltaStr)
	}

	typeChanged := "stable"
	if first.relType != last.relType {
		typeChanged = fmt.Sprintf("%s → %s", first.relType, last.relType)
	}

	sentChanged := "stable"
	if first.sentiment != last.sentiment {
		sentChanged = fmt.Sprintf("%s → %s", first.sentiment, last.sentiment)
	}

	return dimStyle.Render(fmt.Sprintf("  Changes:  type: %s  sentiment: %s  strength: %s net", typeChanged, sentChanged, deltaStr))
}

// Commands

func fetchSnapshots(db *cozoapi.DB, fromPerson, toPerson string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Get names
		nameScript := fmt.Sprintf(`?[name] := *person{id: %q, name}`, fromPerson)
		fromName := fromPerson
		res, err := db.ExecScript(ctx, nameScript, nil, nil)
		if err == nil && len(res.Rows) > 0 {
			if v, ok := res.Rows[0][0].(string); ok {
				fromName = v
			}
		}

		nameScript = fmt.Sprintf(`?[name] := *person{id: %q, name}`, toPerson)
		toName := toPerson
		res, err = db.ExecScript(ctx, nameScript, nil, nil)
		if err == nil && len(res.Rows) > 0 {
			if v, ok := res.Rows[0][0].(string); ok {
				toName = v
			}
		}

		// Get snapshots
		script := fmt.Sprintf(`?[timestamp, relationship_type, sentiment, strength, description] :=
			*relationship{from_person: %q, to_person: %q,
			              timestamp, relationship_type, sentiment,
			              strength, description}
			:order timestamp`, fromPerson, toPerson)

		res, err = db.ExecScript(ctx, script, nil, nil)
		if err != nil {
			return evolutionErrMsg{err}
		}

		var snaps []snapshot
		for _, row := range res.Rows {
			if len(row) < 5 {
				continue
			}
			s := snapshot{}
			if v, ok := row[0].(string); ok {
				s.timestamp = v
			}
			if v, ok := row[1].(string); ok {
				s.relType = v
			}
			if v, ok := row[2].(string); ok {
				s.sentiment = v
			}
			if v, ok := row[3].(float64); ok {
				s.strength = v
			}
			if v, ok := row[4].(string); ok {
				s.description = v
			}
			snaps = append(snaps, s)
		}

		return snapshotsLoadedMsg{fromName: fromName, toName: toName, snapshots: snaps}
	}
}

func sentimentStyle(s string) lipgloss.Style {
	switch s {
	case "positive":
		return positiveStyle
	case "negative":
		return negativeStyle
	default:
		return neutralStyle
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

// Styles

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	colHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

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

	chartStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))

	positiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	neutralStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24"))
	negativeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4672"))
)
