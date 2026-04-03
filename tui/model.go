package tui

import (
	"net/http"
	"regexp"

	"httpDebugger/pkg/proxy"
	"httpDebugger/pkg/session"
	"httpDebugger/pkg/sessiondata"
	"httpDebugger/tui/panels"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	// Proxy
	proxy     *proxy.Proxy
	server    *http.Server
	port      int
	isRunning bool

	// Sessions
	sessionStore    *session.InMemoryStore
	sessions        []*sessiondata.Session
	sessionCount    int
	selectedSession *sessiondata.Session

	// Panels
	sessionsPanel  *panels.SessionsPanel
	requestPanel   *panels.RequestPanel
	responsePanel  *panels.ResponsePanel
	websocketPanel *panels.WebSocketPanel
	tlsPanel       *panels.TLSPanel

	// Navigation
	activePanel ActivePanel
	activeTab   int
	showDetails bool
	showHelp    bool

	// Search
	searchInput    textinput.Model
	isSearching    bool
	filterRegex    string
	compiledFilter *regexp.Regexp

	// Status
	statusMsg string
	errorMsg  string

	// Logger
	logger  *Logger
	verbose bool

	// Layout
	width  int
	height int
}

func NewModel() Model {
	logger, _ := NewLogger(true)

	ti := textinput.New()
	ti.Placeholder = "Regex filter by URL..."
	ti.Prompt = "/ "
	ti.CharLimit = 100

	return Model{
		port:           8080,
		sessionStore:   session.NewInMemoryStore(1000),
		sessions:       make([]*sessiondata.Session, 0),
		sessionsPanel:  panels.NewSessionsPanel(),
		requestPanel:   panels.NewRequestPanel(),
		responsePanel:  panels.NewResponsePanel(),
		websocketPanel: panels.NewWebSocketPanel(),
		tlsPanel:       panels.NewTLSPanel(),
		activePanel:    SessionPanel,
		searchInput:    ti,
		logger:         logger,
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
