package handlers

import (
	"httpDebugger/internal/certs"
	"net/http"

	"httpDebugger/internal/proxy/types"
)

// Manager manages different types of handlers
type Manager struct {
	config      *types.Config
	caCache     *certs.CertCache
	httpHandler *HTTPHandler
	mitmHandler *MITMHandler
	wsHandler   *WebSocketHandler
}

// NewManager creates a new Manager instance
func NewManager(config *types.Config, cache *certs.CertCache) *Manager {
	return &Manager{
		config:      config,
		caCache:     cache,
		httpHandler: NewHTTPHandler(config),
		mitmHandler: NewMITMHandler(config, cache),
		wsHandler:   NewWebSocketHandler(config, cache),
	}
}

func (m *Manager) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	m.httpHandler.Handle(w, r)
}

func (m *Manager) HandleMITM(w http.ResponseWriter, r *http.Request) {
	m.mitmHandler.Handle(w, r)
}

func (m *Manager) GetWebSocketHandler() *WebSocketHandler {
	return m.wsHandler
}
