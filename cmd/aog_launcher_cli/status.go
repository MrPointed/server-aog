package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check server status",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := os.ReadFile("server.pid")
		if err != nil {
			fmt.Println("Status: Offline")
			return
		}

		fmt.Printf("Status: Online (PID %s)\n", string(data))
	},
}

var uptimeCmd = &cobra.Command{
	Use:   "uptime",
	Short: "Check server uptime",
	Run: func(cmd *cobra.Command, args []string) {
		info, err := os.Stat("server.pid")
		if err != nil {
			fmt.Println("Server is not running.")
			return
		}
		
		duration := time.Since(info.ModTime())
		fmt.Printf("Server uptime: %s\n", duration.Round(time.Second))
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(uptimeCmd)
}
