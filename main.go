// main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"tick/calc"
	"tick/tui"

	tea "github.com/charmbracelet/bubbletea"
)

// version is set at build time via ldflags.
var version = "dev"

const usage = `Usage: tick --hosts N --deadline YYYY-MM-DD [--today YYYY-MM-DD] [--once] [--json]

A terminal dashboard showing how many hosts to upgrade per weeknight.

Flags:
  --hosts       Total number of hosts to upgrade (required unless --hosts-file)
  --hosts-file  Path to a file containing the host count (re-read each tick in TUI mode)
  --deadline    Target completion date in YYYY-MM-DD format (required)
  --today       Override today's date (YYYY-MM-DD, defaults to system clock)
  --once        Print a one-liner and exit
  --json        Output JSON and exit (for scripting)
  --version     Print version and exit

Examples:
  tick --hosts 500 --deadline 2026-04-30
  tick --hosts-file /tmp/remaining.txt --deadline 2026-04-30
  tick --hosts 500 --deadline 2026-04-30 --once
  tick --hosts 500 --deadline 2026-04-30 --json
  tick --hosts 500 --deadline 2026-04-30 --today 2026-04-10`

type jsonResult struct {
	HostsPerNight  int    `json:"hosts_per_night"`
	WeekdaysLeft   int    `json:"weekdays_left"`
	TotalHosts     int    `json:"total_hosts"`
	Deadline       string `json:"deadline"`
	Today          string `json:"today"`
	DeadlinePassed bool   `json:"deadline_passed"`
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// readHostsFile reads a positive integer from the first line of the file at path.
func readHostsFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(data))
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid host count %q: %w", s, err)
	}
	if n <= 0 {
		return 0, fmt.Errorf("host count must be positive, got %d", n)
	}
	return n, nil
}

func run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("tick", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { _, _ = fmt.Fprintln(stderr, usage) }

	hosts := fs.Int("hosts", 0, "total number of hosts to upgrade")
	hostsFile := fs.String("hosts-file", "", "path to a file containing the host count")
	deadlineStr := fs.String("deadline", "", "target date in YYYY-MM-DD format (required)")
	todayStr := fs.String("today", "", "override today's date (YYYY-MM-DD, defaults to system clock)")
	once := fs.Bool("once", false, "print a one-liner and exit")
	jsonOut := fs.Bool("json", false, "output JSON and exit (for scripting)")
	showVersion := fs.Bool("version", false, "print version and exit")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *showVersion {
		_, err := fmt.Fprintf(stdout, "tick %s\n", version)
		return err
	}

	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument: %s", fs.Arg(0))
	}

	hostsExplicit := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "hosts" {
			hostsExplicit = true
		}
	})

	if hostsExplicit && *hostsFile != "" {
		return fmt.Errorf("--hosts and --hosts-file are mutually exclusive")
	}
	if *hostsFile != "" {
		n, err := readHostsFile(*hostsFile)
		if err != nil {
			return fmt.Errorf("--hosts-file: %w", err)
		}
		*hosts = n
	}
	if *hosts < 0 || (*hosts == 0 && hostsExplicit) {
		return fmt.Errorf("--hosts: value must be positive, got %d", *hosts)
	}
	if *hosts == 0 {
		return fmt.Errorf("--hosts is required (or use --hosts-file)")
	}
	if *deadlineStr == "" {
		return fmt.Errorf("--deadline is required (YYYY-MM-DD)")
	}

	deadline, err := time.ParseInLocation("2006-01-02", *deadlineStr, time.Local)
	if err != nil {
		return fmt.Errorf("invalid deadline format: %s (expected YYYY-MM-DD)", *deadlineStr)
	}

	today := calc.DateOnly(time.Now())
	if *todayStr != "" {
		today, err = time.ParseInLocation("2006-01-02", *todayStr, time.Local)
		if err != nil {
			return fmt.Errorf("invalid today format: %s (expected YYYY-MM-DD)", *todayStr)
		}
	}

	if *jsonOut {
		result := calc.Calculate(*hosts, deadline, today)
		out := jsonResult{
			HostsPerNight:  result.HostsPerNight,
			WeekdaysLeft:   result.WeekdaysLeft,
			TotalHosts:     result.TotalHosts,
			Deadline:       result.Deadline.Format("2006-01-02"),
			Today:          today.Format("2006-01-02"),
			DeadlinePassed: result.DeadlinePassed,
		}
		return json.NewEncoder(stdout).Encode(out)
	}

	if *once {
		result := calc.Calculate(*hosts, deadline, today)
		if result.DeadlinePassed {
			_, err = fmt.Fprintf(stdout, "Deadline passed — %s hosts remaining (deadline was %s)\n",
				calc.CommaFormat(result.TotalHosts), result.Deadline.Format("2006-01-02"))
			return err
		}
		_, err = fmt.Fprintf(stdout, "%d weekdays remaining — %s hosts/night (%s hosts, deadline %s)\n",
			result.WeekdaysLeft, calc.CommaFormat(result.HostsPerNight), calc.CommaFormat(result.TotalHosts),
			result.Deadline.Format("2006-01-02"))
		return err
	}

	var readHosts func() (int, error)
	if *hostsFile != "" {
		path := *hostsFile
		readHosts = func() (int, error) { return readHostsFile(path) }
	}

	m := tui.New(*hosts, readHosts, deadline, today, *todayStr != "")
	p := tea.NewProgram(m, tea.WithAltScreen())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-sigCh
		p.Quit()
	}()

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error: %v", err)
	}
	signal.Stop(sigCh)
	return nil
}
