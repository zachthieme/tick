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
