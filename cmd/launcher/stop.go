package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var graceful bool

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the server",
	Run: func(cmd *cobra.Command, args []string) {
		if graceful {
			fmt.Println("Stopping server gracefully...")
		} else {
			fmt.Println("Stopping server immediately...")
		}
	},
}

func init() {
	stopCmd.Flags().BoolVarP(&graceful, "graceful", "g", false, "Stop gracefully")
	rootCmd.AddCommand(stopCmd)
}
