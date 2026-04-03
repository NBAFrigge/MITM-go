package tui

import (
	"fmt"

	"httpDebugger/pkg/sessiondata"

	"github.com/charmbracelet/lipgloss"
)

var (
	TabStyle       = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("240"))
	ActiveTabStyle = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("63")).Bold(true).Underline(true)
)

func safe(v int) int {
	if v < 0 {
		return 0
	}
	return v
}

func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) > w {
		if w > 3 {
			return string(runes[:w-3]) + "..."
		}
		return string(runes[:w])
	}
	return s
}

func (m *Model) View() string {
	if m.width == 0 {
		return "Initializing HTTP Debugger..."
	}
	if m.showHelp {
		return m.renderHelpScreen()
	}

	availW := m.width
	availH := m.height - 2

	var sessionW, detailsW int

	if !m.showDetails {
		sessionW = availW
	} else {
		sessionW = availW / 3
		detailsW = availW - sessionW
	}

	sessionStyle := InactiveStyle.Copy().Width(safe(sessionW - 2)).Height(safe(availH - 2))
	detailStyle := InactiveStyle.Copy().Width(safe(detailsW - 2)).Height(safe(availH - 2))

	if m.activePanel == SessionPanel {
		sessionStyle = ActiveStyle.Copy().Width(safe(sessionW - 2)).Height(safe(availH - 2))
	} else {
		detailStyle = ActiveStyle.Copy().Width(safe(detailsW - 2)).Height(safe(availH - 2))
	}

	sessionsContent := sessionStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			m.renderPanelTitle("Sessions", m.activePanel == SessionPanel),
			"",
			m.sessionsPanel.View(),
		),
	)

	var mainContent string

	if m.showDetails {
		selectedSession := m.sessionsPanel.GetSelectedSession()
		rightSide := ""

		if selectedSession != nil && selectedSession.Type == sessiondata.WebSocketSession {
			rightSide = detailStyle.Render(
				lipgloss.JoinVertical(lipgloss.Left,
					m.renderPanelTitle("WebSocket Details", true),
					"",
					m.websocketPanel.View(),
				),
			)
		} else {
			tabs := []string{"Request", "Response", "TLS Fingerprint"}
			var tabHeaders string
			for i, t := range tabs {
				if i == m.activeTab {
					tabHeaders += ActiveTabStyle.Render(t) + " "
				} else {
					tabHeaders += TabStyle.Render(t) + " "
				}
			}

			var detailContent string
			switch m.activeTab {
			case 0:
				detailContent = m.requestPanel.View()
			case 1:
				detailContent = m.responsePanel.View()
			case 2:
				if m.tlsPanel != nil {
					detailContent = m.tlsPanel.View()
				}
			}

			rightSide = detailStyle.Render(
				lipgloss.JoinVertical(lipgloss.Left,
					tabHeaders,
					"",
					detailContent,
				),
			)
		}

		mainContent = lipgloss.JoinHorizontal(lipgloss.Top, sessionsContent, rightSide)
	} else {
		mainContent = sessionsContent
	}

	statusBar := m.renderStatusBar()
	helpBar := m.renderHelpBar()

	return lipgloss.JoinVertical(lipgloss.Left, mainContent, statusBar, helpBar)
}

func (m *Model) renderPanelTitle(title string, isActive bool) string {
	style := TitleStyle
	if isActive {
		style = style.Foreground(lipgloss.Color("63"))
	} else {
		style = style.Foreground(lipgloss.Color("240"))
	}
	return style.Render(title)
}

func (m *Model) renderStatusBar() string {
	if m.isSearching {
		return StatusActiveStyle.Render(m.searchInput.View())
	}
	var status string
	if m.isRunning {
		status = fmt.Sprintf("Proxy running on port %d", m.port)
	} else {
		status = "Proxy stopped"
	}
	if m.statusMsg != "" {
		status += " | " + m.statusMsg
	}
	if m.errorMsg != "" {
		status += " | ERROR: " + m.errorMsg
	}
	if m.selectedSession != nil {
		status += fmt.Sprintf(" | Selected: %s", m.selectedSession.ID[:8])
	}
	if m.verbose {
		status += " | VERBOSE"
	}

	if m.filterRegex != "" {
		status += fmt.Sprintf(" | FILTER: /%s/", m.filterRegex)
	}

	status = truncate(status, m.width-1)

	if m.isRunning {
		return StatusActiveStyle.Render(status)
	}
	return StatusInactiveStyle.Render(status)
}

func (m *Model) renderHelpBar() string {
	help := "Tab: focus panels • Enter: toggle details • Ctrl+S: proxy • Ctrl+R: refresh • Ctrl+D: clear • Q: quit"
	if m.activePanel == SessionPanel {
		help = "↑↓: navigate • " + help
	} else {
		help = "↑↓: scroll • ←→: switch tabs • " + help
	}

	help = truncate(help, m.width-1)
	return HelpStyle.Render(help)
}

func (m *Model) renderHelpScreen() string {
	helpContent := `
HTTP Debugger - Help
GLOBAL COMMANDS:
  Tab               Toggle focus (Left/Right)
  Enter             Select session (in Sessions panel)
  Ctrl+S            Start/Stop proxy
  Ctrl+R            Refresh sessions
  Ctrl+D            Clear all sessions
  Ctrl+C / Q        Quit application
  Escape            Reset selection
  F1                Toggle this help
  F2                Toggle verbose logging
DETAILS PANEL:
  ↑↓                Scroll through content
  ←→                Switch Tab (Req/Res/TLS)
  PgUp/PgDn         Page up/down
`
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(2).
		Width(m.width - 4).
		Height(m.height - 4)
	return style.Render(helpContent)
}
