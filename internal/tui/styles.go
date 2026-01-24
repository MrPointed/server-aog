package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	danger    = lipgloss.AdaptiveColor{Light: "#F05D5E", Dark: "#FF5F87"}

	// Layout
	appStyle = lipgloss.NewStyle().Margin(1, 2)

	// List
	listStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(subtle).
			MarginRight(2).
			Height(8).
			Width(30)

	listHeaderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(subtle).
			MarginRight(2).
			Render

	listItemStyle = lipgloss.NewStyle().PaddingLeft(2).Render

	checkMark = lipgloss.NewStyle().SetString("✓").
			Foreground(special).
			PaddingRight(1).
			String()

	listSelectedStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(highlight).
				Foreground(highlight).
				PaddingLeft(1).
				Render

	// Details
	detailStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	titleStyle = lipgloss.NewStyle().
			Foreground(highlight).
			Bold(true).
			MarginBottom(1)

	statusRunningStyle = lipgloss.NewStyle().
				Foreground(special).
				Bold(true)

	statusStoppedStyle = lipgloss.NewStyle().
				Foreground(danger).
				Bold(true)

	// Tabs
	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	tabStyle = lipgloss.NewStyle().
			Border(tabBorder, true).
			BorderForeground(highlight).
			Padding(0, 1)

	activeTabStyle = tabStyle.Copy().
			Border(activeTabBorder, true)

	tabGap = lipgloss.NewStyle().
		Width(1).
		Render(" ")
)
