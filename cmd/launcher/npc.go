package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var npcCmd = &cobra.Command{
	Use:   "npc",
	Short: "NPC and AI management",
}

var npcReloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload NPCs",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Reloading NPCs...")
	},
}

var npcDisableCmd = &cobra.Command{
	Use:   "disable [group]",
	Short: "Disable AI group",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Disabling AI group %s...\n", args[0])
	},
}

var npcEnableCmd = &cobra.Command{
	Use:   "enable [group]",
	Short: "Enable AI group",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Enabling AI group %s...\n", args[0])
	},
}

var respawnMap string
var npcRespawnCmd = &cobra.Command{
	Use:   "respawn",
	Short: "Respawn NPCs",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Respawning NPCs in map %s...\n", respawnMap)
	},
}

func init() {
	npcRespawnCmd.Flags().StringVarP(&respawnMap, "map", "m", "", "Map ID")

	npcCmd.AddCommand(npcReloadCmd)
	npcCmd.AddCommand(npcDisableCmd)
	npcCmd.AddCommand(npcEnableCmd)
	npcCmd.AddCommand(npcRespawnCmd)
	rootCmd.AddCommand(npcCmd)
}
