// tui/tui.go
package tui

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"tick/calc"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lsferreira42/figlet-go/figlet"
	"github.com/mattn/go-runewidth"
)

var (
	bigStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

type tickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(time.Minute, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Model is the Bubble Tea model for the countdown TUI.
type Model struct {
	totalHosts    int
	hostsFile     string // if set, re-read host count from this file on each tick
	deadline      time.Time
	today         time.Time
	todayOverride bool
	result        calc.Result
	width         int
	height        int
	err           string // transient error shown in the status line
}

// New creates a new TUI model.
func New(totalHosts int, hostsFile string, deadline time.Time, today time.Time, todayOverride bool) Model {
	return Model{
		totalHosts:    totalHosts,
		hostsFile:     hostsFile,
		deadline:      deadline,
		today:         today,
		todayOverride: todayOverride,
		result:        calc.Calculate(totalHosts, deadline, today),
	}
}

// ReadHostsFile reads a positive integer from the first line of the file at path.
func ReadHostsFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(data))
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid host count %q: %w", s, err)
	}
	if n <= 0 {
		return 0, fmt.Errorf("host count must be positive, got %d", n)
	}
	return n, nil
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
		if !m.todayOverride {
			m.today = calc.DateOnly(time.Now())
		}
		if m.hostsFile != "" {
			if n, err := ReadHostsFile(m.hostsFile); err != nil {
				m.err = err.Error()
			} else {
				m.totalHosts = n
				m.err = ""
			}
		}
		m.result = calc.Calculate(m.totalHosts, m.deadline, m.today)
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
			labelStyle.Render(fmt.Sprintf("%s hosts left", calc.CommaFormat(m.result.TotalHosts))),
			labelStyle.Render(fmt.Sprintf("Deadline: %s", m.result.Deadline.Format("2006-01-02"))),
		)
	} else {
		numStr := strconv.Itoa(m.result.HostsPerNight)
		bigText := numStr
		if rendered, err := figlet.Render(numStr, figlet.WithFont("colossal")); err == nil {
			rendered = strings.TrimRight(rendered, "\n ")
			if maxLineWidth(rendered) <= m.width-4 {
				bigText = rendered
			}
		}
		bigNum := bigStyle.Render(bigText)

		lines := []string{
			bigNum,
			labelStyle.Render("hosts per night"),
			labelStyle.Render(fmt.Sprintf("%s hosts left", calc.CommaFormat(m.result.TotalHosts))),
		}

		wd := m.today.Weekday()
		if wd == time.Saturday || wd == time.Sunday {
			lines = append(lines, labelStyle.Render("next work night is Monday"))
		}

		if m.todayOverride {
			lines = append(lines, labelStyle.Render(fmt.Sprintf("Start: %s", m.today.Format("2006-01-02"))))
		}
		lines = append(lines, labelStyle.Render(fmt.Sprintf("Deadline: %s", m.result.Deadline.Format("2006-01-02"))))
		if m.err != "" {
			lines = append(lines, errStyle.Render(m.err))
		}
		content = lipgloss.JoinVertical(lipgloss.Center, lines...)
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func maxLineWidth(s string) int {
	w := 0
	for line := range strings.SplitSeq(s, "\n") {
		if n := runewidth.StringWidth(line); n > w {
			w = n
		}
	}
	return w
}
