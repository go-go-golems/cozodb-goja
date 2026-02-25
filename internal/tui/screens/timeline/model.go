package timeline

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

	entries    []entry
	filtered   []entry
	cursor     int
	loading    bool
	err        error
	typeFilter string // "all", "EVENT", "REL", "BEH"
	personFilter string
	filtering    bool
	filterInput  string
	scrollOffset int
}

type entry struct {
	timestamp string
	entryType string // EVENT, REL, BEH
	summary   string
	detail    string
}

// Messages
type timelineLoadedMsg struct{ entries []entry }
type timelineErrMsg struct{ err error }

func New(db *cozoapi.DB) Model {
	return Model{
		db:         db,
		typeFilter: "all",
		loading:    true,
	}
}

func (m Model) Init() tea.Cmd {
	return fetchTimeline(m.db)
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
				m.ensureVisible()
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}
		case "g":
			m.cursor = 0
			m.scrollOffset = 0
		case "G":
			if len(m.filtered) > 0 {
				m.cursor = len(m.filtered) - 1
				m.ensureVisible()
			}
		case "f", "t":
			m.cycleTypeFilter()
			m.applyFilter()
			m.cursor = 0
			m.scrollOffset = 0
		case "p":
			m.filtering = true
			m.filterInput = m.personFilter
		case "r":
			m.loading = true
			return m, fetchTimeline(m.db)
		}

	case timelineLoadedMsg:
		m.entries = msg.entries
		m.loading = false
		m.applyFilter()
		m.cursor = 0
		m.scrollOffset = 0

	case timelineErrMsg:
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
		m.scrollOffset = 0
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
	types := []string{"all", "EVENT", "REL", "BEH"}
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
	for _, e := range m.entries {
		if m.typeFilter != "all" && e.entryType != m.typeFilter {
			continue
		}
		if pq != "" && !strings.Contains(strings.ToLower(e.summary), pq) {
			continue
		}
		m.filtered = append(m.filtered, e)
	}
}

func (m *Model) ensureVisible() {
	visibleH := m.height - 8
	if visibleH < 5 {
		visibleH = 5
	}
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	if m.cursor >= m.scrollOffset+visibleH {
		m.scrollOffset = m.cursor - visibleH + 1
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

	title := titleStyle.Render(" Timeline ")
	filterInfo := dimStyle.Render(fmt.Sprintf(" filter:[%s]", m.typeFilter))
	if m.personFilter != "" {
		filterInfo += dimStyle.Render(fmt.Sprintf("  person:[%s]", m.personFilter))
	}
	if m.filtering {
		filterInfo = searchActiveStyle.Render(fmt.Sprintf(" person: %s█", m.filterInput))
	}

	if m.loading {
		return lipgloss.JoinVertical(lipgloss.Left, title+filterInfo, "  Loading...")
	}

	if m.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, title+filterInfo,
			errorStyle.Render(fmt.Sprintf("  Error: %v", m.err)))
	}

	// Group entries by timestamp and render as tree
	visibleH := m.height - 8
	if visibleH < 5 {
		visibleH = 5
	}

	var lines []string
	prevTimestamp := ""
	entryIdx := -1
	descW := m.width - 20
	if descW < 20 {
		descW = 20
	}

	for _, e := range m.filtered {
		entryIdx++

		// Group header when timestamp changes
		if e.timestamp != prevTimestamp {
			if prevTimestamp != "" {
				lines = append(lines, "")
			}
			lines = append(lines, timestampStyle.Render("  "+e.timestamp))
			prevTimestamp = e.timestamp
		}

		// Tree connector
		connector := "├── "
		// Check if last in this timestamp group
		isLast := true
		for j := entryIdx + 1; j < len(m.filtered); j++ {
			if m.filtered[j].timestamp == e.timestamp {
				isLast = false
				break
			}
		}
		if isLast {
			connector = "└── "
		}

		// Entry type badge
		badge := typeBadge(e.entryType)

		cursor := " "
		style := lipgloss.NewStyle()
		if entryIdx == m.cursor {
			cursor = "▸"
			style = selectedStyle
		}

		line := fmt.Sprintf("%s %s%s %s", cursor, connector, badge, truncate(e.summary, descW))
		lines = append(lines, style.Render(line))
	}

	if len(m.filtered) == 0 {
		lines = append(lines, dimStyle.Render("  (no entries)"))
	}

	// Apply scroll
	end := m.scrollOffset + visibleH
	if end > len(lines) {
		end = len(lines)
	}
	start := m.scrollOffset
	if start >= len(lines) {
		start = 0
	}
	visibleLines := lines[start:end]

	showing := dimStyle.Render(fmt.Sprintf("  %d entries", len(m.filtered)))
	hints := hintStyle.Render("  j/k:move  f:filter type  p:person  r:refresh  g/G:top/bottom")

	parts := []string{title + filterInfo, showing}
	parts = append(parts, visibleLines...)
	parts = append(parts, "", hints)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func typeBadge(t string) string {
	switch t {
	case "EVENT":
		return eventStyle.Render("EVENT")
	case "REL":
		return relStyle.Render("REL  ")
	case "BEH":
		return behStyle.Render("BEH  ")
	default:
		return t
	}
}

// Commands

func fetchTimeline(db *cozoapi.DB) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		var allEntries []entry

		// Events
		res, err := db.ExecScript(ctx, `?[timestamp, description] := *event{timestamp, description} :order -timestamp`, nil, nil)
		if err != nil {
			return timelineErrMsg{err}
		}
		for _, row := range res.Rows {
			if len(row) < 2 {
				continue
			}
			e := entry{entryType: "EVENT"}
			if v, ok := row[0].(string); ok {
				e.timestamp = v
			}
			if v, ok := row[1].(string); ok {
				e.summary = v
			}
			allEntries = append(allEntries, e)
		}

		// Relationships
		res, err = db.ExecScript(ctx, `?[timestamp, from_name, to_name, rel_type, strength, sentiment] :=
			*relationship{from_person, to_person, timestamp, relationship_type: rel_type, strength, sentiment},
			*person{id: from_person, name: from_name},
			*person{id: to_person, name: to_name}
			:order -timestamp`, nil, nil)
		if err != nil {
			return timelineErrMsg{err}
		}
		for _, row := range res.Rows {
			if len(row) < 6 {
				continue
			}
			e := entry{entryType: "REL"}
			if v, ok := row[0].(string); ok {
				e.timestamp = v
			}
			fromName, _ := row[1].(string)
			toName, _ := row[2].(string)
			relType, _ := row[3].(string)
			strength, _ := row[4].(float64)
			e.summary = fmt.Sprintf("%s → %s  %s  %.2f", fromName, toName, relType, strength)
			allEntries = append(allEntries, e)
		}

		// Behaviors
		res, err = db.ExecScript(ctx, `?[timestamp, name, behavior_type, description] :=
			*behavior{person_id, timestamp, behavior_type, description},
			*person{id: person_id, name}
			:order -timestamp`, nil, nil)
		if err != nil {
			return timelineErrMsg{err}
		}
		for _, row := range res.Rows {
			if len(row) < 4 {
				continue
			}
			e := entry{entryType: "BEH"}
			if v, ok := row[0].(string); ok {
				e.timestamp = v
			}
			name, _ := row[1].(string)
			behType, _ := row[2].(string)
			e.summary = fmt.Sprintf("%s: %s", name, behType)
			allEntries = append(allEntries, e)
		}

		// Sort all by timestamp descending (stable within same timestamp)
		sortEntries(allEntries)

		return timelineLoadedMsg{entries: allEntries}
	}
}

func sortEntries(entries []entry) {
	// Simple insertion sort (stable, small data)
	for i := 1; i < len(entries); i++ {
		for j := i; j > 0 && entries[j].timestamp > entries[j-1].timestamp; j-- {
			entries[j], entries[j-1] = entries[j-1], entries[j]
		}
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

	timestampStyle = lipgloss.NewStyle().
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

	searchActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FBBF24")).
				Bold(true)

	eventStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	relStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	behStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FBBF24")).
			Bold(true)
)
