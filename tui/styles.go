package tui

import "github.com/charmbracelet/lipgloss"

var (
	ActiveStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63"))

	InactiveStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	StatusActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10"))

	StatusInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("9"))
)
