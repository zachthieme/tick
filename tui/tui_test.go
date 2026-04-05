package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func withSize(m Model, w, h int) Model {
	updated, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return updated.(Model)
}

func TestView(t *testing.T) {
	tests := []struct {
		name         string
		hosts        int
		deadline     time.Time
		today        time.Time
		override     bool
		wantContains []string
		wantMissing  []string
	}{
		{
			name:         "normal display",
			hosts:        100,
			deadline:     date(2026, time.April, 10),
			today:        date(2026, time.April, 6),
			wantContains: []string{"hosts per night", "100 hosts left", "Deadline: 2026-04-10"},
		},
		{
			name:         "deadline passed",
			hosts:        100,
			deadline:     date(2026, time.April, 1),
			today:        date(2026, time.April, 5),
			wantContains: []string{"DEADLINE PASSED", "100 hosts left"},
		},
		{
			name:         "today override shows start date",
			hosts:        100,
			deadline:     date(2026, time.April, 10),
			today:        date(2026, time.April, 6),
			override:     true,
			wantContains: []string{"Start: 2026-04-06", "Deadline: 2026-04-10"},
		},
		{
			name:        "no override hides start date",
			hosts:       100,
			deadline:    date(2026, time.April, 10),
			today:       date(2026, time.April, 6),
			wantMissing: []string{"Start:"},
		},
		{
			name:         "weekend shows next work night message",
			hosts:        100,
			deadline:     date(2026, time.April, 10),
			today:        date(2026, time.April, 4), // Saturday
			wantContains: []string{"next work night is Monday"},
		},
		{
			name:        "weekday hides weekend message",
			hosts:       100,
			deadline:    date(2026, time.April, 10),
			today:       date(2026, time.April, 6), // Monday
			wantMissing: []string{"next work night is Monday"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(tt.hosts, "", tt.deadline, tt.today, tt.override)
			m = withSize(m, 120, 40)
			view := m.View()

			for _, want := range tt.wantContains {
				if !strings.Contains(view, want) {
					t.Errorf("View() missing %q", want)
				}
			}
			for _, missing := range tt.wantMissing {
				if strings.Contains(view, missing) {
					t.Errorf("View() should not contain %q", missing)
				}
			}
		})
	}
}

func TestUpdateTickRecalculates(t *testing.T) {
	m := New(100, "", date(2026, time.April, 10), date(2026, time.April, 6), true)
	if m.result.WeekdaysLeft != 5 {
		t.Fatalf("initial WeekdaysLeft = %d, want 5", m.result.WeekdaysLeft)
	}

	updated, cmd := m.Update(tickMsg(time.Now()))
	m = updated.(Model)

	if m.result.WeekdaysLeft != 5 {
		t.Errorf("after tick WeekdaysLeft = %d, want 5 (override keeps today fixed)", m.result.WeekdaysLeft)
	}
	if cmd == nil {
		t.Error("tick should schedule next tick")
	}
}

func TestUpdateTickWithoutOverrideUpdatesToday(t *testing.T) {
	// Use a date far in the past so it differs from time.Now().
	m := New(100, "", date(2030, time.January, 10), date(2026, time.January, 1), false)
	original := m.today

	updated, cmd := m.Update(tickMsg(time.Now()))
	m = updated.(Model)

	if m.today.Equal(original) {
		t.Error("tick without override should update today from system clock")
	}
	if cmd == nil {
		t.Error("tick should schedule next tick")
	}
}

func TestQuitKeys(t *testing.T) {
	m := New(100, "", date(2026, time.April, 10), date(2026, time.April, 6), false)

	tests := []struct {
		name string
		msg  tea.KeyMsg
	}{
		{"q", tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'q'}})},
		{"ctrl+c", tea.KeyMsg(tea.Key{Type: tea.KeyCtrlC})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cmd := m.Update(tt.msg)
			if cmd == nil {
				t.Errorf("pressing %s should produce a quit command", tt.name)
			}
		})
	}
}

func TestViewEmptyBeforeResize(t *testing.T) {
	m := New(100, "", date(2026, time.April, 10), date(2026, time.April, 6), false)
	if m.View() != "" {
		t.Error("View() should be empty before WindowSizeMsg")
	}
}

func TestViewNarrowTerminalFallback(t *testing.T) {
	m := New(100, "", date(2026, time.April, 10), date(2026, time.April, 6), false)
	// 30 columns is too narrow for colossal figlet; should fall back to plain number.
	m = withSize(m, 30, 20)
	view := m.View()

	if !strings.Contains(view, "20") {
		t.Error("narrow View() should still contain the hosts-per-night number")
	}
	if !strings.Contains(view, "hosts per night") {
		t.Error("narrow View() should still contain label")
	}
}

func TestUpdatePipeline(t *testing.T) {
	// Full pipeline: New → WindowSizeMsg → tickMsg → View.
	m := New(100, "", date(2026, time.April, 10), date(2026, time.April, 6), true)

	// Send WindowSizeMsg.
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updated.(Model)

	// Send tickMsg.
	updated, cmd := m.Update(tickMsg(time.Now()))
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("tickMsg should schedule next tick")
	}

	view := m.View()
	for _, want := range []string{"hosts per night", "100 hosts left", "Deadline: 2026-04-10"} {
		if !strings.Contains(view, want) {
			t.Errorf("pipeline View() missing %q", want)
		}
	}
}

func TestIntegrationLifecycle(t *testing.T) {
	// Full integration: Init → resize → multiple ticks → deadline crossing.
	m := New(100, "", date(2026, time.April, 10), date(2026, time.April, 6), true)

	// Init should return a tick command.
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init() should return a tick command")
	}

	// Before resize, view should be empty.
	if m.View() != "" {
		t.Fatal("View() should be empty before resize")
	}

	// Resize.
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updated.(Model)

	// First tick — normal state.
	updated, cmd = m.Update(tickMsg(time.Now()))
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("tick should schedule next tick")
	}
	view := m.View()
	if strings.Contains(view, "DEADLINE PASSED") {
		t.Error("should not show deadline passed yet")
	}
	if !strings.Contains(view, "hosts per night") {
		t.Error("should show hosts per night")
	}

	// Simulate deadline passing by changing today to after deadline.
	m.today = date(2026, time.April, 11)
	updated, _ = m.Update(tickMsg(time.Now()))
	m = updated.(Model)
	view = m.View()
	if !strings.Contains(view, "DEADLINE PASSED") {
		t.Error("should show DEADLINE PASSED after deadline")
	}
}

func TestTickReReadsHostsFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts.txt")
	os.WriteFile(path, []byte("200\n"), 0o644)

	m := New(200, path, date(2026, time.April, 10), date(2026, time.April, 6), true)
	m = withSize(m, 120, 40)

	// Verify initial state.
	if m.result.TotalHosts != 200 {
		t.Fatalf("initial TotalHosts = %d, want 200", m.result.TotalHosts)
	}

	// Update the file and send a tick.
	os.WriteFile(path, []byte("150\n"), 0o644)
	updated, _ := m.Update(tickMsg(time.Now()))
	m = updated.(Model)

	if m.totalHosts != 150 {
		t.Errorf("after tick totalHosts = %d, want 150", m.totalHosts)
	}
	if m.result.TotalHosts != 150 {
		t.Errorf("after tick result.TotalHosts = %d, want 150", m.result.TotalHosts)
	}
	if m.err != "" {
		t.Errorf("unexpected error: %s", m.err)
	}
}

func TestMaxLineWidth(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"hello", 5},
		{"ab\ncd\nefgh", 4},
		{"short\na longer line\nx", 13},
		{"\n\n", 0},
	}
	for _, tt := range tests {
		if got := maxLineWidth(tt.input); got != tt.want {
			t.Errorf("maxLineWidth(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestDoTickReturnsCommand(t *testing.T) {
	cmd := doTick()
	if cmd == nil {
		t.Fatal("doTick() should return a non-nil command")
	}
}

func TestTickHostsFileBadValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts.txt")
	os.WriteFile(path, []byte("200\n"), 0o644)

	m := New(200, path, date(2026, time.April, 10), date(2026, time.April, 6), true)

	// Write invalid content and tick — should set err but keep old host count.
	os.WriteFile(path, []byte("not-a-number\n"), 0o644)
	updated, _ := m.Update(tickMsg(time.Now()))
	m = updated.(Model)

	if m.err == "" {
		t.Error("expected error after writing invalid hosts file")
	}
	if m.totalHosts != 200 {
		t.Errorf("totalHosts should remain 200 on error, got %d", m.totalHosts)
	}
}
