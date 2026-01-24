package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateControl(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.controlCursor > 0 {
				m.controlCursor--
			}
		case "down", "j":
			if m.controlCursor < len(m.controlChoices)-1 {
				m.controlCursor++
			}
		case "enter", " ":
			action := ""
			switch m.controlCursor {
			case 0:
				action = "start"
			case 1:
				action = "restart_graceful"
			case 2:
				action = "stop_graceful"
			case 3:
				action = "stop_force"
			case 4:
				action = "reload_config"
			}
			return m, execServerActionCmd(action)
		}
	}
	return m, nil
}

func (m Model) viewControl() string {
	// Server Status Section
	statusColor := statusRunningStyle
	if m.serverStatus != "RUNNING" {
		statusColor = statusStoppedStyle
	}
	
	uptime := time.Since(m.startTime).Round(time.Second)
	if m.serverStatus != "RUNNING" {
		uptime = 0
	}
	
	statusView := fmt.Sprintf(
		"%s\n%s: %s\n%s: %s\n",
		titleStyle.Render("[Server]"),
		"Status", statusColor.Render(m.serverStatus),
		"Uptime", uptime.String(),
	)

	// Actions Section
	s := titleStyle.Render("[Actions]") + "\n"

	for i, choice := range m.controlChoices {
		cursor := " " // no cursor
		if m.controlCursor == i {
			cursor = "▶"
		}

		if m.controlCursor == i {
			s += listSelectedStyle(fmt.Sprintf("%s %s", cursor, choice)) + "\n"
		} else {
			s += listItemStyle(fmt.Sprintf("%s %s", cursor, choice)) + "\n"
		}
	}
	
	// Last Action Output
	if m.lastActionMsg != "" {
		s += "\n" + titleStyle.Render("[Output]") + "\n" + m.lastActionMsg + "\n"
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		listStyle.Render(statusView),
		detailStyle.Render(s),
	) + "\n\nHelp: Tab to switch views • ↑/↓ to move • enter to select"
}
