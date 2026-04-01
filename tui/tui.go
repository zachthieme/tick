// tui/tui.go
package tui

import (
	"fmt"
	"strconv"
	"time"

	"countdown/calc"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/common-nighthawk/go-figure"
)

var (
	bigStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

type tickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(time.Minute, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Model is the Bubble Tea model for the countdown TUI.
type Model struct {
	TotalHosts int
	Deadline   time.Time
	result     calc.Result
	width      int
	height     int
}

// New creates a new TUI model.
func New(totalHosts int, deadline time.Time) Model {
	return Model{
		TotalHosts: totalHosts,
		Deadline:   deadline,
		result:     calc.Calculate(totalHosts, deadline, time.Now()),
	}
}

func (m Model) Init() tea.Cmd {
	return doTick()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		m.result = calc.Calculate(m.TotalHosts, m.Deadline, time.Now())
		return m, doTick()
	}
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	var content string

	if m.result.DeadlinePassed {
		content = lipgloss.JoinVertical(lipgloss.Center,
			bigStyle.Render("DEADLINE PASSED"),
			labelStyle.Render(fmt.Sprintf("%d hosts left", m.result.TotalHosts)),
			labelStyle.Render(fmt.Sprintf("Deadline: %s", m.result.Deadline.Format("2006-01-02"))),
		)
	} else {
		fig := figure.NewFigure(strconv.Itoa(m.result.HostsPerNight), "block", true)
		bigNum := bigStyle.Render(fig.String())

		content = lipgloss.JoinVertical(lipgloss.Center,
			bigNum,
			labelStyle.Render("hosts per night"),
			labelStyle.Render(fmt.Sprintf("%d hosts left", m.result.TotalHosts)),
			labelStyle.Render(fmt.Sprintf("Deadline: %s", m.result.Deadline.Format("2006-01-02"))),
		)
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
