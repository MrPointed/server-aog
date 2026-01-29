package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

const AdminAPIAddr = "http://localhost:7667"

var worldCmd = &cobra.Command{
	Use:   "world",
	Short: "World and maps management",
}

var worldLoadCmd = &cobra.Command{
	Use:   "load [map...]",
	Short: "Load one or more maps",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, m := range args {
			resp, err := http.Get(fmt.Sprintf("%s/world/load?id=%s", AdminAPIAddr, m))
			if err != nil {
				fmt.Printf("Error loading map %s: %v\n", m, err)
				continue
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Println(string(body))
		}
	},
}

var worldUnloadCmd = &cobra.Command{
	Use:   "unload [map...]",
	Short: "Unload one or more maps",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, m := range args {
			resp, err := http.Get(fmt.Sprintf("%s/world/unload?id=%s", AdminAPIAddr, m))
			if err != nil {
				fmt.Printf("Error unloading map %s: %v\n", m, err)
				continue
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Println(string(body))
		}
	},
}

var worldReloadCmd = &cobra.Command{
	Use:   "reload [map...]",
	Short: "Reload one or more maps",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, m := range args {
			resp, err := http.Get(fmt.Sprintf("%s/world/reload?id=%s", AdminAPIAddr, m))
			if err != nil {
				fmt.Printf("Error reloading map %s: %v\n", m, err)
				continue
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Println(string(body))
		}
	},
}

var worldResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Save world and reload all maps",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(fmt.Sprintf("%s/world/reset", AdminAPIAddr))
		if err != nil {
			fmt.Printf("Error resetting world: %v\n", err)
			return
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Println(string(body))
	},
}

var worldListCmd = &cobra.Command{
	Use:   "list",
	Short: "List loaded maps",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(fmt.Sprintf("%s/world/list", AdminAPIAddr))
		if err != nil {
			fmt.Printf("Error listing maps: %v\n", err)
			return
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("Loaded maps: %s\n", string(body))
	},
}

var worldSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save world maps to cache",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(fmt.Sprintf("%s/world/save", AdminAPIAddr))
		if err != nil {
			fmt.Printf("Error saving world: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	},
}

func init() {
	worldCmd.AddCommand(worldLoadCmd)
	worldCmd.AddCommand(worldUnloadCmd)
	worldCmd.AddCommand(worldReloadCmd)
	worldCmd.AddCommand(worldResetCmd)
	worldCmd.AddCommand(worldSaveCmd)
	rootCmd.AddCommand(worldCmd)
}
