package panels

import (
	"fmt"
	"strings"

	"httpDebugger/pkg/sessiondata"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var wsHeaderStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder(), false, false, true, false).
	BorderForeground(lipgloss.Color("240")).
	MarginBottom(1)

type WebSocketPanel struct {
	viewport      viewport.Model
	headerContent string
	rawMessages   string
}

func NewWebSocketPanel() *WebSocketPanel {
	vp := viewport.New(0, 0)
	return &WebSocketPanel{
		viewport:      vp,
		headerContent: "Select a WebSocket session",
	}
}

func (p *WebSocketPanel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	p.viewport, cmd = p.viewport.Update(msg)
	return cmd
}

func (p *WebSocketPanel) View() string {
	if p.headerContent == "Select a WebSocket session" {
		return lipgloss.NewStyle().Width(p.viewport.Width).Render(p.headerContent)
	}

	header := wsHeaderStyle.Width(p.viewport.Width).Render(p.headerContent)
	return lipgloss.JoinVertical(lipgloss.Left, header, p.viewport.View())
}

func (p *WebSocketPanel) SetSize(width, height int) {
	p.viewport.Width = width

	headerHeight := 0
	if p.headerContent != "" && p.headerContent != "Select a WebSocket session" {
		headerHeight = lipgloss.Height(wsHeaderStyle.Render(p.headerContent))
	}

	safeHeight := height - headerHeight
	if safeHeight < 0 {
		safeHeight = 0
	}
	p.viewport.Height = safeHeight

	if p.rawMessages != "" {
		wrappedContent := lipgloss.NewStyle().Width(width).Render(p.rawMessages)
		p.viewport.SetContent(wrappedContent)
	}
}

func (p *WebSocketPanel) UpdateSession(session *sessiondata.Session) {
	if session == nil || session.Type != sessiondata.WebSocketSession {
		p.headerContent = "Select a WebSocket session"
		p.rawMessages = ""
		p.viewport.SetContent("")
		return
	}

	p.headerContent = fmt.Sprintf("WebSocket: %s\nConnected: %s",
		session.Request.URL,
		session.Timestamp.Format("15:04:05"))

	var content strings.Builder

	if len(session.WebSocket.Messages) == 0 {
		content.WriteString("Waiting for messages...")
	} else {
		for _, msg := range session.WebSocket.Messages {
			timestamp := msg.Timestamp.Format("15:04:05")

			direction := "↗"
			dirColor := "42"

			if msg.Direction == 1 {
				direction = "↙"
				dirColor = "205"
			}

			styledDir := lipgloss.NewStyle().Foreground(lipgloss.Color(dirColor)).Render(direction)
			content.WriteString(fmt.Sprintf("%s %s %s\n", timestamp, styledDir, msg.PayloadText))
		}
	}

	p.rawMessages = content.String()
	wrappedContent := lipgloss.NewStyle().Width(p.viewport.Width).Render(p.rawMessages)
	p.viewport.SetContent(wrappedContent)
}
