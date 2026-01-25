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
		}
	}
	return m, nil
}

func (m Model) viewLogs() string {
	var sb strings.Builder
	
	sb.WriteString(titleStyle.Render("[Live Logs]") + " (Press 'p' to pause/resume)\n")
	
	// Determine how many lines fit
	// Header takes ~2 lines, Tabs ~2 lines.
	// Simple approximation
	maxLines := m.height - 6
	if maxLines < 5 {
		maxLines = 5
	}

	start := len(m.logLines) - maxLines
	if start < 0 {
		start = 0
	}

	visibleLines := m.logLines[start:]

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
		fmt.Sprintf("Status: %s | Lines: %d | Offset: %d", status, len(m.logLines), m.logOffset),
	))

	return detailStyle.Render(sb.String())
}

