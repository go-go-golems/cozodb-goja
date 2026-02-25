package app

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/cozodb-goja/internal/tui/screens/dashboard"
	"github.com/go-go-golems/cozodb-goja/pkg/cozoapi"
)

type Model struct {
	db        *cozoapi.DB
	dashboard dashboard.Model
	width     int
	height    int
	err       error
}

func New(db *cozoapi.DB) Model {
	return Model{
		db:        db,
		dashboard: dashboard.New(db),
	}
}

func (m Model) Init() tea.Cmd {
	return m.dashboard.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			_ = m.db.Close(context.Background())
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.dashboard.SetSize(msg.Width, msg.Height-1) // -1 for status bar
	case errMsg:
		m.err = msg.err
	}

	var cmd tea.Cmd
	m.dashboard, cmd = m.dashboard.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	statusBar := statusBarStyle.Width(m.width).Render(
		"[F1]Dashboard  [q]Quit                                              db:" + m.db.Backend(),
	)

	content := m.dashboard.View()

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}

type errMsg struct{ err error }

var statusBarStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#ADADAD")).
	Background(lipgloss.Color("#333333")).
	Padding(0, 1)
