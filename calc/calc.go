// Package calc provides weekday-aware countdown arithmetic.
package calc

import (
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Result holds the output of a countdown calculation.
type Result struct {
	HostsPerNight  int
	WeekdaysLeft   int
	TotalHosts     int
	Deadline       time.Time
	DeadlinePassed bool
}

// Calculate computes how many hosts to upgrade per weeknight.
// Both today and deadline are inclusive: if today == deadline and it is a
// weekday, WeekdaysLeft will be 1. Uses ceiling division so the nightly
// count never underestimates.
func Calculate(totalHosts int, deadline time.Time, today time.Time) Result {
	r := Result{
		TotalHosts: totalHosts,
		Deadline:   deadline,
	}

	if today.After(deadline) {
		r.DeadlinePassed = true
		return r
	}

	weekdays := countWeekdays(today, deadline)
	r.WeekdaysLeft = weekdays

	if weekdays > 0 {
		r.HostsPerNight = (totalHosts + weekdays - 1) / weekdays
	} else {
		r.HostsPerNight = totalHosts
	}

	return r
}

// countWeekdays returns the number of weekdays (Mon-Fri) from start
// through end, inclusive of both. Runs in O(1) time.
//
// NOTE: This hardcodes a Mon-Fri work schedule. If a different schedule
// is needed (e.g. 6-day weeks), this function's weekday check must be
// parameterised with a workday predicate.
func countWeekdays(start, end time.Time) int {
	// Use UTC midnight for day arithmetic to avoid DST issues.
	s := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	e := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC)

	totalDays := int(e.Sub(s)/(24*time.Hour)) + 1
	if totalDays <= 0 {
		return 0
	}

	fullWeeks := totalDays / 7
	remainder := totalDays % 7
	weekdays := fullWeeks * 5

	// Weekday() returns the same value regardless of timezone for date-only
	// values, so using the original start (not UTC-normalized s) is safe here.
	wd := start.Weekday()
	for i := range remainder {
		d := (wd + time.Weekday(i)) % 7
		if d != time.Sunday && d != time.Saturday {
			weekdays++
		}
	}

	return weekdays
}

// TruncateToDay returns a copy of t with the clock component zeroed out,
// preserving the location. Use this instead of time.DateOnly, which is a
// format-string constant, not a time.Time constructor.
func TruncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// commaPrinter is safe for concurrent use — message.Printer is stateless
// after construction and Sprintf does not mutate the receiver.
var commaPrinter = message.NewPrinter(language.English)

// CommaFormat formats an integer with comma separators.
func CommaFormat(n int) string {
	return commaPrinter.Sprintf("%d", n)
}
