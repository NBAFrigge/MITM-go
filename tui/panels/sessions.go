package panels

import (
	"fmt"
	"strings"

	"httpDebugger/pkg/sessiondata"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type SessionsPanel struct {
	list list.Model
}

func NewSessionsPanel() *SessionsPanel {
	items := []list.Item{}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)

	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(true)

	return &SessionsPanel{
		list: l,
	}
}

func (p *SessionsPanel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return cmd
}

func (p *SessionsPanel) View() string {
	return p.list.View()
}

func (p *SessionsPanel) SetSize(width, height int) {
	p.list.SetSize(width, height)

	items := p.list.Items()
	for i, item := range items {
		if sItem, ok := item.(sessionListItem); ok {
			sItem.width = width
			items[i] = sItem
		}
	}
	p.list.SetItems(items)
}

func (p *SessionsPanel) UpdateSessions(sessions []*sessiondata.Session) {
	items := []list.Item{}
	width := p.list.Width()
	for _, session := range sessions {
		items = append(items, sessionListItem{session: session, width: width})
	}
	p.list.SetItems(items)
}

func (p *SessionsPanel) GetSelectedSession() *sessiondata.Session {
	if selectedItem := p.list.SelectedItem(); selectedItem != nil {
		if sessionItem, ok := selectedItem.(sessionListItem); ok {
			return sessionItem.session
		}
	}
	return nil
}

type sessionListItem struct {
	session *sessiondata.Session
	width   int
}

func (i sessionListItem) FilterValue() string {
	if i.session == nil || i.session.Request == nil {
		return ""
	}
	return i.session.Request.URL + " " + i.session.Request.Method
}

func truncateString(s string, max int) string {
	if max <= 0 {
		return ""
	}
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	runes := []rune(s)
	if len(runes) > max {
		if max > 3 {
			return string(runes[:max-3]) + "..."
		}
		return string(runes[:max])
	}
	return s
}

func (i sessionListItem) Title() string {
	if i.session == nil || i.session.Request == nil {
		return "Invalid Session"
	}

	status := ""
	if i.session.Response != nil && i.session.Response.StatusCode > 0 {
		status = fmt.Sprintf(" [%d]", i.session.Response.StatusCode)
	}

	raw := fmt.Sprintf("%s %s%s", i.session.Request.Method, i.session.Request.URL, status)

	safeWidth := i.width - 6
	if safeWidth < 0 {
		safeWidth = 0
	}
	return truncateString(raw, safeWidth)
}

func (i sessionListItem) Description() string {
	if i.session == nil {
		return ""
	}

	safeWidth := i.width - 6
	if safeWidth < 0 {
		safeWidth = 0
	}

	sessionType := "HTTP"
	if i.session.Type == sessiondata.WebSocketSession {
		sessionType = "WSS"
		return truncateString(fmt.Sprintf("Type: %s", sessionType), safeWidth)
	}

	if i.session.Response != nil && i.session.Response.ContentType != "" {
		raw := fmt.Sprintf("Type: %s | Content type: %s | Duration: %s | %s",
			sessionType,
			i.session.Response.ContentType,
			i.session.Duration.String(),
			i.session.Timestamp.Format("15:04:05"))
		return truncateString(raw, safeWidth)
	}

	raw := fmt.Sprintf("Type: %s | Duration: %s | %s",
		sessionType,
		i.session.Duration.String(),
		i.session.Timestamp.Format("15:04:05"))
	return truncateString(raw, safeWidth)
}
