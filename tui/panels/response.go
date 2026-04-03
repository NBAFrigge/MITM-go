package panels

import (
	"fmt"
	"strings"

	"httpDebugger/pkg/sessiondata"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ResponsePanel struct {
	viewport   viewport.Model
	rawContent string
}

func NewResponsePanel() *ResponsePanel {
	vp := viewport.New(0, 0)

	return &ResponsePanel{
		viewport:   vp,
		rawContent: "Select a session",
	}
}

func (p *ResponsePanel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	p.viewport, cmd = p.viewport.Update(msg)
	return cmd
}

func (p *ResponsePanel) View() string {
	return p.viewport.View()
}

func (p *ResponsePanel) SetSize(width, height int) {
	p.viewport.Width = width
	p.viewport.Height = height

	if p.rawContent != "" {
		wrappedContent := lipgloss.NewStyle().Width(width).Render(p.rawContent)
		p.viewport.SetContent(wrappedContent)
	}
}

func (p *ResponsePanel) UpdateSession(session *sessiondata.Session) {
	if session == nil {
		p.rawContent = "Select a session"
		p.viewport.SetContent(lipgloss.NewStyle().Width(p.viewport.Width).Render(p.rawContent))
		return
	}

	if session.Response == nil {
		p.rawContent = "Response not found"
		p.viewport.SetContent(lipgloss.NewStyle().Width(p.viewport.Width).Render(p.rawContent))
		return
	}
	if session.Response.StatusCode == 0 {
		return
	}
	details := fmt.Sprintf("Status: %s\nDuration: %v\n\n",
		session.Response.Status,
		session.Duration)

	details += "Headers:\n"
	for _, key := range session.Response.Headers.Order {
		if value, ok := session.Response.Headers.Entries[key]; ok {
			if slice, ok := value.([]string); ok {
				details += fmt.Sprintf(" %s: %s\n", key, strings.Join(slice, ", "))
			} else {
				details += fmt.Sprintf(" %s: %s\n", key, value)
			}
		}
	}

	if len(session.Response.Body) > 0 {
		details += fmt.Sprintf("\nBody (%d bytes):\n%s",
			len(session.Response.Body),
			string(session.Response.Body))
	} else {
		details += "\nNo body"
	}

	p.rawContent = details
	wrappedContent := lipgloss.NewStyle().Width(p.viewport.Width).Render(details)
	p.viewport.SetContent(wrappedContent)
}
