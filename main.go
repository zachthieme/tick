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
