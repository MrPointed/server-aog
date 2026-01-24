package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var env string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Starting server in %s environment...\n", env)
		// Here you would typically call the server start logic
		// For now, it's a placeholder to show the structure
	},
}

func init() {
	startCmd.Flags().StringVarP(&env, "env", "e", "dev", "Environment (dev, prod)")
	rootCmd.AddCommand(startCmd)
}

