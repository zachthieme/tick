# Countdown Timer — Host Upgrade Scheduler

## Purpose

A terminal app that calculates how many hosts need to be upgraded per weeknight to meet a deadline. Provides an at-a-glance dashboard for ops teams tracking nightly upgrade quotas.

## CLI Interface

Single binary with three flags:

- `--hosts` (required) — total number of hosts to upgrade
- `--deadline` (required) — target date in `YYYY-MM-DD` format
- `--once` (optional) — print a one-liner and exit

Today's date comes from the system clock.

```
countdown --hosts 500 --deadline 2026-04-30
countdown --hosts 500 --deadline 2026-04-30 --once
```

## Core Calculation

Pure function: `func Calculate(totalHosts int, deadline time.Time, today time.Time) Result`

- Count remaining weekdays (Mon–Fri) from today through deadline, inclusive of both
- `hostsPerNight = ceil(totalHosts / remainingWeekdays)`
- Returns: `hostsPerNight`, `remainingWeekdays`, `totalHosts`, `deadline`

Edge cases:
- Today is past the deadline: show "deadline passed"
- Today is the deadline: all remaining hosts tonight
- Today is a weekend: show the count, noting next work night is Monday

## TUI Mode (default)

Full-screen Bubble Tea app. Everything centered vertically and horizontally.

Layout (top to bottom, no gaps between lines):

```
        ██████  ██████
             █       █
         █████   █████
        █            █
        ███████ ██████
        hosts per night
          500 hosts left
      Deadline: 2026-04-30
```

- Big number rendered via charmbracelet big-text library, colored (bold cyan or green)
- "hosts per night" label centered below big number
- "X hosts left" centered below label
- "Deadline: YYYY-MM-DD" centered below that
- No gaps between the label and detail lines
- Recalculates on a tick (once per minute) to catch day rollover
- Quit with `q` or `ctrl+c`

Styling: Lip Gloss for colors. Big number in a bold/bright color, detail lines in a subdued color.

## One-Shot Mode (`--once`)

Prints a single line and exits:

```
22 weekdays remaining — 23 hosts/night (500 hosts, deadline 2026-04-30)
```

## Dependencies

- `charmbracelet/bubbletea` — TUI framework
- `charmbracelet/lipgloss` — styling
- `charmbracelet/x/exp/term/ansi/figlet` or equivalent charmbracelet big-text library — large digit rendering

## Project Structure

```
countdown/
├── main.go          # flag parsing, mode dispatch
├── calc/
│   └── calc.go      # pure calculation logic
└── tui/
    └── tui.go       # Bubble Tea model/update/view
```
