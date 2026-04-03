package tui

import (
	"httpDebugger/pkg/sessiondata"
)

type WebSocketMessageEvent struct {
	SessionID string
	Message   sessiondata.WebSocketMessage
}
