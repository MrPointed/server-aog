//go:build !windows
package main

import (
	"fmt"
	"os"
	"syscall"
)

func handleStop(pid int, graceful bool) {
	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("Failed to find process %d: %v\n", pid, err)
		return
	}

	if graceful {
		fmt.Printf("Stopping server (PID %d) gracefully...\n", pid)
		process.Signal(syscall.SIGTERM)
	} else {
		fmt.Printf("Stopping server (PID %d) immediately...\n", pid)
		process.Kill()
	}
	
os.Remove("server.pid")
}
