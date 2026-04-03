package tui

import (
	"time"

	"httpDebugger/pkg/sessiondata"

	tea "github.com/charmbracelet/bubbletea"
)

type WebSocketMessageEvent struct {
	SessionID string
	Message   sessiondata.WebSocketMessage
}

type TickEvent struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickEvent{}
	})
}
