package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var graceful bool

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the server",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := os.ReadFile("server.pid")
		if err != nil {
			fmt.Println("Server is not running (PID file not found).")
			return
		}

		pid, err := strconv.Atoi(string(data))
		if err != nil {
			fmt.Println("Invalid PID file.")
			return
		}

		handleStop(pid, graceful)
	},
}

func init() {
	stopCmd.Flags().BoolVarP(&graceful, "graceful", "g", false, "Stop gracefully")
	rootCmd.AddCommand(stopCmd)
}