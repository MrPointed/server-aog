package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
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
	lastActionMsg  string

	// Monitor State
	monitorStats  MonitorStatsMsg
	netInHistory  []uint64
	netOutHistory []uint64
	lastTotalIn   uint64
	lastTotalOut  uint64

	// Log State
	logLines      []string
	logOffset     int64
	logAutoScroll bool
	logViewOffset int // Lines from bottom

	// Users State
	connectedUsers []struct {
		Addr string `json:"addr"`
		User string `json:"user"`
	}
	filteredUsers []struct {
		Addr string `json:"addr"`
		User string `json:"user"`
	}
	userListCursor     int
	userActionMenuOpen bool
	userActionCursor   int
	selectedUser       struct {
		Addr string
		User string
	}
	userFilter textinput.Model

	// Maps State
	loadedMaps []struct {
		ID    int `json:"id"`
		Users int `json:"users"`
		NPCs  int `json:"npcs"`
	}
	mapListCursor      int
	mapActionMenuOpen  bool
	mapActionCursor    int
	selectedMapID      int
	
	// Config State
	configItems      []ConfigItem
	configListCursor int
	configEditing    bool
	configInput      textinput.Model
}

func InitialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Filter users... (Press / to focus)"
	ti.CharLimit = 156
	ti.Width = 30

	ci := textinput.New()
	ci.Placeholder = "Enter new value..."
	ci.CharLimit = 100
	ci.Width = 40

	return Model{
		tabs: []string{
			"Control", "Monitor", "Logs", "Users", "Maps", 
			"Config", "Econ", "Events", "Mod", "Sim",
		},
		activeTab: 0,
		
		// Control Tab Init
		controlChoices: []string{
			"Start",
			"Restart (graceful)",
			"Stop (graceful)",
			"Stop (force)",
			"Reload config",
		},
		controlCursor:  0,
		serverStatus:   "CHECKING...",
		startTime:      time.Now(),
		
		logAutoScroll: true,
		userFilter:    ti,
		configInput:   ci,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), checkServerStatusCmd())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if !m.configEditing {
				m.quitting = true
				return m, tea.Quit
			}
		case "tab":
			if !m.configEditing {
				m.activeTab++
				if m.activeTab >= len(m.tabs) {
					m.activeTab = 0
				}
			}
		case "shift+tab":
			if !m.configEditing {
				m.activeTab--
				if m.activeTab < 0 {
					m.activeTab = len(m.tabs) - 1
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		// Global ticks
		cmds = append(cmds, tickCmd(), checkServerStatusCmd())
		
		// Tab specific ticks
		if m.activeTab == 1 { // Monitor Tab
			cmds = append(cmds, fetchMonitorStatsCmd())
		}
		if m.activeTab == 2 { // Logs Tab
			cmds = append(cmds, readLogsCmd(m.logOffset))
		}
		if m.activeTab == 3 { // Users Tab
			cmds = append(cmds, fetchUserListCmd())
		}
		if m.activeTab == 4 { // Maps Tab
			cmds = append(cmds, fetchMapsCmd())
		}
		if m.activeTab == 5 { // Config Tab
			if !m.configEditing {
				cmds = append(cmds, fetchConfigListCmd())
			}
		}
	
	case ServerStatusMsg:
		m.serverStatus = msg.Status
		if msg.Running {
			m.startTime = msg.StartTime
		}
	
	case ActionMsg:
		if msg.Err != nil {
			m.lastActionMsg = fmt.Sprintf("Error: %v", msg.Err)
		} else {
			m.lastActionMsg = msg.Output
		}
	
	case MonitorStatsMsg:
		m.monitorStats = msg
		
		currIn := msg.Network.TotalIn
		currOut := msg.Network.TotalOut

		deltaIn := currIn - m.lastTotalIn
		deltaOut := currOut - m.lastTotalOut

		// Ignore huge jump on first run/restart or if stats reset
		if m.lastTotalIn == 0 || currIn < m.lastTotalIn {
			deltaIn = 0
		}
		if m.lastTotalOut == 0 || currOut < m.lastTotalOut {
			deltaOut = 0
		}

		m.lastTotalIn = currIn
		m.lastTotalOut = currOut

		m.netInHistory = append(m.netInHistory, deltaIn)
		m.netOutHistory = append(m.netOutHistory, deltaOut)

		maxHistory := 30
		if len(m.netInHistory) > maxHistory {
			m.netInHistory = m.netInHistory[1:]
		}
		if len(m.netOutHistory) > maxHistory {
			m.netOutHistory = m.netOutHistory[1:]
		}

	case LogMsg:
		if len(msg.Lines) > 0 {
			m.logLines = append(m.logLines, msg.Lines...)
			// Keep buffer size managed
			if len(m.logLines) > 1000 {
				m.logLines = m.logLines[len(m.logLines)-1000:]
			}
			m.logOffset = msg.NewOffset
		}

	case UserListMsg:
		if msg.Err == nil {
			m.connectedUsers = msg.Users
			m.filterUsers()
		}
		
	case MapListMsg:
		if msg.Err == nil {
			m.loadedMaps = msg.Maps
		}

	case ConfigListMsg:
		if msg.Err == nil {
			m.configItems = msg.Items
		}
	}

	// Dispatch to active tab logic if needed (e.g. navigation)
	var cmd tea.Cmd
	switch m.activeTab {
	case 0:
		m, cmd = m.updateControl(msg)
	case 1:
		m, cmd = m.updateMonitor(msg)
	case 2:
		m, cmd = m.updateLogs(msg)
	case 3:
		m, cmd = m.updateUsers(msg)
	case 4:
		m, cmd = m.updateMaps(msg)
	case 5:
		m, cmd = m.updateConfig(msg)
	default:
		// Other tabs not implemented yet
	}
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) filterUsers() {
	filter := strings.ToLower(m.userFilter.Value())
	if filter == "" {
		m.filteredUsers = m.connectedUsers
		return
	}
	var filtered []struct {
		Addr string `json:"addr"`
		User string `json:"user"`
	}
	for _, u := range m.connectedUsers {
		if strings.Contains(strings.ToLower(u.User), filter) || strings.Contains(u.Addr, filter) {
			filtered = append(filtered, u)
		}
	}
	m.filteredUsers = filtered
	
	// Reset cursor if out of bounds
	if m.userListCursor >= len(m.filteredUsers) {
		m.userListCursor = 0
	}
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
	case 2:
		content = m.viewLogs()
	case 3:
		content = m.viewUsers()
	case 4:
		content = m.viewMaps()
	case 5:
		content = m.viewConfig()
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