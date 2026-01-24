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
	if m.monitorStats.Err != nil {
		return fmt.Sprintf("Error fetching stats: %v", m.monitorStats.Err)
	}

	// Connections
	connections := m.monitorStats.Connections
	
	// System Stats
	goroutines := m.monitorStats.System.Goroutines
	heapAlloc := m.monitorStats.System.HeapAlloc / 1024 / 1024 // MB
	
	// Layout
	// [Connections] 1243
	// [System]      Goroutines: 12  Heap: 45MB
	// [Maps]
	//   map_1  312
	//   map_2  487

	col1 := lipgloss.JoinVertical(lipgloss.Left,
		fmt.Sprintf("%s %d", titleStyle.Render("[Connections]"), connections),
		"\n",
		titleStyle.Render("[System]"),
		fmt.Sprintf("Goroutines: %d", goroutines),
		fmt.Sprintf("Heap Alloc: %d MB", heapAlloc),
	)

	// Maps List
	mapsStr := titleStyle.Render("[Top Maps]") + "\n"
	if len(m.monitorStats.Maps) == 0 {
		mapsStr += "No active users on maps."
	} else {
		for _, mapStat := range m.monitorStats.Maps {
			mapsStr += fmt.Sprintf("  Map %d: %d users\n", mapStat.ID, mapStat.Users)
		}
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		listStyle.Render(col1),
		detailStyle.Render(mapsStr),
	)
}
