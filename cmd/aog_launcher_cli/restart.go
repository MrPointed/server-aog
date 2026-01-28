package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ao-go-server/internal/server"
	"github.com/spf13/cobra"
)

var restartGraceful bool
var restartRolling bool

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the server",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := os.ReadFile("server.pid")
		if err == nil {
			pid, err := strconv.Atoi(string(data))
			if err == nil {
				fmt.Printf("Stopping existing server (PID %d)...\n", pid)
				handleStop(pid, restartGraceful)

				// Wait for process to exit
				for i := 0; i < 30; i++ { // Wait up to 3 seconds
					process, err := os.FindProcess(pid)
					if err != nil {
						break
					}
					// On Unix, Signal(0) checks if process exists.
					// This is a simple cross-platform way to check if it's gone
					// although FindProcess always returns something on Unix.
					if isProcessDead(process) {
						break
					}
					time.Sleep(100 * time.Millisecond)
				}
			}
		}

		if restartRolling {
			fmt.Println("Rolling restart not fully implemented, performing normal restart...")
		}

		fmt.Printf("Restarting server in %s environment on port %s...\n", env, port)

		resourcesPath := "resources"
		absPath, _ := filepath.Abs(resourcesPath)
		fmt.Printf("Using resources from: %s\n", absPath)

		s := server.NewServer(":"+port, resourcesPath)
		if err := s.Start(); err != nil {
			fmt.Printf("Server failed: %v\n", err)
		}
	},
}

// isProcessDead checks if the process is no longer running.
func isProcessDead(p *os.Process) bool {
	// On Unix-like systems, p.Signal(0) is the standard way to check for existence.
	// FindProcess on Unix always succeeds, so we must check if we can send a signal.
	err := p.Signal(os.Signal(nil))
	return err != nil
}

func init() {
	restartCmd.Flags().BoolVarP(&restartGraceful, "graceful", "g", false, "Restart gracefully")
	restartCmd.Flags().BoolVarP(&restartRolling, "rolling", "r", false, "Rolling restart")
	
	// Reuse env and port variables from start.go
	restartCmd.Flags().StringVarP(&env, "env", "e", "dev", "Environment (dev, prod)")
	restartCmd.Flags().StringVarP(&port, "port", "p", "7666", "Port to listen on")
	
	rootCmd.AddCommand(restartCmd)
}
