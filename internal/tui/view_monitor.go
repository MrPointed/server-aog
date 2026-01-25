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
	
	// Layout - Compact
	// [System Status]
	// Connections: 1243 | Goroutines: 12 | Heap: 45 MB

	infoLine := fmt.Sprintf("Connections: %d  |  Goroutines: %d  |  Heap: %d MB", 
		connections, goroutines, heapAlloc)

	col1 := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("[System Status]"),
		infoLine,
	)

	// Maps List
	// Calculate available space
	// Tabs (~3) + Margin (~2) + System Stats (~2) + Spacer (1) + Maps Header (1) = ~9 lines used
	availableLines := m.height - 9
	if availableLines < 0 {
		availableLines = 0
	}

	mapsStr := titleStyle.Render("[Top Maps]") + "\n"
	if len(m.monitorStats.Maps) == 0 {
		mapsStr += "No active users on maps."
	} else {
		count := 0
		for _, mapStat := range m.monitorStats.Maps {
			if count >= availableLines {
				mapsStr += fmt.Sprintf("... (+%d more)\n", len(m.monitorStats.Maps)-count)
				break
			}
			mapsStr += fmt.Sprintf("  Map %d: %d users\n", mapStat.ID, mapStat.Users)
			count++
		}
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		detailStyle.Render(col1),
		"\n",
		detailStyle.Render(mapsStr),
	)
}
