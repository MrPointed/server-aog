package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var restartGraceful bool

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the server",
	Run: func(cmd *cobra.Command, args []string) {
		if restartGraceful {
			fmt.Println("Restarting server gracefully...")
		} else {
			fmt.Println("Restarting server immediately...")
		}
	},
}

func init() {
	restartCmd.Flags().BoolVarP(&restartGraceful, "graceful", "g", false, "Restart gracefully")
	rootCmd.AddCommand(restartCmd)
}
