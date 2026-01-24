package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

const AdminAPIAddrConfig = "http://localhost:7667"

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
		resp, err := http.Get(fmt.Sprintf("%s/config/get?key=%s", AdminAPIAddrConfig, key))
		if err != nil {
			fmt.Printf("Error getting config: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Config %s: %s\n", key, string(body))
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key=value]",
	Short: "Set configuration values",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, arg := range args {
			if strings.Contains(arg, "=") {
				parts := strings.Split(arg, "=")
				key := parts[0]
				val := parts[1]
				resp, err := http.Get(fmt.Sprintf("%s/config/set?key=%s&value=%s", AdminAPIAddrConfig, key, val))
				if err != nil {
					fmt.Printf("Error setting %s: %v\n", key, err)
					continue
				}
				defer resp.Body.Close()
				body, _ := io.ReadAll(resp.Body)
				fmt.Println(string(body))
			} else {
				fmt.Printf("Invalid format for %s, use key=value\n", arg)
			}
		}
	},
}

var configReloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload configuration (Not fully implemented in API)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Reloading configuration via API not implemented yet.")
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available configuration keys",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(fmt.Sprintf("%s/config/list", AdminAPIAddrConfig))
		if err != nil {
			fmt.Printf("Error listing config keys: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Available configuration keys:")
		fmt.Println(string(body))
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configReloadCmd)
	configCmd.AddCommand(configListCmd)
	rootCmd.AddCommand(configCmd)
}
