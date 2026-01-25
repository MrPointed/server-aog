package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateLogs(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "p":
			m.logAutoScroll = !m.logAutoScroll
			if m.logAutoScroll {
				m.logViewOffset = 0
			}
		case "up", "k":
			if !m.logAutoScroll {
				m.logViewOffset++
				if m.logViewOffset > len(m.logLines)-1 {
					m.logViewOffset = len(m.logLines) - 1
				}
			}
		case "down", "j":
			if !m.logAutoScroll {
				m.logViewOffset--
				if m.logViewOffset < 0 {
					m.logViewOffset = 0
				}
			}
		}
	}
	return m, nil
}

func (m Model) viewLogs() string {
	var sb strings.Builder
	
	sb.WriteString(titleStyle.Render("[Live Logs]") + " (Press 'p' to pause/scroll)\n")
	
	// Determine how many lines fit
	maxLines := m.height - 6
	if maxLines < 5 {
		maxLines = 5
	}

	totalLines := len(m.logLines)
	if totalLines == 0 {
		return detailStyle.Render(sb.String() + "No logs yet...")
	}

	// Calculate start index based on scroll
	// Default (AutoScroll): Show last maxLines
	// Manual: Show last maxLines - logViewOffset
	
	end := totalLines - m.logViewOffset
	if end > totalLines { end = totalLines }
	if end < 0 { end = 0 }
	
	start := end - maxLines
	if start < 0 {
		start = 0
	}
	
	// Adjust end if we hit top
	if end < start {
		end = start
	}

	visibleLines := m.logLines[start:end]

	for _, line := range visibleLines {
		style := lipgloss.NewStyle()
		if strings.Contains(line, "ERROR") || strings.Contains(line, "Error") || strings.Contains(line, "fail") {
			style = style.Foreground(danger)
		} else if strings.Contains(line, "Warning") || strings.Contains(line, "WARN") {
			style = style.Foreground(lipgloss.Color("#FFA500")) // Orange
		} else if strings.Contains(line, "DEBUG") {
			style = style.Foreground(subtle)
		}
		
		sb.WriteString(style.Render(line) + "\n")
	}

	status := "SCROLLING"
	if !m.logAutoScroll {
		status = "PAUSED"
	}
	
	sb.WriteString(lipgloss.NewStyle().Foreground(subtle).Render(
		fmt.Sprintf("Status: %s | Lines: %d | Scroll: -%d", status, totalLines, m.logViewOffset),
	))

	return detailStyle.Render(sb.String())
}

