package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time

type Model struct {
	// Global State
	activeTab int
	tabs      []string
	width     int
	height    int
	quitting  bool

	// Server Control State
	controlChoices []string
	controlCursor  int
	serverStatus   string
	startTime      time.Time
}

func InitialModel() Model {
	return Model{
		tabs: []string{
			"Control", "Monitor", "Logs", "Users", "Maps", 
			"Config", "Econ", "Debug", "Events", "Mod", "Sim",
		},
		activeTab: 0,
		
		// Control Tab Init
		controlChoices: []string{
			"Restart (graceful)",
			"Stop (graceful)",
			"Stop (force)",
			"Reload config",
		},
		controlCursor:  0,
		serverStatus:   "RUNNING",
		startTime:      time.Now().Add(-3 * time.Hour),
	}
}

func (m Model) Init() tea.Cmd {
	return tickCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "tab":
			m.activeTab++
			if m.activeTab >= len(m.tabs) {
				m.activeTab = 0
			}
		case "shift+tab":
			m.activeTab--
			if m.activeTab < 0 {
				m.activeTab = len(m.tabs) - 1
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		return m, tickCmd()
	}

	// Dispatch to active tab
	var cmd tea.Cmd
	switch m.activeTab {
	case 0:
		m, cmd = m.updateControl(msg)
	case 1:
		m, cmd = m.updateMonitor(msg)
	default:
		// Other tabs not implemented yet
	}

	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return "Bye!\n"
	}

	doc := strings.Builder{}

	// Tabs
	var renderedTabs []string
	for i, t := range m.tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.tabs)-1, i == m.activeTab
		if isActive {
			style = activeTabStyle.Copy()
		} else {
			style = tabStyle.Copy()
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	
	// Content Border
	// We want a box that connects with the active tab.
	// For simplicity in this iteration, we just render the content below with margin.
	
	var content string
	switch m.activeTab {
	case 0:
		content = m.viewControl()
	case 1:
		content = m.viewMonitor()
	default:
		content = fmt.Sprintf("View for %s not implemented yet.", m.tabs[m.activeTab])
	}
	
	doc.WriteString(appStyle.Render(content))

	return doc.String()
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}