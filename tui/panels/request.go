package panels

import (
	"fmt"
	"strings"

	"httpDebugger/pkg/sessiondata"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type RequestPanel struct {
	viewport   viewport.Model
	rawContent string
}

func NewRequestPanel() *RequestPanel {
	vp := viewport.New(0, 0)

	return &RequestPanel{
		viewport:   vp,
		rawContent: "Select a session",
	}
}

func (p *RequestPanel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	p.viewport, cmd = p.viewport.Update(msg)
	return cmd
}

func (p *RequestPanel) View() string {
	return p.viewport.View()
}

func (p *RequestPanel) SetSize(width, height int) {
	p.viewport.Width = width
	p.viewport.Height = height

	if p.rawContent != "" {
		wrappedContent := lipgloss.NewStyle().Width(width).Render(p.rawContent)
		p.viewport.SetContent(wrappedContent)
	}
}

func (p *RequestPanel) UpdateSession(session *sessiondata.Session) {
	if session == nil {
		p.rawContent = "Select a session"
		p.viewport.SetContent(lipgloss.NewStyle().Width(p.viewport.Width).Render(p.rawContent))
		return
	}
	details := fmt.Sprintf("Method: %s\nURL: %s\nTimestamp: %s\n\n",
		session.Request.Method,
		session.Request.URL,
		session.Timestamp.Format("2006-01-02 15:04:05"))

	details += "Headers:\n"
	for _, key := range session.Request.Headers.Order {
		if value, ok := session.Request.Headers.Entries[key]; ok {
			if slice, ok := value.([]string); ok {
				details += fmt.Sprintf(" %s: %s\n", key, strings.Join(slice, ", "))
			} else {
				details += fmt.Sprintf(" %s: %s\n", key, value)
			}
		}
	}

	if len(session.Request.Body) > 0 {
		details += fmt.Sprintf("\nBody (%d bytes):\n%s",
			len(session.Request.Body),
			string(session.Request.Body))
	} else {
		details += "\nNo body"
	}

	p.rawContent = details

	wrappedContent := lipgloss.NewStyle().Width(p.viewport.Width).Render(details)
	p.viewport.SetContent(wrappedContent)
}
