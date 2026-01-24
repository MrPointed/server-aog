package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor server in real-time",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting monitor TUI...")
		// Logic for TUI
	},
}

func init() {
	rootCmd.AddCommand(monitorCmd)
}
