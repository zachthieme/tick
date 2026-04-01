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
