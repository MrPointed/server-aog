package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage server configuration",
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		fmt.Printf("Config %s: <value>\n", key)
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key=value] [flags...]",
	Short: "Set configuration values",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, arg := range args {
			if strings.Contains(arg, "=") {
				parts := strings.Split(arg, "=")
				fmt.Printf("Setting %s to %s\n", parts[0], parts[1])
			} else {
				fmt.Printf("Applying flag/option: %s\n", arg)
			}
		}
	},
}

var configReloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload configuration from file",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Reloading configuration...")
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configReloadCmd)
	rootCmd.AddCommand(configCmd)
}
