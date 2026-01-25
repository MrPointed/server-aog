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

	// Network Stats
	// Calculate current rate from last history item
	var rateIn, rateOut uint64
	if len(m.netInHistory) > 0 {
		rateIn = m.netInHistory[len(m.netInHistory)-1]
	}
	if len(m.netOutHistory) > 0 {
		rateOut = m.netOutHistory[len(m.netOutHistory)-1]
	}
	
	// Convert to KB/s
	kbIn := float64(rateIn) / 1024.0
	kbOut := float64(rateOut) / 1024.0
	
	sparkIn := renderSparkline(m.netInHistory, 15)
	sparkOut := renderSparkline(m.netOutHistory, 15)
	
	sparkStyle := lipgloss.NewStyle().Foreground(special)
	
	netInfo := fmt.Sprintf("IN : %s %6.1f KB/s\nOUT: %s %6.1f KB/s", 
		sparkStyle.Render(sparkIn), kbIn, 
		sparkStyle.Render(sparkOut), kbOut)
	
	colNet := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("[Network Traffic]"),
		netInfo,
	)

	// Sort users by traffic (Total Bytes)
	type userTraffic struct {
		User  string
		Total uint64
		In    uint64
		Out   uint64
	}
	var users []userTraffic
	for _, u := range m.monitorStats.Network.Connections {
		users = append(users, userTraffic{
			User:  u.User,
			Total: u.BytesIn + u.BytesOut,
			In:    u.BytesIn,
			Out:   u.BytesOut,
		})
	}
	// Simple bubble sort for top 5
	for i := 0; i < len(users)-1; i++ {
		for j := 0; j < len(users)-i-1; j++ {
			if users[j].Total < users[j+1].Total {
				users[j], users[j+1] = users[j+1], users[j]
			}
		}
	}
	
	topUsersStr := titleStyle.Render("[Top Users Traffic]") + "\n"
	if len(users) == 0 {
		topUsersStr += "No active traffic."
	} else {
		count := 0
		for _, u := range users {
			if count >= 5 { break }
			topUsersStr += fmt.Sprintf("  %-15s In: %-8d Out: %-8d\n", u.User, u.In, u.Out)
			count++
		}
	}

	// Maps List (Limit to 5)
	mapsStr := titleStyle.Render("[Top Maps]") + "\n"
	if len(m.monitorStats.Maps) == 0 {
		mapsStr += "No active users on maps."
	} else {
		count := 0
		for _, mapStat := range m.monitorStats.Maps {
			if count >= 5 {
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
		detailStyle.Render(colNet),
		"\n",
		detailStyle.Render(topUsersStr),
		"\n",
		detailStyle.Render(mapsStr),
	)
}
