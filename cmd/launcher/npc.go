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

var disableAll bool
var npcDisableCmd = &cobra.Command{
	Use:   "disable [group]",
	Short: "Disable AI group",
	Run: func(cmd *cobra.Command, args []string) {
		if disableAll {
			fmt.Println("Disabling all AI groups...")
		} else if len(args) > 0 {
			fmt.Printf("Disabling AI group %s...\n", args[0])
		} else {
			fmt.Println("Please specify group or --all")
		}
	},
}

var enableAll bool
var npcEnableCmd = &cobra.Command{
	Use:   "enable [group]",
	Short: "Enable AI group",
	Run: func(cmd *cobra.Command, args []string) {
		if enableAll {
			fmt.Println("Enabling all AI groups...")
		} else if len(args) > 0 {
			fmt.Printf("Enabling AI group %s...\n", args[0])
		} else {
			fmt.Println("Please specify group or --all")
		}
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
	npcDisableCmd.Flags().BoolVarP(&disableAll, "all", "", false, "Disable all AI")
	npcEnableCmd.Flags().BoolVarP(&enableAll, "all", "", false, "Enable all AI")
	npcRespawnCmd.Flags().StringVarP(&respawnMap, "map", "m", "", "Map ID")

	npcCmd.AddCommand(npcReloadCmd)
	npcCmd.AddCommand(npcDisableCmd)
	npcCmd.AddCommand(npcEnableCmd)
	npcCmd.AddCommand(npcRespawnCmd)
	rootCmd.AddCommand(npcCmd)
}
