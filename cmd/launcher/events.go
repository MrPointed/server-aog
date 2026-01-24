package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var eventCmd = &cobra.Command{
	Use:   "event",
	Short: "Game events management",
}

var eventStartCmd = &cobra.Command{
	Use:   "start [name]",
	Short: "Start an event",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Starting event %s...\n", args[0])
	},
}

var eventStopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop an event",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Stopping event %s...\n", args[0])
	},
}

var eventStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check active events",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("No active events.")
	},
}

func init() {
	eventCmd.AddCommand(eventStartCmd)
	eventCmd.AddCommand(eventStopCmd)
	eventCmd.AddCommand(eventStatusCmd)
	rootCmd.AddCommand(eventCmd)
}

