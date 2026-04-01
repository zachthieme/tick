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

func commaFormat(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}

func main() {
	hosts := flag.Int("hosts", 0, "total number of hosts to upgrade (required)")
	deadlineStr := flag.String("deadline", "", "target date in YYYY-MM-DD format (required)")
	todayStr := flag.String("today", "", "override today's date (YYYY-MM-DD, defaults to system clock)")
	once := flag.Bool("once", false, "print a one-liner and exit")
	flag.Parse()

	if *hosts <= 0 || *deadlineStr == "" {
		fmt.Fprintln(os.Stderr, "Usage: countdown --hosts N --deadline YYYY-MM-DD [--today YYYY-MM-DD] [--once]")
		os.Exit(1)
	}

	deadline, err := time.Parse("2006-01-02", *deadlineStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid deadline format: %s (expected YYYY-MM-DD)\n", *deadlineStr)
		os.Exit(1)
	}

	today := time.Now()
	if *todayStr != "" {
		today, err = time.Parse("2006-01-02", *todayStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid today format: %s (expected YYYY-MM-DD)\n", *todayStr)
			os.Exit(1)
		}
	}

	if *once {
		result := calc.Calculate(*hosts, deadline, today)
		if result.DeadlinePassed {
			fmt.Printf("Deadline passed — %s hosts remaining (deadline was %s)\n",
				commaFormat(result.TotalHosts), result.Deadline.Format("2006-01-02"))
		} else {
			fmt.Printf("%d weekdays remaining — %s hosts/night (%s hosts, deadline %s)\n",
				result.WeekdaysLeft, commaFormat(result.HostsPerNight), commaFormat(result.TotalHosts),
				result.Deadline.Format("2006-01-02"))
		}
		return
	}

	m := tui.New(*hosts, deadline, today)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
