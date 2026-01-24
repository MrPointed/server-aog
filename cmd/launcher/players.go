package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var playerCmd = &cobra.Command{
	Use:   "player",
	Short: "Player management",
}

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Account management",
}

var accountLockCmd = &cobra.Command{
	Use:   "lock [id]",
	Short: "Lock an account",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Locking account %s...\n", args[0])
	},
}

var accountUnlockCmd = &cobra.Command{
	Use:   "unlock [id]",
	Short: "Unlock an account",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Unlocking account %s...\n", args[0])
	},
}

var accountResetPasswordCmd = &cobra.Command{
	Use:   "reset-password [id]",
	Short: "Reset account password",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Resetting password for account %s...\n", args[0])
	},
}

var playerTeleportCmd = &cobra.Command{
	Use:   "teleport [id] [map] [x] [y]",
	Short: "Teleport a player",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Teleporting player %s to map %s (%s, %s)...\n", args[0], args[1], args[2], args[3])
	},
}

var saveAll bool
var playerSaveCmd = &cobra.Command{
	Use:   "save [id]",
	Short: "Save player data",
	Run: func(cmd *cobra.Command, args []string) {
		if saveAll {
			fmt.Println("Saving all players...")
		} else if len(args) > 0 {
			fmt.Printf("Saving player %s...\n", args[0])
		} else {
			fmt.Println("Please specify player ID or --all")
		}
	},
}

func init() {
	playerSaveCmd.Flags().BoolVarP(&saveAll, "all", "", false, "Save all players")

	accountCmd.AddCommand(accountLockCmd)
	accountCmd.AddCommand(accountUnlockCmd)
	accountCmd.AddCommand(accountResetPasswordCmd)
	
	playerCmd.AddCommand(playerTeleportCmd)
	playerCmd.AddCommand(playerSaveCmd)
	
	rootCmd.AddCommand(accountCmd)
	rootCmd.AddCommand(playerCmd)
}
