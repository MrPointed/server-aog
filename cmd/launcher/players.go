package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

const AdminAPIAddrPlayer = "http://localhost:7667"

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
		resp, err := http.Get(fmt.Sprintf("%s/account/lock?nick=%s", AdminAPIAddrPlayer, args[0]))
		if err != nil {
			fmt.Printf("Error locking account: %v\n", err)
			return
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Println(string(body))
	},
}

var accountUnlockCmd = &cobra.Command{
	Use:   "unlock [id]",
	Short: "Unlock an account",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(fmt.Sprintf("%s/account/unlock?nick=%s", AdminAPIAddrPlayer, args[0]))
		if err != nil {
			fmt.Printf("Error unlocking account: %v\n", err)
			return
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Println(string(body))
	},
}

var accountResetPasswordCmd = &cobra.Command{
	Use:   "reset-password [id]",
	Short: "Reset account password",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(fmt.Sprintf("%s/account/reset-password?nick=%s", AdminAPIAddrPlayer, args[0]))
		if err != nil {
			fmt.Printf("Error resetting password: %v\n", err)
			return
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Println(string(body))
	},
}

var playerTeleportCmd = &cobra.Command{
	Use:   "teleport [id] [map] [x] [y]",
	Short: "Teleport a player",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		url := fmt.Sprintf("%s/player/teleport?nick=%s&map=%s&x=%s&y=%s", AdminAPIAddrPlayer, args[0], args[1], args[2], args[3])
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Error teleporting player: %v\n", err)
			return
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Println(string(body))
	},
}

var saveAll bool
var playerSaveCmd = &cobra.Command{
	Use:   "save [id]",
	Short: "Save player data",
	Run: func(cmd *cobra.Command, args []string) {
		var url string
		if saveAll {
			url = fmt.Sprintf("%s/player/save?all=true", AdminAPIAddrPlayer)
		} else if len(args) > 0 {
			url = fmt.Sprintf("%s/player/save?nick=%s", AdminAPIAddrPlayer, args[0])
		} else {
			fmt.Println("Please specify player ID or --all")
			return
		}

		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Error saving: %v\n", err)
			return
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Println(string(body))
	},
}

var playerInfoCmd = &cobra.Command{
	Use:   "info [nick]",
	Short: "Get online player information",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(fmt.Sprintf("%s/player/info?nick=%s", AdminAPIAddrPlayer, args[0]))
		if err != nil {
			fmt.Printf("Error getting player info: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	},
}

var playerKickCmd = &cobra.Command{
	Use:   "kick [nick]",
	Short: "Kick an online player",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(fmt.Sprintf("%s/conn/kick?name=%s", AdminAPIAddrPlayer, args[0]))
		if err != nil {
			fmt.Printf("Error kicking player: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	},
}

func init() {
	playerSaveCmd.Flags().BoolVarP(&saveAll, "all", "", false, "Save all players")

	accountCmd.AddCommand(accountLockCmd)
	accountCmd.AddCommand(accountUnlockCmd)
	accountCmd.AddCommand(accountResetPasswordCmd)
	
	playerCmd.AddCommand(playerTeleportCmd)
	playerCmd.AddCommand(playerSaveCmd)
	playerCmd.AddCommand(playerInfoCmd)
	playerCmd.AddCommand(playerKickCmd)
	
	rootCmd.AddCommand(accountCmd)
	rootCmd.AddCommand(playerCmd)
}
