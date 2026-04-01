# Countdown Timer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go TUI app that shows how many hosts to upgrade per weeknight to meet a deadline.

**Architecture:** Single binary with flag-driven mode dispatch. Pure `calc` package for weekday math, `tui` package for Bubble Tea rendering. `--once` flag switches to one-liner output mode.

**Tech Stack:** Go, Bubble Tea, Lip Gloss, go-figure (figlet rendering)

---

## File Structure

```
countdown/
├── go.mod
├── go.sum
├── main.go              # flag parsing, mode dispatch, one-shot output
├── calc/
│   ├── calc.go          # pure calculation logic
│   └── calc_test.go     # tests for calculation
└── tui/
    └── tui.go           # Bubble Tea model/update/view
```

---

### Task 1: Initialize Go Module and Dependencies

**Files:**
- Create: `go.mod`, `go.sum`

- [ ] **Step 1: Initialize module**

```bash
cd /home/zach/code/countdown
go mod init countdown
```

- [ ] **Step 2: Install dependencies**

```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/common-nighthawk/go-figure
```

- [ ] **Step 3: Commit**

```bash
git init
git add go.mod go.sum
git commit -m "chore: initialize go module with dependencies"
```

---

### Task 2: Calc Package — Write Failing Tests

**Files:**
- Create: `calc/calc_test.go`
- Create: `calc/calc.go` (minimal stub to define types)

- [ ] **Step 1: Create the type stub in calc.go**

This is needed so the test file can reference `Result`. No logic yet.

```go
// calc/calc.go
package calc

import "time"

// Result holds the output of a countdown calculation.
type Result struct {
	HostsPerNight  int
	WeekdaysLeft   int
	TotalHosts     int
	Deadline       time.Time
	DeadlinePassed bool
	IsWeekend      bool
}

// Calculate computes how many hosts to upgrade per weeknight.
func Calculate(totalHosts int, deadline time.Time, today time.Time) Result {
	return Result{}
}
```

- [ ] **Step 2: Write the test file**

```go
// calc/calc_test.go
package calc

import (
	"testing"
	"time"
)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestNormalWeekdays(t *testing.T) {
	// Mon Apr 6 to Fri Apr 10 = 5 weekdays (both inclusive)
	// 100 hosts / 5 nights = 20 per night
	result := Calculate(100, date(2026, time.April, 10), date(2026, time.April, 6))
	if result.HostsPerNight != 20 {
		t.Errorf("expected 20 hosts/night, got %d", result.HostsPerNight)
	}
	if result.WeekdaysLeft != 5 {
		t.Errorf("expected 5 weekdays left, got %d", result.WeekdaysLeft)
	}
	if result.DeadlinePassed {
		t.Error("expected DeadlinePassed to be false")
	}
	if result.IsWeekend {
		t.Error("expected IsWeekend to be false")
	}
}

func TestCeilDivision(t *testing.T) {
	// Mon Apr 6 to Fri Apr 10 = 5 weekdays
	// 101 hosts / 5 nights = 21 per night (ceil)
	result := Calculate(101, date(2026, time.April, 10), date(2026, time.April, 6))
	if result.HostsPerNight != 21 {
		t.Errorf("expected 21 hosts/night, got %d", result.HostsPerNight)
	}
}

func TestSpansWeekend(t *testing.T) {
	// Fri Apr 3 to Tue Apr 7 = Fri + Mon + Tue = 3 weekdays
	// 90 hosts / 3 = 30 per night
	result := Calculate(90, date(2026, time.April, 7), date(2026, time.April, 3))
	if result.HostsPerNight != 30 {
		t.Errorf("expected 30 hosts/night, got %d", result.HostsPerNight)
	}
	if result.WeekdaysLeft != 3 {
		t.Errorf("expected 3 weekdays left, got %d", result.WeekdaysLeft)
	}
}

func TestTodayIsWeekend(t *testing.T) {
	// Sat Apr 4 to Fri Apr 10 = Mon-Fri = 5 weekdays
	result := Calculate(100, date(2026, time.April, 10), date(2026, time.April, 4))
	if !result.IsWeekend {
		t.Error("expected IsWeekend to be true")
	}
	if result.WeekdaysLeft != 5 {
		t.Errorf("expected 5 weekdays left, got %d", result.WeekdaysLeft)
	}
	if result.HostsPerNight != 20 {
		t.Errorf("expected 20 hosts/night, got %d", result.HostsPerNight)
	}
}

func TestDeadlinePassed(t *testing.T) {
	result := Calculate(100, date(2026, time.April, 1), date(2026, time.April, 5))
	if !result.DeadlinePassed {
		t.Error("expected DeadlinePassed to be true")
	}
}

func TestDeadlineIsToday(t *testing.T) {
	// Today is the deadline, today is a weekday (Mon) = 1 weekday
	// All hosts tonight
	result := Calculate(100, date(2026, time.April, 6), date(2026, time.April, 6))
	if result.HostsPerNight != 100 {
		t.Errorf("expected 100 hosts/night, got %d", result.HostsPerNight)
	}
	if result.WeekdaysLeft != 1 {
		t.Errorf("expected 1 weekday left, got %d", result.WeekdaysLeft)
	}
}

func TestReturnsTotalHostsAndDeadline(t *testing.T) {
	dl := date(2026, time.April, 10)
	result := Calculate(500, dl, date(2026, time.April, 6))
	if result.TotalHosts != 500 {
		t.Errorf("expected TotalHosts 500, got %d", result.TotalHosts)
	}
	if !result.Deadline.Equal(dl) {
		t.Errorf("expected Deadline %v, got %v", dl, result.Deadline)
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

```bash
cd /home/zach/code/countdown && go test ./calc/ -v
```

Expected: Most tests FAIL (the stub returns zero values).

- [ ] **Step 4: Commit**

```bash
git add calc/
git commit -m "test: add calc package tests for weekday counting and edge cases"
```

---

### Task 3: Calc Package — Implement to Pass Tests

**Files:**
- Modify: `calc/calc.go`

- [ ] **Step 1: Implement the Calculate function**

Replace the stub body in `calc/calc.go` with the full implementation:

```go
// calc/calc.go
package calc

import "time"

// Result holds the output of a countdown calculation.
type Result struct {
	HostsPerNight  int
	WeekdaysLeft   int
	TotalHosts     int
	Deadline       time.Time
	DeadlinePassed bool
	IsWeekend      bool
}

// Calculate computes how many hosts to upgrade per weeknight.
func Calculate(totalHosts int, deadline time.Time, today time.Time) Result {
	r := Result{
		TotalHosts: totalHosts,
		Deadline:   deadline,
	}

	if today.After(deadline) {
		r.DeadlinePassed = true
		return r
	}

	wd := today.Weekday()
	r.IsWeekend = wd == time.Saturday || wd == time.Sunday

	// Count weekdays from today through deadline, inclusive of both
	weekdays := 0
	for d := today; !d.After(deadline); d = d.AddDate(0, 0, 1) {
		w := d.Weekday()
		if w != time.Saturday && w != time.Sunday {
			weekdays++
		}
	}
	r.WeekdaysLeft = weekdays

	if weekdays > 0 {
		r.HostsPerNight = (totalHosts + weekdays - 1) / weekdays
	} else {
		r.HostsPerNight = totalHosts
	}

	return r
}
```

- [ ] **Step 2: Run tests to verify they pass**

```bash
cd /home/zach/code/countdown && go test ./calc/ -v
```

Expected: All 7 tests PASS.

- [ ] **Step 3: Commit**

```bash
git add calc/calc.go
git commit -m "feat: implement weekday calculation with ceil division"
```

---

### Task 4: TUI Package

**Files:**
- Create: `tui/tui.go`

- [ ] **Step 1: Write the Bubble Tea model**

```go
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
```

- [ ] **Step 2: Verify it compiles**

```bash
cd /home/zach/code/countdown && go build ./tui/
```

Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add tui/
git commit -m "feat: add Bubble Tea TUI with big-number display"
```

---

### Task 5: Main — Flag Parsing and Mode Dispatch

**Files:**
- Create: `main.go`

- [ ] **Step 1: Write main.go**

```go
// main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"countdown/calc"
	"countdown/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	hosts := flag.Int("hosts", 0, "total number of hosts to upgrade (required)")
	deadlineStr := flag.String("deadline", "", "target date in YYYY-MM-DD format (required)")
	once := flag.Bool("once", false, "print a one-liner and exit")
	flag.Parse()

	if *hosts <= 0 || *deadlineStr == "" {
		fmt.Fprintln(os.Stderr, "Usage: countdown --hosts N --deadline YYYY-MM-DD [--once]")
		os.Exit(1)
	}

	deadline, err := time.Parse("2006-01-02", *deadlineStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid deadline format: %s (expected YYYY-MM-DD)\n", *deadlineStr)
		os.Exit(1)
	}

	if *once {
		result := calc.Calculate(*hosts, deadline, time.Now())
		if result.DeadlinePassed {
			fmt.Printf("Deadline passed — %d hosts remaining (deadline was %s)\n",
				result.TotalHosts, result.Deadline.Format("2006-01-02"))
		} else {
			fmt.Printf("%d weekdays remaining — %d hosts/night (%d hosts, deadline %s)\n",
				result.WeekdaysLeft, result.HostsPerNight, result.TotalHosts,
				result.Deadline.Format("2006-01-02"))
		}
		return
	}

	m := tui.New(*hosts, deadline)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd /home/zach/code/countdown && go build -o countdown .
```

Expected: Binary `countdown` is created with no errors.

- [ ] **Step 3: Run all tests**

```bash
cd /home/zach/code/countdown && go test ./... -v
```

Expected: All tests pass.

- [ ] **Step 4: Smoke test one-shot mode**

```bash
cd /home/zach/code/countdown && ./countdown --hosts 500 --deadline 2026-04-30 --once
```

Expected: A single line like `21 weekdays remaining — 24 hosts/night (500 hosts, deadline 2026-04-30)`

- [ ] **Step 5: Smoke test TUI mode**

```bash
cd /home/zach/code/countdown && ./countdown --hosts 500 --deadline 2026-04-30
```

Expected: Full-screen TUI with big number, centered layout. Press `q` to exit.

- [ ] **Step 6: Commit**

```bash
git add main.go
git commit -m "feat: add main with flag parsing, one-shot and TUI modes"
```

---

### Task 6: Clean Up

- [ ] **Step 1: Run go mod tidy**

```bash
cd /home/zach/code/countdown && go mod tidy
```

- [ ] **Step 2: Final full test run**

```bash
cd /home/zach/code/countdown && go test ./... -v
```

Expected: All tests pass.

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "chore: tidy module"
```
