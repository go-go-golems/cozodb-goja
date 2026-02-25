package dashboard

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

	// Data
	counts       map[string]int
	topRels      []relRow
	recentEvents []eventRow
	sentimentMap map[string]int
	avgStrength  float64

	// State
	loading    bool
	err        error
	focusPanel int // 0=counts, 1=topRels, 2=events, 3=sentiment
	cursor     int
}

type relRow struct {
	fromName string
	toName   string
	relType  string
	strength float64
}

type eventRow struct {
	timestamp   string
	description string
}

// Messages
type countsMsg struct{ counts map[string]int }
type topRelsMsg struct{ rows []relRow }
type recentEventsMsg struct{ rows []eventRow }
type sentimentMsg struct {
	dist        map[string]int
	avgStrength float64
}
type queryErrMsg struct{ err error }

func New(db *cozoapi.DB) Model {
	return Model{
		db:           db,
		counts:       map[string]int{},
		sentimentMap: map[string]int{},
		loading:      true,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchCounts(m.db),
		fetchTopRels(m.db),
		fetchRecentEvents(m.db),
		fetchSentiment(m.db),
	)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.focusPanel = (m.focusPanel + 1) % 4
			m.cursor = 0
		case "shift+tab":
			m.focusPanel = (m.focusPanel + 3) % 4
			m.cursor = 0
		case "j", "down":
			m.cursor++
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "r":
			m.loading = true
			return m, tea.Batch(
				fetchCounts(m.db),
				fetchTopRels(m.db),
				fetchRecentEvents(m.db),
				fetchSentiment(m.db),
			)
		}

	case countsMsg:
		m.counts = msg.counts
		m.loading = false
	case topRelsMsg:
		m.topRels = msg.rows
	case recentEventsMsg:
		m.recentEvents = msg.rows
	case sentimentMsg:
		m.sentimentMap = msg.dist
		m.avgStrength = msg.avgStrength
	case queryErrMsg:
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

	title := titleStyle.Render(" CozoDB Relationship Explorer ")

	// Left column: entity counts + recent events
	countsBox := m.renderCounts()
	eventsBox := m.renderEvents()
	leftCol := lipgloss.JoinVertical(lipgloss.Left, countsBox, eventsBox)

	// Right column: top relationships + sentiment
	relsBox := m.renderTopRels()
	sentBox := m.renderSentiment()
	rightCol := lipgloss.JoinVertical(lipgloss.Left, relsBox, sentBox)

	leftWidth := m.width/3 - 1
	rightWidth := m.width - leftWidth - 3

	leftCol = lipgloss.NewStyle().Width(leftWidth).Render(leftCol)
	rightCol = lipgloss.NewStyle().Width(rightWidth).Render(rightCol)

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, "  ", rightCol)

	var errLine string
	if m.err != nil {
		errLine = errorStyle.Render(fmt.Sprintf(" Error: %v ", m.err))
	}

	hints := hintStyle.Render("  tab:panel  j/k:move  r:refresh  q:quit")

	parts := []string{title, "", body}
	if errLine != "" {
		parts = append(parts, "", errLine)
	}
	parts = append(parts, "", hints)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) renderCounts() string {
	focused := m.focusPanel == 0
	style := panelStyle
	if focused {
		style = focusedPanelStyle
	}

	lines := []string{panelTitleStyle.Render("Entities")}
	entities := []struct {
		label string
		key   string
	}{
		{"Persons", "person"},
		{"Relationships", "relationship"},
		{"Behaviors", "behavior"},
		{"Events", "event"},
	}
	for _, e := range entities {
		count := m.counts[e.key]
		lines = append(lines, fmt.Sprintf("  %-16s %d", e.label, count))
	}

	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderTopRels() string {
	focused := m.focusPanel == 1
	style := panelStyle
	if focused {
		style = focusedPanelStyle
	}

	lines := []string{panelTitleStyle.Render("Top Relationships")}
	if len(m.topRels) == 0 {
		lines = append(lines, mutedStyle.Render("  (no data)"))
	}
	for i, r := range m.topRels {
		bar := renderBar(r.strength, 10)
		cursor := "  "
		rowStyle := lipgloss.NewStyle()
		if focused && i == m.cursor {
			cursor = "▸ "
			rowStyle = selectedStyle
		}
		line := fmt.Sprintf("%s%-14s→ %-14s %-10s %s",
			cursor,
			truncate(r.fromName, 14),
			truncate(r.toName, 14),
			r.relType,
			bar,
		)
		lines = append(lines, rowStyle.Render(line))
	}

	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderEvents() string {
	focused := m.focusPanel == 2
	style := panelStyle
	if focused {
		style = focusedPanelStyle
	}

	lines := []string{panelTitleStyle.Render("Recent Events")}
	if len(m.recentEvents) == 0 {
		lines = append(lines, mutedStyle.Render("  (no data)"))
	}
	for i, e := range m.recentEvents {
		cursor := "  "
		rowStyle := lipgloss.NewStyle()
		if focused && i == m.cursor {
			cursor = "▸ "
			rowStyle = selectedStyle
		}
		line := fmt.Sprintf("%s%s  %s", cursor, dimStyle.Render(e.timestamp), truncate(e.description, 30))
		lines = append(lines, rowStyle.Render(line))
	}

	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderSentiment() string {
	focused := m.focusPanel == 3
	style := panelStyle
	if focused {
		style = focusedPanelStyle
	}

	total := 0
	for _, v := range m.sentimentMap {
		total += v
	}

	lines := []string{panelTitleStyle.Render("Sentiment Distribution")}

	sentiments := []struct {
		label string
		key   string
		style lipgloss.Style
	}{
		{"Positive", "positive", positiveStyle},
		{"Neutral", "neutral", neutralStyle},
		{"Negative", "negative", negativeStyle},
	}

	for _, s := range sentiments {
		count := m.sentimentMap[s.key]
		pct := 0
		if total > 0 {
			pct = count * 100 / total
		}
		barLen := count * 20 / max(total, 1)
		bar := s.style.Render(strings.Repeat("█", barLen))
		lines = append(lines, fmt.Sprintf("  %-10s %s %d%%", s.label, bar, pct))
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  Avg strength: %.2f", m.avgStrength))

	return style.Render(strings.Join(lines, "\n"))
}

// Commands

func fetchCounts(db *cozoapi.DB) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		counts := map[string]int{}

		for _, rel := range []string{"person", "relationship", "behavior", "event"} {
			script := fmt.Sprintf("?[count(id)] := *%s{id}", rel)
			res, err := db.ExecScript(ctx, script, nil, nil)
			if err != nil {
				return queryErrMsg{err}
			}
			if len(res.Rows) > 0 && len(res.Rows[0]) > 0 {
				if v, ok := res.Rows[0][0].(float64); ok {
					counts[rel] = int(v)
				}
			}
		}
		return countsMsg{counts}
	}
}

func fetchTopRels(db *cozoapi.DB) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		script := `?[from_name, to_name, type, strength] :=
			*relationship{from_person, to_person, relationship_type: type, strength},
			*person{id: from_person, name: from_name},
			*person{id: to_person, name: to_name}
		:order -strength
		:limit 8`

		res, err := db.ExecScript(ctx, script, nil, nil)
		if err != nil {
			return queryErrMsg{err}
		}

		var rows []relRow
		for _, row := range res.Rows {
			if len(row) < 4 {
				continue
			}
			r := relRow{}
			if v, ok := row[0].(string); ok {
				r.fromName = v
			}
			if v, ok := row[1].(string); ok {
				r.toName = v
			}
			if v, ok := row[2].(string); ok {
				r.relType = v
			}
			if v, ok := row[3].(float64); ok {
				r.strength = v
			}
			rows = append(rows, r)
		}
		return topRelsMsg{rows}
	}
}

func fetchRecentEvents(db *cozoapi.DB) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		script := `?[timestamp, description] := *event{timestamp, description}
		:order -timestamp
		:limit 6`

		res, err := db.ExecScript(ctx, script, nil, nil)
		if err != nil {
			return queryErrMsg{err}
		}

		var rows []eventRow
		for _, row := range res.Rows {
			if len(row) < 2 {
				continue
			}
			e := eventRow{}
			if v, ok := row[0].(string); ok {
				e.timestamp = v
			}
			if v, ok := row[1].(string); ok {
				e.description = v
			}
			rows = append(rows, e)
		}
		return recentEventsMsg{rows}
	}
}

func fetchSentiment(db *cozoapi.DB) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Sentiment counts
		script := `?[sentiment, count(id)] := *relationship{id, sentiment}`
		res, err := db.ExecScript(ctx, script, nil, nil)
		if err != nil {
			return queryErrMsg{err}
		}

		dist := map[string]int{}
		for _, row := range res.Rows {
			if len(row) < 2 {
				continue
			}
			if k, ok := row[0].(string); ok {
				if v, ok := row[1].(float64); ok {
					dist[k] = int(v)
				}
			}
		}

		// Average strength
		script2 := `?[mean(strength)] := *relationship{strength}`
		res2, err := db.ExecScript(ctx, script2, nil, nil)
		if err != nil {
			return queryErrMsg{err}
		}

		avg := 0.0
		if len(res2.Rows) > 0 && len(res2.Rows[0]) > 0 {
			if v, ok := res2.Rows[0][0].(float64); ok {
				avg = math.Round(v*100) / 100
			}
		}

		return sentimentMsg{dist: dist, avgStrength: avg}
	}
}

// Helpers

func renderBar(strength float64, maxLen int) string {
	n := int(strength * float64(maxLen))
	if n < 1 {
		n = 1
	}
	return strengthBarStyle.Render(strings.Repeat("█", n))
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

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3D3D3D")).
			Padding(0, 1).
			MarginBottom(1)

	focusedPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Padding(0, 1).
				MarginBottom(1)

	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#3D3D3D")).
			Foreground(lipgloss.Color("#FFFFFF"))

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

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

	strengthBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
)
