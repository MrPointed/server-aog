package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateUsers(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	
	// Handle filter focus
	if m.userFilter.Focused() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter", "esc":
				m.userFilter.Blur()
				return m, nil
			}
		}
		m.userFilter, cmd = m.userFilter.Update(msg)
		m.filterUsers() // Re-filter on every keystroke
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.userActionMenuOpen {
			switch msg.String() {
			case "esc":
				m.userActionMenuOpen = false
			case "up", "k":
				if m.userActionCursor > 0 {
					m.userActionCursor--
				}
			case "down", "j":
				if m.userActionCursor < 2 { // 0: Inspect, 1: Kick, 2: Ban
					m.userActionCursor++
				}
			case "enter":
				action := m.userActionCursor
				m.userActionMenuOpen = false
				user := m.selectedUser.User
				
				switch action {
				case 0:
					return m, inspectUserCmd(user)
				case 1:
					return m, kickUserCmd(user)
				case 2:
					return m, banUserCmd(user)
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "/":
			m.userFilter.Focus()
			return m, textinput.Blink
		case "up", "k":
			if m.userListCursor > 0 {
				m.userListCursor--
			}
		case "down", "j":
			if m.userListCursor < len(m.filteredUsers)-1 {
				m.userListCursor++
			}
		case "enter":
			if len(m.filteredUsers) > 0 {
				m.userActionMenuOpen = true
				m.userActionCursor = 0
				u := m.filteredUsers[m.userListCursor]
				m.selectedUser.User = u.User
				m.selectedUser.Addr = u.Addr
			}
		}
	}
	return m, nil
}

func (m Model) viewUsers() string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("[Connected Users]") + "\n")
	sb.WriteString(m.userFilter.View() + "\n\n")

	if len(m.filteredUsers) == 0 {
		sb.WriteString("No users found.\n")
		return detailStyle.Render(sb.String())
	}

	// Calculate visible range
	maxItems := m.height - 10 // Reserve space for menu/status/filter
	if maxItems < 5 {
		maxItems = 5
	}
	// Cap at 20 items to avoid overwhelming list
	if maxItems > 20 {
		maxItems = 20
	}

	start := 0
	end := len(m.filteredUsers)

	if len(m.filteredUsers) > maxItems {
		// Simple scrolling logic
		if m.userListCursor < maxItems/2 {
			start = 0
			end = maxItems
		} else if m.userListCursor >= len(m.filteredUsers)-maxItems/2 {
			start = len(m.filteredUsers) - maxItems
			end = len(m.filteredUsers)
		} else {
			start = m.userListCursor - maxItems/2
			end = start + maxItems
		}
	}

	for i := start; i < end; i++ {
		user := m.filteredUsers[i]
		cursor := " "
		style := listItemStyle
		
		if i == m.userListCursor {
			cursor = "▶"
			style = listSelectedStyle
		}

		line := fmt.Sprintf("%s %-20s %s", cursor, user.User, user.Addr)
		sb.WriteString(style(line) + "\n")
	}
	
	sb.WriteString(lipgloss.NewStyle().Foreground(subtle).Render(
		fmt.Sprintf("\nTotal: %d | Filtered: %d | Selected: %d", len(m.connectedUsers), len(m.filteredUsers), m.userListCursor+1),
	))

	// Action Menu Overlay
	if m.userActionMenuOpen {
		actions := []string{"Inspect", "Kick", "Ban"}
		menu := "\n\n" + titleStyle.Render(fmt.Sprintf("[Action: %s]", m.selectedUser.User)) + "\n"
		
		for i, action := range actions {
			cursor := " "
			if i == m.userActionCursor {
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
