package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var connCmd = &cobra.Command{
	Use:   "conn",
	Short: "Connection management",
}

var connListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active connections",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing active connections...")
	},
}

var connCountCmd = &cobra.Command{
	Use:   "count",
	Short: "Count active connections",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Active connections: 0")
	},
}

var kickAccount string
var kickIP string
var kickInactive bool
var kickSince string
var connKickCmd = &cobra.Command{
	Use:   "kick",
	Short: "Kick a connection",
	Run: func(cmd *cobra.Command, args []string) {
		if kickInactive {
			fmt.Printf("Kicking inactive connections since %s...\n", kickSince)
		} else if kickAccount != "" {
			fmt.Printf("Kicking account %s...\n", kickAccount)
		} else if kickIP != "" {
			fmt.Printf("Kicking IP %s...\n", kickIP)
		} else {
			fmt.Println("Please specify --account, --ip or --inactive")
		}
	},
}

var banDuration string
var connBanCmd = &cobra.Command{
	Use:   "ban",
	Short: "Ban an account",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Banning account %s for %s...\n", kickAccount, banDuration)
	},
}

func init() {
	connKickCmd.Flags().StringVarP(&kickAccount, "account", "a", "", "Account ID")
	connKickCmd.Flags().StringVarP(&kickIP, "ip", "i", "", "IP address")
	connKickCmd.Flags().BoolVarP(&kickInactive, "inactive", "", false, "Kick inactive connections")
	connKickCmd.Flags().StringVarP(&kickSince, "since", "", "30m", "Inactivity duration")
	
	connBanCmd.Flags().StringVarP(&kickAccount, "account", "a", "", "Account ID")
	connBanCmd.Flags().StringVarP(&banDuration, "duration", "d", "24h", "Ban duration")

	connCmd.AddCommand(connListCmd)
	connCmd.AddCommand(connCountCmd)
	connCmd.AddCommand(connKickCmd)
	connCmd.AddCommand(connBanCmd)
	rootCmd.AddCommand(connCmd)
}
