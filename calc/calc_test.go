// calc/calc_test.go
package calc

import (
	"testing"
	"time"
)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestCalculate(t *testing.T) {
	tests := []struct {
		name         string
		hosts        int
		deadline     time.Time
		today        time.Time
		wantPerNight int
		wantWeekdays int
		wantPassed   bool
	}{
		{
			name:         "normal weekdays",
			hosts:        100,
			deadline:     date(2026, time.April, 10), // Fri
			today:        date(2026, time.April, 6),  // Mon
			wantPerNight: 20,
			wantWeekdays: 5,
		},
		{
			name:         "ceiling division",
			hosts:        101,
			deadline:     date(2026, time.April, 10),
			today:        date(2026, time.April, 6),
			wantPerNight: 21,
			wantWeekdays: 5,
		},
		{
			name:         "spans weekend",
			hosts:        90,
			deadline:     date(2026, time.April, 7), // Tue
			today:        date(2026, time.April, 3), // Fri
			wantPerNight: 30,
			wantWeekdays: 3, // Fri + Mon + Tue
		},
		{
			name:         "today is weekend",
			hosts:        100,
			deadline:     date(2026, time.April, 10), // Fri
			today:        date(2026, time.April, 4),  // Sat
			wantPerNight: 20,
			wantWeekdays: 5, // Mon-Fri
		},
		{
			name:       "deadline passed",
			hosts:      100,
			deadline:   date(2026, time.April, 1),
			today:      date(2026, time.April, 5),
			wantPassed: true,
		},
		{
			name:         "deadline is today weekday",
			hosts:        100,
			deadline:     date(2026, time.April, 6), // Mon
			today:        date(2026, time.April, 6),
			wantPerNight: 100,
			wantWeekdays: 1,
		},
		{
			name:         "deadline is today weekend",
			hosts:        100,
			deadline:     date(2026, time.April, 4), // Sat
			today:        date(2026, time.April, 4),
			wantPerNight: 100, // 0 weekdays, all hosts
			wantWeekdays: 0,
		},
		{
			name:         "large range spanning many weeks",
			hosts:        1000,
			deadline:     date(2026, time.May, 29),  // Fri
			today:        date(2026, time.April, 6), // Mon
			wantPerNight: 25,                        // 40 weekdays, ceil(1000/40)=25
			wantWeekdays: 40,
		},
		{
			name:         "returns total hosts and deadline",
			hosts:        500,
			deadline:     date(2026, time.April, 10),
			today:        date(2026, time.April, 6),
			wantPerNight: 100,
			wantWeekdays: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := Calculate(tt.hosts, tt.deadline, tt.today)

			if result.DeadlinePassed != tt.wantPassed {
				t.Errorf("DeadlinePassed = %v, want %v", result.DeadlinePassed, tt.wantPassed)
			}
			if tt.wantPassed {
				return
			}
			if result.HostsPerNight != tt.wantPerNight {
				t.Errorf("HostsPerNight = %d, want %d", result.HostsPerNight, tt.wantPerNight)
			}
			if result.WeekdaysLeft != tt.wantWeekdays {
				t.Errorf("WeekdaysLeft = %d, want %d", result.WeekdaysLeft, tt.wantWeekdays)
			}
			if result.TotalHosts != tt.hosts {
				t.Errorf("TotalHosts = %d, want %d", result.TotalHosts, tt.hosts)
			}
			if !result.Deadline.Equal(tt.deadline) {
				t.Errorf("Deadline = %v, want %v", result.Deadline, tt.deadline)
			}
		})
	}
}

func TestCountWeekdaysEndBeforeStart(t *testing.T) {
	// Exercises the defensive totalDays <= 0 branch in countWeekdays when
	// end is before start (should never happen via Calculate, but the guard
	// must return 0 rather than a negative count).
	got := countWeekdays(date(2026, time.April, 10), date(2026, time.April, 5))
	if got != 0 {
		t.Errorf("countWeekdays(end < start) = %d, want 0", got)
	}
}

func TestTruncateToDay(t *testing.T) {
	input := time.Date(2026, time.April, 6, 14, 30, 45, 123, time.UTC)
	got := TruncateToDay(input)
	want := time.Date(2026, time.April, 6, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("TruncateToDay() = %v, want %v", got, want)
	}
	if got.Location() != input.Location() {
		t.Errorf("TruncateToDay() location = %v, want %v", got.Location(), input.Location())
	}
}

func TestCommaFormat(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{999, "999"},
		{1000, "1,000"},
		{10000, "10,000"},
		{1000000, "1,000,000"},
		{1234567, "1,234,567"},
		{-1, "-1"},
		{-1000, "-1,000"},
		{-1234567, "-1,234,567"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := CommaFormat(tt.input); got != tt.want {
				t.Errorf("CommaFormat(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
