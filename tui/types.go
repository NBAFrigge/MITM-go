package tui

import (
	"fmt"

	"httpDebugger/pkg/sessiondata"
)

type ActivePanel int

const (
	SessionPanel ActivePanel = iota
	RequestPanel
	ResponsePanel
)

type ProxyStatusMsg struct {
	Running bool
	Error   error
}

type SessionsUpdatedMsg struct {
	Sessions []*sessiondata.Session
}

type SessionItem struct {
	Session *sessiondata.Session 
}

func (i SessionItem) FilterValue() string {
	return i.Session.Request.URL + " " + i.Session.Request.Method
}

func (i SessionItem) Title() string {
	status := ""
	if i.Session.Response != nil && i.Session.Response.StatusCode > 0 {
		status = fmt.Sprintf(" [%d]", i.Session.Response.StatusCode)
	}
	return fmt.Sprintf("%s %s%s", i.Session.Request.Method, i.Session.Request.URL, status)
}

func (i SessionItem) Description() string {
	sessionType := "HTTP"
	if i.Session.Type == sessiondata.WebSocketSession {
		sessionType = "WebSocket"
	}
	return fmt.Sprintf("Type: %s | Duration: %s | %s",
		sessionType,
		i.Session.Duration.String(),
		i.Session.Timestamp.Format("15:04:05"))
}
