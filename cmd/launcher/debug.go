package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Internal debug commands",
}

var debugHeapCmd = &cobra.Command{
	Use:   "heap",
	Short: "Dump heap profile",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Dumping heap profile...")
	},
}

var debugGoroutinesCmd = &cobra.Command{
	Use:   "goroutines",
	Short: "List all goroutines",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing goroutines...")
	},
}

var debugDeadlocksCmd = &cobra.Command{
	Use:   "deadlocks",
	Short: "Check for deadlocks",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking for deadlocks...")
	},
}

var debugDumpStateCmd = &cobra.Command{
	Use:   "dump-state",
	Short: "Dump full server state",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Dumping server state...")
	},
}

func init() {
	debugCmd.AddCommand(debugHeapCmd)
	debugCmd.AddCommand(debugGoroutinesCmd)
	debugCmd.AddCommand(debugDeadlocksCmd)
	debugCmd.AddCommand(debugDumpStateCmd)
	rootCmd.AddCommand(debugCmd)
}
