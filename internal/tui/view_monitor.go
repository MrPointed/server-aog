package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Placeholder for monitoring state
// In a real implementation, this would be updated via a subscription or tick
func (m Model) updateMonitor(msg tea.Msg) (Model, tea.Cmd) {
	// Handle keys specific to monitor view if needed
	return m, nil
}

func (m Model) viewMonitor() string {
	// Mock Data
	connections := 1243
	// packetsPerSec := 250 // Unused for now
	
	// Layout
	// [Connections] 1243
	// [Packets/s]   ████▆▆▅▃
	// [Maps]
	//   map_1  312
	//   map_2  487

	col1 := lipgloss.JoinVertical(lipgloss.Left,
		fmt.Sprintf("%s %d", titleStyle.Render("[Connections]"), connections),
		fmt.Sprintf("%s %s", titleStyle.Render("[Packets/s]"), lipgloss.NewStyle().Foreground(special).Render("████▆▆▅▃")),
	)

	mapsList := fmt.Sprintf(
		"%s\n  %s %d\n  %s %d",
		titleStyle.Render("[Maps]"),
		"map_1", 312,
		"map_2", 487,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		col1,
		"\n",
		mapsList,
	)
}
