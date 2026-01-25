package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateMaps(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.mapActionMenuOpen {
			switch msg.String() {
			case "esc":
				m.mapActionMenuOpen = false
			case "up", "k":
				if m.mapActionCursor > 0 {
					m.mapActionCursor--
				}
			case "down", "j":
				if m.mapActionCursor < 1 { // 0: Reload, 1: Unload
					m.mapActionCursor++
				}
			case "enter":
				action := m.mapActionCursor
				m.mapActionMenuOpen = false
				mapID := m.selectedMapID
				
				switch action {
				case 0:
					return m, reloadMapCmd(mapID)
				case 1:
					return m, unloadMapCmd(mapID)
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.mapListCursor > 0 {
				m.mapListCursor--
			}
		case "down", "j":
			if m.mapListCursor < len(m.loadedMaps)-1 {
				m.mapListCursor++
			}
		case "enter":
			if len(m.loadedMaps) > 0 {
				m.mapActionMenuOpen = true
				m.mapActionCursor = 0
				m.selectedMapID = m.loadedMaps[m.mapListCursor].ID
			}
		}
	}
	return m, nil
}

func (m Model) viewMaps() string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("[Loaded Maps]") + "\n")

	if len(m.loadedMaps) == 0 {
		sb.WriteString("No maps loaded.\n")
		return detailStyle.Render(sb.String())
	}

	// Calculate visible range
	maxItems := m.height - 8 // Reserve space
	if maxItems < 5 {
		maxItems = 5
	}
	// Cap at 20 items
	if maxItems > 20 {
		maxItems = 20
	}

	start := 0
	end := len(m.loadedMaps)

	if len(m.loadedMaps) > maxItems {
		if m.mapListCursor < maxItems/2 {
			start = 0
			end = maxItems
		} else if m.mapListCursor >= len(m.loadedMaps)-maxItems/2 {
			start = len(m.loadedMaps) - maxItems
			end = len(m.loadedMaps)
		} else {
			start = m.mapListCursor - maxItems/2
			end = start + maxItems
		}
	}
	
	// Header
	sb.WriteString(lipgloss.NewStyle().Foreground(subtle).Render(
		fmt.Sprintf("  %-5s %-10s %-10s", "ID", "Users", "NPCs"),
	) + "\n")

	for i := start; i < end; i++ {
		mInfo := m.loadedMaps[i]
		cursor := " "
		style := listItemStyle
		
		if i == m.mapListCursor {
			cursor = "▶"
			style = listSelectedStyle
		}

		line := fmt.Sprintf("%s %-5d %-10d %-10d", cursor, mInfo.ID, mInfo.Users, mInfo.NPCs)
		sb.WriteString(style(line) + "\n")
	}
	
	sb.WriteString(lipgloss.NewStyle().Foreground(subtle).Render(
		fmt.Sprintf("\nTotal Loaded: %d", len(m.loadedMaps)),
	))

	// Action Menu Overlay
	if m.mapActionMenuOpen {
		actions := []string{"Reload", "Unload"}
		menu := "\n\n" + titleStyle.Render(fmt.Sprintf("[Action: Map %d]", m.selectedMapID)) + "\n"
		
		for i, action := range actions {
			cursor := " "
			if i == m.mapActionCursor {
				cursor = "▶"
				menu += listSelectedStyle(fmt.Sprintf("%s %s", cursor, action)) + "\n"
			} else {
				menu += listItemStyle(fmt.Sprintf("%s %s", cursor, action)) + "\n"
			}
		}
		menu += lipgloss.NewStyle().Foreground(subtle).Render("(Esc to cancel)")
		sb.WriteString(menu)
	} else if m.lastActionMsg != "" {
		// Last Action Output
		sb.WriteString("\n\n" + titleStyle.Render("[Output]") + "\n" + m.lastActionMsg + "\n")
	}

	return detailStyle.Render(sb.String())
}
