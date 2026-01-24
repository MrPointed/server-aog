package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var worldCmd = &cobra.Command{
	Use:   "world",
	Short: "World and maps management",
}

var worldLoadCmd = &cobra.Command{
	Use:   "load [map]",
	Short: "Load a map",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Loading map %s...\n", args[0])
	},
}

var worldUnloadCmd = &cobra.Command{
	Use:   "unload [map]",
	Short: "Unload a map",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Unloading map %s...\n", args[0])
	},
}

var worldReloadCmd = &cobra.Command{
	Use:   "reload [map]",
	Short: "Reload a map",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Reloading map %s...\n", args[0])
	},
}

var worldListCmd = &cobra.Command{
	Use:   "list",
	Short: "List loaded maps",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing loaded maps...")
	},
}

func init() {
	worldCmd.AddCommand(worldLoadCmd)
	worldCmd.AddCommand(worldUnloadCmd)
	worldCmd.AddCommand(worldReloadCmd)
	worldCmd.AddCommand(worldListCmd)
	rootCmd.AddCommand(worldCmd)
}
