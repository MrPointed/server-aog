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
		if kickAccount == "" {
			fmt.Println("Please specify account with --account or -a")
			return
		}
		resp, err := http.Get(fmt.Sprintf("%s/conn/ban?nick=%s", AdminAPIAddrConn, kickAccount))
		if err != nil {
			fmt.Printf("Error banning: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	},
}

var connUnbanCmd = &cobra.Command{
	Use:   "unban",
	Short: "Unban an account",
	Run: func(cmd *cobra.Command, args []string) {
		if kickAccount == "" {
			fmt.Println("Please specify account with --account or -a")
			return
		}
		resp, err := http.Get(fmt.Sprintf("%s/conn/unban?nick=%s", AdminAPIAddrConn, kickAccount))
		if err != nil {
			fmt.Printf("Error unbanning: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	},
}

func init() {
	connKickCmd.Flags().StringVarP(&kickAccount, "account", "a", "", "Account ID")
	connKickCmd.Flags().StringVarP(&kickIP, "ip", "i", "", "IP address")
	connKickCmd.Flags().BoolVarP(&kickInactive, "inactive", "", false, "Kick inactive connections")
	connKickCmd.Flags().StringVarP(&kickSince, "since", "", "30m", "Inactivity duration")
	
	connBanCmd.Flags().StringVarP(&kickAccount, "account", "a", "", "Account ID")
	connBanCmd.Flags().StringVarP(&banDuration, "duration", "d", "24h", "Ban duration")

	connUnbanCmd.Flags().StringVarP(&kickAccount, "account", "a", "", "Account ID")

	connCmd.AddCommand(connListCmd)
	connCmd.AddCommand(connCountCmd)
	connCmd.AddCommand(connKickCmd)
	connCmd.AddCommand(connBanCmd)
	connCmd.AddCommand(connUnbanCmd)
	rootCmd.AddCommand(connCmd)
}
