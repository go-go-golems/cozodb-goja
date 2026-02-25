package relationships

import (
	"context"
	"fmt"
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
	rows     []relRow
	filtered []relRow

	// State
	cursor       int
	loading      bool
	err          error
	typeFilter   string // "all" or specific type
	personFilter string // "" or person name substring
	filtering    bool   // typing person filter
	filterInput  string

	// Detail
	detail *detailData
}

type relRow struct {
	id          string
	fromPerson  string
	fromName    string
	toPerson    string
	toName      string
	relType     string
	sentiment   string
	strength    float64
	timestamp   string
	description string
}

type detailData struct {
	fromName      string
	toName        string
	relType       string
	description   string
	sentiment     string
	strength      float64
	snapshotCount int
	snapshotDates []string
}

// Messages
type relsLoadedMsg struct{ rows []relRow }
type detailLoadedMsg struct{ data *detailData }
type relsErrMsg struct{ err error }

func New(db *cozoapi.DB) Model {
	return Model{
		db:         db,
		typeFilter: "all",
		loading:    true,
	}
}

func (m Model) Init() tea.Cmd {
	return fetchRelationships(m.db)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.filtering {
			return m.updateFilter(msg)
		}
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				return m, m.loadDetail()
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				return m, m.loadDetail()
			}
		case "g":
			m.cursor = 0
			return m, m.loadDetail()
		case "G":
			if len(m.filtered) > 0 {
				m.cursor = len(m.filtered) - 1
			}
			return m, m.loadDetail()
		case "t":
			m.cycleTypeFilter()
			m.applyFilter()
			m.cursor = 0
			return m, m.loadDetail()
		case "p":
			m.filtering = true
			m.filterInput = m.personFilter
			return m, nil
		case "r":
			m.loading = true
			return m, fetchRelationships(m.db)
		}

	case relsLoadedMsg:
		m.rows = msg.rows
		m.loading = false
		m.applyFilter()
		if m.cursor >= len(m.filtered) {
			m.cursor = 0
		}
		return m, m.loadDetail()

	case detailLoadedMsg:
		m.detail = msg.data
		return m, nil

	case relsErrMsg:
		m.err = msg.err
		m.loading = false
	}

	return m, nil
}

func (m Model) updateFilter(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "escape":
		m.filtering = false
		m.filterInput = ""
	case "enter":
		m.filtering = false
		m.personFilter = m.filterInput
		m.applyFilter()
		m.cursor = 0
		return m, m.loadDetail()
	case "backspace":
		if len(m.filterInput) > 0 {
			m.filterInput = m.filterInput[:len(m.filterInput)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.filterInput += msg.String()
		}
	}
	return m, nil
}

func (m *Model) cycleTypeFilter() {
	types := []string{"all", "mentor", "collaborator", "supervisor", "peer"}
	for i, t := range types {
		if t == m.typeFilter {
			m.typeFilter = types[(i+1)%len(types)]
			return
		}
	}
	m.typeFilter = "all"
}

func (m *Model) applyFilter() {
	m.filtered = nil
	pq := strings.ToLower(m.personFilter)
	for _, r := range m.rows {
		if m.typeFilter != "all" && r.relType != m.typeFilter {
			continue
		}
		if pq != "" {
			if !strings.Contains(strings.ToLower(r.fromName), pq) &&
				!strings.Contains(strings.ToLower(r.toName), pq) {
				continue
			}
		}
		m.filtered = append(m.filtered, r)
	}
}

func (m Model) loadDetail() tea.Cmd {
	if len(m.filtered) == 0 || m.cursor >= len(m.filtered) {
		return nil
	}
	r := m.filtered[m.cursor]
	return fetchDetail(m.db, r.id, r.fromPerson, r.toPerson, r.fromName, r.toName, r.relType, r.description, r.sentiment, r.strength)
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	title := titleStyle.Render(" Relationships ")
	filterInfo := dimStyle.Render(fmt.Sprintf(" type:[%s]", m.typeFilter))
	if m.personFilter != "" {
		filterInfo += dimStyle.Render(fmt.Sprintf("  person:[%s]", m.personFilter))
	}
	if m.filtering {
		filterInfo = searchActiveStyle.Render(fmt.Sprintf(" person: %s█", m.filterInput))
	}

	// Column layout
	fromW := 16
	toW := 16
	typeW := 12
	sentW := 6
	strW := 5
	whenW := 10

	colHeader := fmt.Sprintf("  %-*s %-*s %-*s %-*s %*s %-*s",
		fromW, "From", toW, "To", typeW, "Type", sentW, "Sent", strW, "Str", whenW, "When")
	sep := "  " + strings.Repeat("─", m.width-4)

	// Rows
	tableH := m.height - 14
	if tableH < 3 {
		tableH = 3
	}

	var tableRows []string
	showing := fmt.Sprintf("  Showing %d of %d", len(m.filtered), len(m.rows))

	for i, r := range m.filtered {
		if i >= tableH {
			break
		}
		cursor := "  "
		style := lipgloss.NewStyle()
		if i == m.cursor {
			cursor = "▸ "
			style = selectedStyle
		}

		sentStyle := sentimentStyle(r.sentiment)
		sentText := sentStyle.Render(abbreviateSentiment(r.sentiment))

		line := fmt.Sprintf("%s%-*s %-*s %-*s %-*s %*.1f %-*s",
			cursor,
			fromW, truncate(r.fromName, fromW),
			toW, truncate(r.toName, toW),
			typeW, truncate(r.relType, typeW),
			sentW, sentText,
			strW, r.strength,
			whenW, r.timestamp,
		)
		tableRows = append(tableRows, style.Render(line))
	}

	if len(m.filtered) == 0 {
		tableRows = append(tableRows, dimStyle.Render("  (no results)"))
	}

	// Detail
	detailBox := m.renderDetail()

	// Error
	var errLine string
	if m.err != nil {
		errLine = errorStyle.Render(fmt.Sprintf(" Error: %v ", m.err))
	}

	hints := hintStyle.Render("  j/k:move  t:type  p:person  r:refresh  F1:dashboard  F7:query")

	parts := []string{title + filterInfo, showing, colHeaderStyle.Render(colHeader), dimStyle.Render(sep)}
	parts = append(parts, tableRows...)
	parts = append(parts, "", detailBox)
	if errLine != "" {
		parts = append(parts, errLine)
	}
	parts = append(parts, hints)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) renderDetail() string {
	if m.detail == nil {
		return panelStyle.Width(m.width - 4).Render(dimStyle.Render("  (select a relationship to view details)"))
	}

	d := m.detail
	header := panelTitleStyle.Render(fmt.Sprintf("%s → %s (%s)", d.fromName, d.toName, d.relType))

	var lines []string
	lines = append(lines, header)
	if d.description != "" {
		lines = append(lines, "  "+d.description)
	}

	sentStyle := sentimentStyle(d.sentiment)
	statsLine := fmt.Sprintf("  Sentiment: %s  Strength: %.2f",
		sentStyle.Render(d.sentiment), d.strength)
	lines = append(lines, statsLine)

	if d.snapshotCount > 1 {
		lines = append(lines, fmt.Sprintf("  Snapshots: %d (%s)", d.snapshotCount, strings.Join(d.snapshotDates, ", ")))
	}

	return panelStyle.Width(m.width - 4).Render(strings.Join(lines, "\n"))
}

// Commands

func fetchRelationships(db *cozoapi.DB) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		script := `?[id, from_person, from_name, to_person, to_name, relationship_type, sentiment, strength, timestamp, description] :=
			*relationship{id, from_person, to_person, relationship_type, sentiment, strength, timestamp, description},
			*person{id: from_person, name: from_name},
			*person{id: to_person, name: to_name}
			:order from_name, to_name, timestamp`

		res, err := db.ExecScript(ctx, script, nil, nil)
		if err != nil {
			return relsErrMsg{err}
		}

		var rows []relRow
		for _, row := range res.Rows {
			if len(row) < 10 {
				continue
			}
			r := relRow{}
			if v, ok := row[0].(string); ok {
				r.id = v
			}
			if v, ok := row[1].(string); ok {
				r.fromPerson = v
			}
			if v, ok := row[2].(string); ok {
				r.fromName = v
			}
			if v, ok := row[3].(string); ok {
				r.toPerson = v
			}
			if v, ok := row[4].(string); ok {
				r.toName = v
			}
			if v, ok := row[5].(string); ok {
				r.relType = v
			}
			if v, ok := row[6].(string); ok {
				r.sentiment = v
			}
			if v, ok := row[7].(float64); ok {
				r.strength = v
			}
			if v, ok := row[8].(string); ok {
				r.timestamp = v
			}
			if v, ok := row[9].(string); ok {
				r.description = v
			}
			rows = append(rows, r)
		}
		return relsLoadedMsg{rows}
	}
}

func fetchDetail(db *cozoapi.DB, relID, fromPerson, toPerson, fromName, toName, relType, description, sentiment string, strength float64) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		d := &detailData{
			fromName:    fromName,
			toName:      toName,
			relType:     relType,
			description: description,
			sentiment:   sentiment,
			strength:    strength,
		}

		// Count snapshots for this relationship pair
		script := fmt.Sprintf(`?[timestamp] :=
			*relationship{id: %q, from_person: %q, to_person: %q, timestamp}
			:order timestamp`, relID, fromPerson, toPerson)
		res, err := db.ExecScript(ctx, script, nil, nil)
		if err == nil {
			d.snapshotCount = len(res.Rows)
			for _, row := range res.Rows {
				if len(row) > 0 {
					if v, ok := row[0].(string); ok {
						d.snapshotDates = append(d.snapshotDates, v)
					}
				}
			}
		}

		return detailLoadedMsg{data: d}
	}
}

func abbreviateSentiment(s string) string {
	switch s {
	case "positive":
		return "pos"
	case "negative":
		return "neg"
	case "neutral":
		return "neu"
	default:
		return s[:3]
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

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3D3D3D")).
			Padding(0, 1)

	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#3D3D3D")).
			Foreground(lipgloss.Color("#FFFFFF"))

	colHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4672")).
			Bold(true)

	searchActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FBBF24")).
				Bold(true)

	positiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	neutralStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24"))
	negativeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4672"))
)
