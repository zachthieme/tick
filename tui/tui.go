// tui/tui.go
package tui

import (
	"fmt"
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
	readHosts     func() (int, error) // if non-nil, called on each tick to refresh host count
	deadline      time.Time
	today         time.Time
	todayOverride bool
	result        calc.Result
	width         int
	height        int
	err           string // transient error shown in the status line
}

// New creates a new TUI model. If readHosts is non-nil, it is called on each
// tick to refresh the host count.
func New(totalHosts int, readHosts func() (int, error), deadline time.Time, today time.Time, todayOverride bool) Model {
	return Model{
		totalHosts:    totalHosts,
		readHosts:     readHosts,
		deadline:      deadline,
		today:         today,
		todayOverride: todayOverride,
		result:        calc.Calculate(totalHosts, deadline, today),
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
		if !m.todayOverride {
			m.today = calc.DateOnly(time.Now())
		}
		if m.readHosts != nil {
			if n, err := m.readHosts(); err != nil {
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
			"",
			labelStyle.Render(fmt.Sprintf("%s hosts left", calc.CommaFormat(m.result.TotalHosts))),
			labelStyle.Render(fmt.Sprintf("Deadline: %s", m.result.Deadline.Format("2006-01-02"))),
		)
	} else {
		numStr := strconv.Itoa(m.result.HostsPerNight)
		bigText := numStr
		if rendered := renderBigNum(numStr); rendered != "" {
			if maxLineWidth(rendered) <= m.width-4 {
				bigText = rendered
			}
		}
		bigNum := bigStyle.Render(bigText)

		hostsLabel := fmt.Sprintf("%s hosts left", calc.CommaFormat(m.result.TotalHosts))
		if m.err != "" {
			hostsLabel += " (stale)"
		}

		lines := []string{
			bigNum,
			labelStyle.Render("hosts per night"),
			"",
			labelStyle.Render(hostsLabel),
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

const digitGap = " "

// renderBigNum renders each digit with figlet individually and joins them
// side-by-side with a small gap. Returns "" if rendering fails.
func renderBigNum(numStr string) string {
	var blocks [][]string
	maxLines := 0
	for _, ch := range numStr {
		rendered, err := figlet.Render(string(ch), figlet.WithFont("colossal"))
		if err != nil {
			return ""
		}
		lines := strings.Split(strings.TrimRight(rendered, "\n "), "\n")
		blocks = append(blocks, lines)
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}

	// Pad each block to the same height and normalise line widths.
	for i, blk := range blocks {
		w := 0
		for _, l := range blk {
			if n := runewidth.StringWidth(l); n > w {
				w = n
			}
		}
		for len(blk) < maxLines {
			blk = append(blk, "")
		}
		for j, l := range blk {
			if pad := w - runewidth.StringWidth(l); pad > 0 {
				blk[j] = l + strings.Repeat(" ", pad)
			}
		}
		blocks[i] = blk
	}

	// Join blocks side-by-side.
	var out []string
	for row := range maxLines {
		var parts []string
		for _, blk := range blocks {
			parts = append(parts, blk[row])
		}
		out = append(out, strings.Join(parts, digitGap))
	}
	return strings.Join(out, "\n")
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
