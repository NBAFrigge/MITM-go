package tui

import (
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
