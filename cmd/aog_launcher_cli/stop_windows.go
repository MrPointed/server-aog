//go:build windows
package main

import (
	"fmt"
	"os"
)

func handleStop(pid int, graceful bool) {
	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("Failed to find process %d: %v\n", pid, err)
		return
	}

	if graceful {
		// En Windows, Process.Signal no soporta señales como SIGTERM.
		// La forma estándar es enviar un evento de consola o simplemente Kill.
		fmt.Printf("Graceful stop not fully supported on Windows CLI. Killing process %d...\n", pid)
	} else {
		fmt.Printf("Stopping server (PID %d) immediately...\n", pid)
	}
	
	err = process.Kill()
	if err != nil {
		fmt.Printf("Error killing process: %v\n", err)
	}
	
os.Remove("server.pid")
}
