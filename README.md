# tick

A terminal dashboard that shows how many hosts you need to upgrade per night to hit your deadline.

Counts weekdays (Mon-Fri) between now and your deadline, then displays the nightly quota as a big number you can glance at from across the room.

## Install

```bash
go install .
```

## Usage

```bash
# Fullscreen TUI dashboard
tick --hosts 500 --deadline 2026-04-30

# One-liner output
tick --hosts 500 --deadline 2026-04-30 --once
# 21 weekdays remaining — 24 hosts/night (500 hosts, deadline 2026-04-30)

# Override today's date (adds a "Start:" line to the TUI)
tick --hosts 500 --deadline 2026-04-30 --today 2026-04-10
```

## Flags

| Flag | Required | Description |
|------|----------|-------------|
| `--hosts` | Yes | Total number of hosts to upgrade |
| `--deadline` | Yes | Target date (`YYYY-MM-DD`) |
| `--today` | No | Override today's date (`YYYY-MM-DD`) |
| `--once` | No | Print a single line and exit |

## TUI

The dashboard shows a large number for hosts-per-night, with the total hosts remaining and deadline underneath. It recalculates automatically when the date rolls over. Press `q` or `ctrl+c` to exit.

```
 .d8888b.      d8888
d88P  Y88b    d8P888
       888   d8P 888
     .d88P  d8P  888
 .od888P"  d88   888
d88P"      8888888888
888"             888
888888888        888
    hosts per night
    500 hosts left
 Deadline: 2026-04-30
```
