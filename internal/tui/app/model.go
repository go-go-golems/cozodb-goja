package app

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/cozodb-goja/internal/tui/screens/dashboard"
	"github.com/go-go-golems/cozodb-goja/internal/tui/screens/query"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
)

type screen int

const (
	screenDashboard screen = iota
	screenQuery
)

type Model struct {
	db        *cozoapi.DB
	width     int
	height    int
	err       error
	screen    screen

	dashboard dashboard.Model
	query     query.Model
}

func New(db *cozoapi.DB) Model {
	return Model{
		db:        db,
		dashboard: dashboard.New(db),
		query:     query.New(db),
		screen:    screenDashboard,
	}
}

func (m Model) Init() tea.Cmd {
	return m.dashboard.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			// Only quit from dashboard; query console uses 'q' for text input
			if m.screen == screenDashboard {
				_ = m.db.Close(context.Background())
				return m, tea.Quit
			}
		case "ctrl+c":
			_ = m.db.Close(context.Background())
			return m, tea.Quit
		case "f1":
			m.screen = screenDashboard
			return m, nil
		case "f7":
			if m.screen != screenQuery {
				m.screen = screenQuery
				return m, m.query.Init()
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		contentH := msg.Height - 1 // -1 for status bar
		m.dashboard.SetSize(msg.Width, contentH)
		m.query.SetSize(msg.Width, contentH)
	case errMsg:
		m.err = msg.err
	}

	var cmd tea.Cmd
	switch m.screen {
	case screenDashboard:
		m.dashboard, cmd = m.dashboard.Update(msg)
	case screenQuery:
		m.query, cmd = m.query.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	screenLabel := "Dashboard"
	switch m.screen {
	case screenQuery:
		screenLabel = "Query"
	}

	var screenNames []string
	screens := []struct {
		key    string
		name   string
		active bool
	}{
		{"F1", "Dashboard", m.screen == screenDashboard},
		{"F7", "Query", m.screen == screenQuery},
	}
	for _, s := range screens {
		label := "[" + s.key + "]" + s.name
		if s.active {
			label = activeTabStyle.Render(label)
		} else {
			label = inactiveTabStyle.Render(label)
		}
		screenNames = append(screenNames, label)
	}

	_ = screenLabel
	tabs := lipgloss.JoinHorizontal(lipgloss.Top, screenNames...)
	statusRight := dimStatusStyle.Render("db:" + m.db.Backend())
	gap := m.width - lipgloss.Width(tabs) - lipgloss.Width(statusRight) - 2
	if gap < 0 {
		gap = 0
	}
	statusFill := statusBarStyle.Width(gap).Render("")
	statusBar := statusBarStyle.Width(m.width).Render(
		tabs + statusFill + statusRight,
	)

	var content string
	switch m.screen {
	case screenDashboard:
		content = m.dashboard.View()
	case screenQuery:
		content = m.query.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}

type errMsg struct{ err error }

var (
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ADADAD")).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1)

	activeTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ADADAD")).
				Background(lipgloss.Color("#333333")).
				Padding(0, 1)

	dimStatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1)
)
