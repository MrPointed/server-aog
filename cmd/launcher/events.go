package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

const AdminAPIAddrEvent = "http://localhost:7667"

var eventCmd = &cobra.Command{
	Use:   "event",
	Short: "Server events management",
}

var eventStartCmd = &cobra.Command{
	Use:   "start [name]",
	Short: "Start a server event",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		resp, err := http.Get(fmt.Sprintf("%s/event/start?name=%s", AdminAPIAddrEvent, name))
		if err != nil {
			fmt.Printf("Error starting event: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	},
}

var eventStopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop a server event",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		resp, err := http.Get(fmt.Sprintf("%s/event/stop?name=%s", AdminAPIAddrEvent, name))
		if err != nil {
			fmt.Printf("Error stopping event: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	},
}

var eventListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available server events",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(fmt.Sprintf("%s/event/list", AdminAPIAddrEvent))
		if err != nil {
			fmt.Printf("Error listing events: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Available server events:")
		fmt.Println(string(body))
	},
}

func init() {
	eventCmd.AddCommand(eventStartCmd)
	eventCmd.AddCommand(eventStopCmd)
	eventCmd.AddCommand(eventListCmd)
	rootCmd.AddCommand(eventCmd)
}