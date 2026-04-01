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
