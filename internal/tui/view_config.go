package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateConfig(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	// Editing mode
	if m.configEditing {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				// Save
				key := m.configItems[m.configListCursor].Key
				value := m.configInput.Value()
				m.configEditing = false
				m.configInput.Blur()
				return m, setConfigCmd(key, value)
			case "esc":
				// Cancel
				m.configEditing = false
				m.configInput.Blur()
				return m, nil
			}
		}
		m.configInput, cmd = m.configInput.Update(msg)
		return m, cmd
	}

	// Navigation mode
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.configListCursor > 0 {
				m.configListCursor--
			}
		case "down", "j":
			if m.configListCursor < len(m.configItems)-1 {
				m.configListCursor++
			}
		case "enter":
			if len(m.configItems) > 0 {
				item := m.configItems[m.configListCursor]
				m.configEditing = true
				m.configInput.SetValue(fmt.Sprintf("%v", item.Value))
				m.configInput.Focus()
				return m, nil
			}
		}
	}
	return m, nil
}

func (m Model) viewConfig() string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("[Server Configuration]") + "\n")

	if len(m.configItems) == 0 {
		sb.WriteString("Loading configuration...\n")
		return detailStyle.Render(sb.String())
	}

	// Calculate visible range
	maxItems := m.height - 8
	if maxItems < 5 {
		maxItems = 5
	}
	// Cap at 20 items
	if maxItems > 20 {
		maxItems = 20
	}

	start := 0
	end := len(m.configItems)

	if len(m.configItems) > maxItems {
		if m.configListCursor < maxItems/2 {
			start = 0
			end = maxItems
		} else if m.configListCursor >= len(m.configItems)-maxItems/2 {
			start = len(m.configItems) - maxItems
			end = len(m.configItems)
		} else {
			start = m.configListCursor - maxItems/2
			end = start + maxItems
		}
	}

	// Header
	sb.WriteString(lipgloss.NewStyle().Foreground(subtle).Render(
		fmt.Sprintf("  %-25s %-15s %s", "Key", "Type", "Value"),
	) + "\n")

	for i := start; i < end; i++ {
		item := m.configItems[i]
		cursor := " "
		style := listItemStyle
		
		if i == m.configListCursor {
			cursor = "▶"
			style = listSelectedStyle
		}

		valStr := fmt.Sprintf("%v", item.Value)
		if len(valStr) > 30 {
			valStr = valStr[:27] + "..."
		}

		line := fmt.Sprintf("%s %-25s %-15s %s", cursor, item.Key, item.Type, valStr)
		sb.WriteString(style(line) + "\n")
	}
	
	// Details/Editing pane
	sb.WriteString("\n")
	if m.configEditing {
		sb.WriteString(titleStyle.Render("[Edit Value]") + "\n")
		sb.WriteString(m.configInput.View() + "\n")
		sb.WriteString(lipgloss.NewStyle().Foreground(subtle).Render("(Enter to save, Esc to cancel)"))
	} else if len(m.configItems) > 0 {
		item := m.configItems[m.configListCursor]
		sb.WriteString(titleStyle.Render("[Description]") + "\n")
		sb.WriteString(item.Description + "\n\n")
		if m.lastActionMsg != "" {
			sb.WriteString(titleStyle.Render("[Output]") + "\n" + m.lastActionMsg)
		}
	}

	return detailStyle.Render(sb.String())
}
