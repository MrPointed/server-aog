package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

const AdminAPIAddrNPC = "http://localhost:7667"

var npcCmd = &cobra.Command{
	Use:   "npc",
	Short: "NPC and AI management",
}

var npcReloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload NPCs",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(fmt.Sprintf("%s/npc/reload", AdminAPIAddrNPC))
		if err != nil {
			fmt.Printf("Error reloading NPCs: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	},
}

var disableAll bool
var npcDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable AI",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(fmt.Sprintf("%s/npc/disable", AdminAPIAddrNPC))
		if err != nil {
			fmt.Printf("Error disabling AI: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	},
}

var enableAll bool
var npcEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable AI",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(fmt.Sprintf("%s/npc/enable", AdminAPIAddrNPC))
		if err != nil {
			fmt.Printf("Error enabling AI: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	},
}

var respawnMap string
var npcRespawnCmd = &cobra.Command{
	Use:   "respawn",
	Short: "Respawn NPCs",
	Run: func(cmd *cobra.Command, args []string) {
		if respawnMap == "" {
			fmt.Println("Please specify map ID with --map or -m")
			return
		}
		resp, err := http.Get(fmt.Sprintf("%s/npc/respawn?map=%s", AdminAPIAddrNPC, respawnMap))
		if err != nil {
			fmt.Printf("Error respawning NPCs: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	},
}

var npcListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active NPCs",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(fmt.Sprintf("%s/npc/list", AdminAPIAddrNPC))
		if err != nil {
			fmt.Printf("Error listing NPCs: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	},
}

func init() {
	npcRespawnCmd.Flags().StringVarP(&respawnMap, "map", "m", "", "Map ID")

	npcCmd.AddCommand(npcReloadCmd)
	npcCmd.AddCommand(npcDisableCmd)
	npcCmd.AddCommand(npcEnableCmd)
	npcCmd.AddCommand(npcRespawnCmd)
	npcCmd.AddCommand(npcListCmd)
	rootCmd.AddCommand(npcCmd)
}
