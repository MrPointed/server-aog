package main

import (
	"fmt"

	"os"

	"github.com/ao-go-server/internal/tui"
	"github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor server in real-time",
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(tui.InitialModel())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(monitorCmd)
}
