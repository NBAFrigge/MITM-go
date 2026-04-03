package tui

import (
	"fmt"

	"httpDebugger/pkg/sessiondata"
	"httpDebugger/tui/helpers"

	"github.com/charmbracelet/lipgloss"
)

var (
	TabStyle       = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("240"))
	ActiveTabStyle = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("63")).Bold(true).Underline(true)
)

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

	sessionStyle := InactiveStyle.Copy().Width(helpers.SafeInt(sessionW - 2)).Height(helpers.SafeInt(availH - 2))
	detailStyle := InactiveStyle.Copy().Width(helpers.SafeInt(detailsW - 2)).Height(helpers.SafeInt(availH - 2))

	if m.activePanel == SessionPanel {
		sessionStyle = ActiveStyle.Copy().Width(helpers.SafeInt(sessionW - 2)).Height(helpers.SafeInt(availH - 2))
	} else {
		detailStyle = ActiveStyle.Copy().Width(helpers.SafeInt(detailsW - 2)).Height(helpers.SafeInt(availH - 2))
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

	left := "● Proxy running"
	leftStyle := StatusActiveStyle
	if !m.isRunning {
		left = "○ Proxy stopped"
		leftStyle = StatusInactiveStyle
	}

	var right string
	if m.errorMsg != "" {
		right = "ERR: " + m.errorMsg
	} else if m.statusMsg != "" {
		right = m.statusMsg
	}

	if m.filterRegex != "" {
		if right != "" {
			right += " "
		}
		right += fmt.Sprintf("/%s/", m.filterRegex)
	}

	maxRight := m.width - len(left) - 4
	if maxRight > 0 && len(right) > maxRight {
		right = helpers.TruncateString(right, maxRight)
	}

	if right != "" {
		return leftStyle.Render(left) + "  " + HelpStyle.Render(right)
	}
	return leftStyle.Render(left)
}

func (m *Model) renderHelpBar() string {
	help := "Tab: panels • Enter: details • Ctrl+S: proxy • /: filter • r: replay • c: curl • q: quit"
	help = helpers.TruncateString(help, m.width-1)
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
  /                 Search (regex filter by URL)
  r                 Replay selected request
  c                 Copy as cURL
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
