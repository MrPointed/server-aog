package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check server status",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking server status...")
		// Logic to check if server is running
	},
}

var uptimeCmd = &cobra.Command{
	Use:   "uptime",
	Short: "Check server uptime",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Server uptime: 0h 0m 0s")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(uptimeCmd)
}
