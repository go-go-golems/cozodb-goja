package people

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
	rows     []personRow
	filtered []personRow

	// State
	cursor    int
	loading   bool
	err       error
	sortField string
	sortDir   string
	search    string
	searching bool

	// Preview
	preview *previewData
}

type personRow struct {
	id             string
	name           string
	description    string
	firstMentioned string
	relCount       int
	behCount       int
}

type previewData struct {
	personID      string
	name          string
	description   string
	relationships []previewRel
	behaviors     []string
}

type previewRel struct {
	otherName string
	relType   string
	strength  float64
}

// Messages
type peopleLoadedMsg struct{ rows []personRow }
type previewLoadedMsg struct{ data *previewData }
type peopleErrMsg struct{ err error }

func New(db *cozoapi.DB) Model {
	return Model{
		db:        db,
		sortField: "name",
		sortDir:   "asc",
		loading:   true,
	}
}

func (m Model) Init() tea.Cmd {
	return fetchPeople(m.db)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searching {
			return m.updateSearch(msg)
		}
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				return m, m.loadPreview()
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				return m, m.loadPreview()
			}
		case "g":
			m.cursor = 0
			return m, m.loadPreview()
		case "G":
			if len(m.filtered) > 0 {
				m.cursor = len(m.filtered) - 1
				return m, m.loadPreview()
			}
		case "/":
			m.searching = true
			return m, nil
		case "s":
			m.cycleSort()
			m.applyFilter()
			m.cursor = 0
			return m, nil
		case "r":
			m.loading = true
			return m, fetchPeople(m.db)
		}

	case peopleLoadedMsg:
		m.rows = msg.rows
		m.loading = false
		m.applyFilter()
		if m.cursor >= len(m.filtered) {
			m.cursor = 0
		}
		return m, m.loadPreview()

	case previewLoadedMsg:
		m.preview = msg.data
		return m, nil

	case peopleErrMsg:
		m.err = msg.err
		m.loading = false
	}

	return m, nil
}

func (m Model) updateSearch(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "escape":
		m.searching = false
		m.search = ""
		m.applyFilter()
		m.cursor = 0
	case "enter":
		m.searching = false
	case "backspace":
		if len(m.search) > 0 {
			m.search = m.search[:len(m.search)-1]
			m.applyFilter()
			m.cursor = 0
		}
	default:
		if len(msg.String()) == 1 {
			m.search += msg.String()
			m.applyFilter()
			m.cursor = 0
		}
	}
	return m, nil
}

func (m *Model) cycleSort() {
	fields := []string{"name", "first_mentioned", "relations"}
	for i, f := range fields {
		if f == m.sortField {
			m.sortField = fields[(i+1)%len(fields)]
			return
		}
	}
	m.sortField = "name"
}

func (m *Model) applyFilter() {
	if m.search == "" {
		m.filtered = m.rows
		return
	}
	q := strings.ToLower(m.search)
	m.filtered = nil
	for _, r := range m.rows {
		if strings.Contains(strings.ToLower(r.name), q) ||
			strings.Contains(strings.ToLower(r.description), q) {
			m.filtered = append(m.filtered, r)
		}
	}
}

func (m Model) loadPreview() tea.Cmd {
	if len(m.filtered) == 0 || m.cursor >= len(m.filtered) {
		return nil
	}
	p := m.filtered[m.cursor]
	return fetchPreview(m.db, p.id, p.name, p.description)
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Title line
	title := titleStyle.Render(" People ")
	searchHint := ""
	if m.searching {
		searchHint = searchActiveStyle.Render(fmt.Sprintf(" /%s█ ", m.search))
	} else if m.search != "" {
		searchHint = dimStyle.Render(fmt.Sprintf(" /%s ", m.search))
	}

	headerLine := fmt.Sprintf("  Sort: [%s %s]  Showing %d of %d",
		m.sortField, m.sortDir, len(m.filtered), len(m.rows))

	// Table header
	nameW := 18
	dateW := 12
	relW := 9
	behW := 9
	descW := m.width - nameW - dateW - relW - behW - 10
	if descW < 10 {
		descW = 10
	}

	colHeader := fmt.Sprintf("  %-*s %-*s %*s %*s  %-s",
		nameW, "Name", dateW, "First Seen", relW, "Relations", behW, "Behaviors", "Description")
	sep := "  " + strings.Repeat("─", m.width-4)

	// Rows
	tableH := m.height - 14 // title + header + sep + preview + hints
	if tableH < 3 {
		tableH = 3
	}

	var tableRows []string
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
		line := fmt.Sprintf("%s%-*s %-*s %*d %*d  %s",
			cursor,
			nameW, truncate(r.name, nameW),
			dateW, r.firstMentioned,
			relW, r.relCount,
			behW, r.behCount,
			truncate(r.description, descW),
		)
		tableRows = append(tableRows, style.Render(line))
	}

	if len(m.filtered) == 0 {
		tableRows = append(tableRows, dimStyle.Render("  (no results)"))
	}

	// Preview
	previewBox := m.renderPreview()

	// Error
	var errLine string
	if m.err != nil {
		errLine = errorStyle.Render(fmt.Sprintf(" Error: %v ", m.err))
	}

	hints := hintStyle.Render("  j/k:move  /:search  s:sort  r:refresh  F1:dashboard  F7:query")

	parts := []string{title + searchHint, headerLine, colHeaderStyle.Render(colHeader), dimStyle.Render(sep)}
	parts = append(parts, tableRows...)
	parts = append(parts, "", previewBox)
	if errLine != "" {
		parts = append(parts, errLine)
	}
	parts = append(parts, hints)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) renderPreview() string {
	if m.preview == nil {
		return panelStyle.Width(m.width - 4).Render(dimStyle.Render("  (select a person to preview)"))
	}

	lines := []string{
		panelTitleStyle.Render(fmt.Sprintf("Preview: %s", m.preview.name)),
		"  " + m.preview.description,
	}

	if len(m.preview.relationships) > 0 {
		var relParts []string
		for _, r := range m.preview.relationships {
			relParts = append(relParts, fmt.Sprintf("%s (%s)", r.otherName, r.relType))
		}
		lines = append(lines, "  Relationships: "+strings.Join(relParts, ", "))
	}

	if len(m.preview.behaviors) > 0 {
		lines = append(lines, "  Behaviors: "+strings.Join(m.preview.behaviors, ", "))
	}

	return panelStyle.Width(m.width - 4).Render(strings.Join(lines, "\n"))
}

// Commands

func fetchPeople(db *cozoapi.DB) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		script := `
			rel_count[person_id, count(rid)] :=
				*relationship{id: rid, from_person: person_id}
			rel_count[person_id, count(rid)] :=
				*relationship{id: rid, to_person: person_id}
			beh_count[person_id, count(bid)] :=
				*behavior{id: bid, person_id}
			?[id, name, first_mentioned, description, rc, bc] :=
				*person{id, name, first_mentioned, description},
				rel_count[id, rc],
				beh_count[id, bc]
			:order name`
		res, err := db.ExecScript(ctx, script, nil, nil)
		if err != nil {
			return peopleErrMsg{err}
		}

		var rows []personRow
		for _, row := range res.Rows {
			if len(row) < 6 {
				continue
			}
			r := personRow{}
			if v, ok := row[0].(string); ok {
				r.id = v
			}
			if v, ok := row[1].(string); ok {
				r.name = v
			}
			if v, ok := row[2].(string); ok {
				r.firstMentioned = v
			}
			if v, ok := row[3].(string); ok {
				r.description = v
			}
			if v, ok := row[4].(float64); ok {
				r.relCount = int(v)
			}
			if v, ok := row[5].(float64); ok {
				r.behCount = int(v)
			}
			rows = append(rows, r)
		}
		return peopleLoadedMsg{rows}
	}
}

func fetchPreview(db *cozoapi.DB, personID, name, description string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		p := &previewData{
			personID:    personID,
			name:        name,
			description: description,
		}

		// Relationships
		relScript := fmt.Sprintf(`
			?[other_name, rel_type, strength] :=
				*relationship{from_person: %q, to_person, relationship_type: rel_type, strength},
				*person{id: to_person, name: other_name}
			?[other_name, rel_type, strength] :=
				*relationship{to_person: %q, from_person, relationship_type: rel_type, strength},
				*person{id: from_person, name: other_name}
			:order -strength
			:limit 5`, personID, personID)
		res, err := db.ExecScript(ctx, relScript, nil, nil)
		if err == nil {
			for _, row := range res.Rows {
				if len(row) < 3 {
					continue
				}
				r := previewRel{}
				if v, ok := row[0].(string); ok {
					r.otherName = v
				}
				if v, ok := row[1].(string); ok {
					r.relType = v
				}
				if v, ok := row[2].(float64); ok {
					r.strength = v
				}
				p.relationships = append(p.relationships, r)
			}
		}

		// Behaviors
		behScript := fmt.Sprintf(`
			?[behavior_type] :=
				*behavior{person_id: %q, behavior_type}`, personID)
		res, err = db.ExecScript(ctx, behScript, nil, nil)
		if err == nil {
			seen := map[string]bool{}
			for _, row := range res.Rows {
				if len(row) > 0 {
					if v, ok := row[0].(string); ok && !seen[v] {
						p.behaviors = append(p.behaviors, v)
						seen[v] = true
					}
				}
			}
		}

		return previewLoadedMsg{data: p}
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
)
