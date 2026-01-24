package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "launcher",
	Short: "Argentum Online Go Server Launcher",
	Long:  `A powerful CLI to manage the Argentum Online Go Server lifecycle, maps, connections, and more.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be defined here
}
