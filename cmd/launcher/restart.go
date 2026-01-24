package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var restartGraceful bool
var restartRolling bool

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the server",
	Run: func(cmd *cobra.Command, args []string) {
		if restartRolling {
			fmt.Println("Performing rolling restart...")
		} else if restartGraceful {
			fmt.Println("Restarting server gracefully...")
		} else {
			fmt.Println("Restarting server immediately...")
		}
	},
}

func init() {
	restartCmd.Flags().BoolVarP(&restartGraceful, "graceful", "g", false, "Restart gracefully")
	restartCmd.Flags().BoolVarP(&restartRolling, "rolling", "r", false, "Rolling restart")
	rootCmd.AddCommand(restartCmd)
}
