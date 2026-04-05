//go:build !windows

package main

import (
	"os"
	"syscall"
)

func quitSignals() []os.Signal {
	return []os.Signal{syscall.SIGTERM, syscall.SIGHUP}
}
