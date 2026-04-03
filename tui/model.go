package tui

import (
	"net/http"
	"sync"
	"time"

	"httpDebugger/pkg/proxy"
	"httpDebugger/pkg/session"
	"httpDebugger/pkg/sessiondata"
	"httpDebugger/tui/panels"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	sessionStore *session.InMemoryStore
	proxy        *proxy.Proxy
	server       *http.Server
	logger       *Logger
	mu           sync.RWMutex
	isRunning    bool
	port         int
	verbose      bool
	sessions     []*sessiondata.Session
	statusMsg    string
	errorMsg     string
	lastUpdate   time.Time
	activeTab    int
	tlsPanel     *panels.TLSPanel

	searchInput     textinput.Model
	isSearching     bool
	filterRegex     string
	sessionsPanel   *panels.SessionsPanel
	requestPanel    *panels.RequestPanel
	responsePanel   *panels.ResponsePanel
	websocketPanel  *panels.WebSocketPanel
	showDetails     bool
	activePanel     ActivePanel
	selectedSession *sessiondata.Session
	showHelp        bool
	width           int
	height          int
}

func NewModel() Model {
	logger, _ := NewLogger(true)
	ti := textinput.New()
	ti.Placeholder = "Regex filter by URL..."
	ti.Prompt = "/ "
	ti.CharLimit = 100
	return Model{
		sessionStore:   session.NewInMemoryStore(1000),
		port:           8080,
		verbose:        false,
		sessions:       make([]*sessiondata.Session, 0),
		sessionsPanel:  panels.NewSessionsPanel(),
		requestPanel:   panels.NewRequestPanel(),
		responsePanel:  panels.NewResponsePanel(),
		websocketPanel: panels.NewWebSocketPanel(),
		activePanel:    SessionPanel,
		showHelp:       false,
		logger:         logger,
		showDetails:    false,
		tlsPanel:       panels.NewTLSPanel(),
		searchInput:    ti,
		isSearching:    false,
		filterRegex:    "",
	}
}

func (m *Model) Init() tea.Cmd {
	return m.tickCmd()
}

func (m *Model) OnShutdown() {
	if m.logger != nil {
		m.logger.Close()
	}
}
