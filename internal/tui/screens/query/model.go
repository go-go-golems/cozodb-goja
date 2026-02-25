package query

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
)

type Model struct {
	db     *cozoapi.DB
	width  int
	height int

	editor   textarea.Model
	focus    int // 0=editor, 1=results
	running  bool
	paneSplit float64 // fraction of height for editor

	// Results
	columns   []string
	rows      [][]string
	execTime  time.Duration
	rowCount  int
	queryErr  string

	// Results scroll
	resultOffset int

	// History
	history      []historyEntry
	historyIdx   int
	historyOpen  bool
	historyCursor int
}

type historyEntry struct {
	query    string
	rows     int
	duration time.Duration
	ts       time.Time
}

// Messages
type queryResultMsg struct {
	columns  []string
	rows     [][]string
	rowCount int
	duration time.Duration
}
type queryErrorMsg struct {
	err      string
	duration time.Duration
}

func New(db *cozoapi.DB) Model {
	ta := textarea.New()
	ta.Placeholder = "Enter CozoScript query..."
	ta.SetHeight(8)
	ta.ShowLineNumbers = true
	ta.Focus()
	ta.SetValue("?[name, description] := *person{name, description}\n:limit 10")

	return Model{
		db:        db,
		editor:    ta,
		paneSplit: 0.4,
		focus:     0,
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.historyOpen {
			return m.updateHistory(msg)
		}
		switch msg.String() {
		case "ctrl+e":
			if !m.running {
				m.running = true
				m.queryErr = ""
				return m, executeQuery(m.db, m.editor.Value())
			}
		case "ctrl+h":
			m.historyOpen = true
			m.historyCursor = 0
			return m, nil
		case "tab":
			m.focus = (m.focus + 1) % 2
			if m.focus == 0 {
				m.editor.Focus()
			} else {
				m.editor.Blur()
			}
			return m, nil
		case "ctrl+up":
			if m.paneSplit > 0.2 {
				m.paneSplit -= 0.1
			}
			return m, nil
		case "ctrl+down":
			if m.paneSplit < 0.8 {
				m.paneSplit += 0.1
			}
			return m, nil
		}

		// When results are focused, handle scroll
		if m.focus == 1 {
			switch msg.String() {
			case "j", "down":
				if m.resultOffset < len(m.rows)-1 {
					m.resultOffset++
				}
				return m, nil
			case "k", "up":
				if m.resultOffset > 0 {
					m.resultOffset--
				}
				return m, nil
			case "g":
				m.resultOffset = 0
				return m, nil
			case "G":
				if len(m.rows) > 0 {
					m.resultOffset = len(m.rows) - 1
				}
				return m, nil
			}
			return m, nil
		}

	case queryResultMsg:
		m.running = false
		m.columns = msg.columns
		m.rows = msg.rows
		m.rowCount = msg.rowCount
		m.execTime = msg.duration
		m.resultOffset = 0

		// Add to history
		m.history = append([]historyEntry{{
			query:    m.editor.Value(),
			rows:     msg.rowCount,
			duration: msg.duration,
			ts:       time.Now(),
		}}, m.history...)
		if len(m.history) > 50 {
			m.history = m.history[:50]
		}
		return m, nil

	case queryErrorMsg:
		m.running = false
		m.queryErr = msg.err
		m.columns = nil
		m.rows = nil
		m.execTime = msg.duration
		return m, nil
	}

	if m.focus == 0 {
		var cmd tea.Cmd
		m.editor, cmd = m.editor.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) updateHistory(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "escape", "ctrl+h":
		m.historyOpen = false
	case "j", "down":
		if m.historyCursor < len(m.history)-1 {
			m.historyCursor++
		}
	case "k", "up":
		if m.historyCursor > 0 {
			m.historyCursor--
		}
	case "enter":
		if m.historyCursor < len(m.history) {
			m.editor.SetValue(m.history[m.historyCursor].query)
			m.historyOpen = false
		}
	}
	return m, nil
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.editor.SetWidth(w - 6)
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	title := titleStyle.Render(" Query Console ")
	innerW := m.width - 4

	// Editor pane
	editorH := int(float64(m.height-6) * m.paneSplit)
	if editorH < 3 {
		editorH = 3
	}
	m.editor.SetHeight(editorH)
	m.editor.SetWidth(innerW - 2)

	editorBorder := panelStyle
	if m.focus == 0 {
		editorBorder = focusedPanelStyle
	}
	editorPane := editorBorder.Width(innerW).Render(m.editor.View())

	// Results pane
	resultsH := m.height - editorH - 7
	if resultsH < 3 {
		resultsH = 3
	}

	resultsBorder := panelStyle
	if m.focus == 1 {
		resultsBorder = focusedPanelStyle
	}
	resultsContent := m.renderResults(innerW-4, resultsH)
	resultsPane := resultsBorder.Width(innerW).Height(resultsH).Render(resultsContent)

	// Hints
	var status string
	if m.running {
		status = runningStyle.Render(" Running... ")
	} else if len(m.history) > 0 {
		status = dimStyle.Render(fmt.Sprintf(" History: %d queries ", len(m.history)))
	}
	hints := hintStyle.Render("  ctrl+e:run  tab:focus  ctrl+h:history  ctrl+up/dn:resize")

	parts := []string{title, editorPane, resultsPane, status + hints}

	if m.historyOpen {
		overlay := m.renderHistory(innerW)
		parts = append(parts, "", overlay)
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) renderResults(w, h int) string {
	if m.queryErr != "" {
		return errorStyle.Render(m.queryErr)
	}

	if m.columns == nil {
		return dimStyle.Render("Press ctrl+e to execute query")
	}

	// Header
	header := m.renderResultHeader()

	// Column widths
	colWidths := m.calcColWidths(w)

	// Header row
	headerCells := make([]string, len(m.columns))
	for i, col := range m.columns {
		cw := colWidths[i]
		headerCells[i] = colHeaderStyle.Render(padOrTruncate(col, cw))
	}
	headerRow := strings.Join(headerCells, " ")

	// Separator
	sepParts := make([]string, len(m.columns))
	for i := range m.columns {
		sepParts[i] = strings.Repeat("─", colWidths[i])
	}
	sep := dimStyle.Render(strings.Join(sepParts, "─"))

	// Data rows
	visible := h - 4 // header + colheader + sep + margin
	if visible < 1 {
		visible = 1
	}
	end := m.resultOffset + visible
	if end > len(m.rows) {
		end = len(m.rows)
	}

	var dataLines []string
	for _, row := range m.rows[m.resultOffset:end] {
		cells := make([]string, len(row))
		for i, cell := range row {
			cw := colWidths[i]
			cells[i] = padOrTruncate(cell, cw)
		}
		dataLines = append(dataLines, strings.Join(cells, " "))
	}

	parts := []string{header, headerRow, sep}
	parts = append(parts, dataLines...)

	if len(m.rows) > visible {
		scrollInfo := dimStyle.Render(fmt.Sprintf("  [%d-%d of %d]", m.resultOffset+1, end, len(m.rows)))
		parts = append(parts, scrollInfo)
	}

	return strings.Join(parts, "\n")
}

func (m Model) renderResultHeader() string {
	return dimStyle.Render(fmt.Sprintf("Results: %d rows, %s", m.rowCount, m.execTime.Round(time.Millisecond)))
}

func (m Model) calcColWidths(totalW int) []int {
	if len(m.columns) == 0 {
		return nil
	}

	// Start with column name lengths
	widths := make([]int, len(m.columns))
	for i, col := range m.columns {
		widths[i] = len(col)
	}

	// Expand based on data
	for _, row := range m.rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Cap each column
	maxCol := totalW / len(m.columns)
	if maxCol < 8 {
		maxCol = 8
	}
	for i := range widths {
		if widths[i] > maxCol {
			widths[i] = maxCol
		}
		if widths[i] < 4 {
			widths[i] = 4
		}
	}

	return widths
}

func (m Model) renderHistory(w int) string {
	lines := []string{panelTitleStyle.Render("Query History (enter to load, esc to close)")}

	if len(m.history) == 0 {
		lines = append(lines, dimStyle.Render("  (no history)"))
	}

	for i, h := range m.history {
		cursor := "  "
		style := lipgloss.NewStyle()
		if i == m.historyCursor {
			cursor = "▸ "
			style = selectedStyle
		}
		q := strings.ReplaceAll(h.query, "\n", " ")
		if len(q) > w-20 {
			q = q[:w-23] + "..."
		}
		line := fmt.Sprintf("%s%s  (%d rows, %s)", cursor, q, h.rows, h.duration.Round(time.Millisecond))
		lines = append(lines, style.Render(line))
	}

	return historyPanelStyle.Width(w).Render(strings.Join(lines, "\n"))
}

// Commands

func executeQuery(db *cozoapi.DB, script string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		start := time.Now()
		res, err := db.ExecScript(ctx, script, nil, nil)
		elapsed := time.Since(start)

		if err != nil {
			return queryErrorMsg{err: err.Error(), duration: elapsed}
		}

		// Convert rows to strings
		rows := make([][]string, len(res.Rows))
		for i, row := range res.Rows {
			cells := make([]string, len(row))
			for j, v := range row {
				cells[j] = fmt.Sprintf("%v", v)
			}
			rows[i] = cells
		}

		return queryResultMsg{
			columns:  res.Headers,
			rows:     rows,
			rowCount: len(res.Rows),
			duration: elapsed,
		}
	}
}

func padOrTruncate(s string, w int) string {
	if len(s) > w {
		return s[:w-1] + "…"
	}
	return s + strings.Repeat(" ", w-len(s))
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

	focusedPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

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

	runningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FBBF24")).
			Bold(true)

	historyPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("#FBBF24")).
				Padding(0, 1)
)
