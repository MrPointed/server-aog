package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

const AdminAPIAddrConn = "http://localhost:7667"

var connCmd = &cobra.Command{
	Use:   "conn",
	Short: "Connection management",
}

var connListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active connections",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(fmt.Sprintf("%s/conn/list", AdminAPIAddrConn))
		if err != nil {
			fmt.Printf("Error listing connections: %v\n", err)
			return
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("Active connections: %s\n", string(body))
	},
}

var connCountCmd = &cobra.Command{
	Use:   "count",
	Short: "Count active connections",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(fmt.Sprintf("%s/conn/count", AdminAPIAddrConn))
		if err != nil {
			fmt.Printf("Error counting connections: %v\n", err)
			return
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("Active connections: %s\n", string(body))
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
		var url string
		if kickAccount != "" {
			url = fmt.Sprintf("%s/conn/kick?name=%s", AdminAPIAddrConn, kickAccount)
		} else if kickIP != "" {
			url = fmt.Sprintf("%s/conn/kick?ip=%s", AdminAPIAddrConn, kickIP)
		} else if kickInactive {
			fmt.Println("Kick inactive not yet implemented in API")
			return
		} else {
			fmt.Println("Please specify --account, --ip or --inactive")
			return
		}

		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Error kicking: %v\n", err)
			return
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Println(string(body))
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
