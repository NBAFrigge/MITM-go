package tui

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"httpDebugger/pkg/certs"
	"httpDebugger/pkg/proxy"
	"httpDebugger/pkg/sessiondata"

	"github.com/atotto/clipboard"
	key "github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type TickMsg struct{}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.isSearching {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				m.isSearching = false
				m.searchInput.Blur()
				m.filterRegex = m.searchInput.Value()
				m.applyFilter()
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				m.isSearching = false
				m.searchInput.Blur()
				m.searchInput.SetValue(m.filterRegex)
				return m, nil
			}
		}

		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updatePanelSizes()

	case SessionsUpdatedMsg:
		m.sessions = msg.Sessions
		m.sessionCount = len(msg.Sessions)
		m.applyFilter()

		if m.showDetails && m.selectedSession != nil {
			for _, session := range m.sessions {
				if session.ID == m.selectedSession.ID {
					m.selectedSession = session
					m.updatePanelsForSession(session)
					break
				}
			}
		}

		return m, m.tickCmd()

	case ProxyStatusMsg:
		m.isRunning = msg.Running
		if !msg.Running {
			m.server = nil
		}
		if msg.Error != nil {
			m.errorMsg = msg.Error.Error()
		} else {
			m.errorMsg = ""
		}

	case ClearStatusMsg:
		m.statusMsg = ""
		m.errorMsg = ""
		return m, nil

	case ReplayResultMsg:
		if msg.Error != nil {
			m.errorMsg = fmt.Sprintf("Replay failed: %v", msg.Error)
			if m.logger != nil {
				m.logger.LogError(msg.Error, "Replay failed")
			}
		} else {
			m.statusMsg = "Replay sent"
			if m.logger != nil {
				m.logger.LogInfo("Replay: sent successfully")
			}
		}
		return m, clearStatusCmd()

	case TickMsg:
		return m, m.tickCmd()

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c", "q"))):
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("/"))):
			m.isSearching = true
			m.searchInput.Focus()
			return m, textinput.Blink

		case key.Matches(msg, key.NewBinding(key.WithKeys("tab")), key.NewBinding(key.WithKeys("shift+tab"))):
			m.switchPanel()

		case key.Matches(msg, key.NewBinding(key.WithKeys("right"))):
			if m.showDetails && m.activePanel != SessionPanel {
				m.activeTab = (m.activeTab + 1) % 3
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("left"))):
			if m.showDetails && m.activePanel != SessionPanel {
				m.activeTab--
				if m.activeTab < 0 {
					m.activeTab = 2
				}
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.activePanel == SessionPanel {
				highlighted := m.sessionsPanel.GetSelectedSession()

				if m.showDetails {
					if highlighted != nil && m.selectedSession != nil && highlighted.ID == m.selectedSession.ID {
						m.showDetails = false
					} else {
						m.updateSelectedSession()
					}
				} else {
					m.showDetails = true
					m.updateSelectedSession()
				}

				m.updatePanelSizes()
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+s"))):
			return m, m.toggleProxyCmd()

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+r"))):
			return m, m.refreshSessionsCmd()

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+d"))):
			m.clearSessions()

		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			var session *sessiondata.Session

			if m.activePanel == SessionPanel {
				session = m.sessionsPanel.GetSelectedSession()
			} else if m.showDetails && m.selectedSession != nil {
				session = m.selectedSession
			}

			if session == nil {
				m.errorMsg = "No session selected"
				return m, clearStatusCmd()
			}

			m.statusMsg = fmt.Sprintf("Replaying %s %s...", session.Request.Method, session.Request.URL)
			if m.logger != nil {
				m.logger.LogInfo(fmt.Sprintf("Replay: %s %s -> :%d", session.Request.Method, session.Request.URL, m.port))
			}

			return m, m.replayCmd(session)

		case key.Matches(msg, key.NewBinding(key.WithKeys("escape"))):
			if m.showDetails {
				m.showDetails = false
				m.activePanel = SessionPanel
				m.updatePanelSizes()
			} else if m.filterRegex != "" {
				m.filterRegex = ""
				m.searchInput.SetValue("")
				m.applyFilter()
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
			if m.activePanel == SessionPanel {
				highlighted := m.sessionsPanel.GetSelectedSession()
				if highlighted != nil {
					err := clipboard.WriteAll(highlighted.ToCurl())
					if err != nil {
						m.errorMsg = "Error " + err.Error()
					} else {
						m.statusMsg = "cURL copied to clipboard"
					}
				} else {
					m.errorMsg = "No session selected"
				}
			}
			return m, clearStatusCmd()

		case key.Matches(msg, key.NewBinding(key.WithKeys("f1"))):
			m.showHelp = !m.showHelp

		case key.Matches(msg, key.NewBinding(key.WithKeys("f2"))):
			m.verbose = !m.verbose
			if m.logger != nil {
				m.logger.verbose = m.verbose
			}
		}
	}

	cmds = append(cmds, m.propagateToActivePanel(msg))
	return m, tea.Batch(cmds...)
}

func (m *Model) propagateToActivePanel(msg tea.Msg) tea.Cmd {
	switch m.activePanel {
	case SessionPanel:
		return m.sessionsPanel.Update(msg)
	default:
		if m.selectedSession != nil && m.selectedSession.Type == sessiondata.WebSocketSession {
			return m.websocketPanel.Update(msg)
		}
		switch m.activeTab {
		case 0:
			return m.requestPanel.Update(msg)
		case 1:
			return m.responsePanel.Update(msg)
		case 2:
			if m.tlsPanel != nil {
				return m.tlsPanel.Update(msg)
			}
		}
	}
	return nil
}

func (m *Model) updatePanelsForSession(session *sessiondata.Session) {
	if session.Type == sessiondata.WebSocketSession {
		m.websocketPanel.UpdateSession(session)
	} else {
		m.requestPanel.UpdateSession(session)
		m.responsePanel.UpdateSession(session)
	}
	if m.tlsPanel != nil {
		m.tlsPanel.UpdateSession(session)
	}
}

func (m *Model) switchPanel() {
	if !m.showDetails {
		return
	}
	if m.activePanel == SessionPanel {
		m.activePanel = RequestPanel
	} else {
		m.activePanel = SessionPanel
	}
}

func (m *Model) updateSelectedSession() {
	session := m.sessionsPanel.GetSelectedSession()
	if session == nil {
		return
	}
	m.selectedSession = session
	m.updatePanelsForSession(session)
	if session.Type == sessiondata.WebSocketSession {
		m.statusMsg = fmt.Sprintf("Selected WebSocket: %s (%d messages)",
			session.Request.URL, session.WebSocket.MessageCount)
	} else {
		m.statusMsg = fmt.Sprintf("Selected session: %s %s", session.Request.Method, session.Request.URL)
	}
}

func (m *Model) clearSessions() {
	if m.sessionStore != nil {
		m.sessionStore.Clear()
		m.showDetails = false
		m.sessions = []*sessiondata.Session{}
		m.sessionsPanel.UpdateSessions(m.sessions)
		m.selectedSession = nil
		m.requestPanel.UpdateSession(nil)
		m.responsePanel.UpdateSession(nil)
		if m.tlsPanel != nil {
			m.tlsPanel.UpdateSession(nil)
		}
		if m.websocketPanel != nil {
			m.websocketPanel.UpdateSession(nil)
		}
		m.statusMsg = "Sessions cleared"
	}
}

func (m *Model) resetSelection() {
	m.selectedSession = nil
	m.requestPanel.UpdateSession(nil)
	m.responsePanel.UpdateSession(nil)
	if m.tlsPanel != nil {
		m.tlsPanel.UpdateSession(nil)
	}
	if m.websocketPanel != nil {
		m.websocketPanel.UpdateSession(nil)
	}
	m.statusMsg = "Selection reset"
}

func (m *Model) updatePanelSizes() {
	if m.width == 0 || m.height == 0 {
		return
	}

	availW := m.width
	availH := m.height - 2

	safe := func(v int) int {
		if v < 0 {
			return 0
		}
		return v
	}

	if !m.showDetails {
		m.sessionsPanel.SetSize(safe(availW-2), safe(availH-4))
		return
	}

	sessionW := availW / 3
	detailsW := availW - sessionW

	m.sessionsPanel.SetSize(safe(sessionW-2), safe(availH-4))

	if m.requestPanel != nil {
		m.requestPanel.SetSize(safe(detailsW-2), safe(availH-4))
	}
	if m.responsePanel != nil {
		m.responsePanel.SetSize(safe(detailsW-2), safe(availH-4))
	}
	if m.tlsPanel != nil {
		m.tlsPanel.SetSize(safe(detailsW-2), safe(availH-4))
	}
	if m.websocketPanel != nil {
		m.websocketPanel.SetSize(safe(detailsW-2), safe(availH-4))
	}
}

func (m *Model) applyFilter() {
	if m.filterRegex == "" {
		m.errorMsg = ""
		m.compiledFilter = nil
		m.sessionsPanel.UpdateSessions(m.sessions)
		return
	}

	if m.compiledFilter == nil || m.compiledFilter.String() != m.filterRegex {
		re, err := regexp.Compile(m.filterRegex)
		if err != nil {
			m.errorMsg = "Invalid Regex: " + err.Error()
			m.compiledFilter = nil
			m.sessionsPanel.UpdateSessions(m.sessions)
			return
		}
		m.compiledFilter = re
	}

	m.errorMsg = ""
	var filtered []*sessiondata.Session
	for _, s := range m.sessions {
		if m.compiledFilter.MatchString(s.Request.URL) {
			filtered = append(filtered, s)
		}
	}
	m.sessionsPanel.UpdateSessions(filtered)
}

func (m *Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		if m.sessionStore != nil && m.sessionStore.SessionCount() != m.sessionCount {
			return SessionsUpdatedMsg{Sessions: m.sessionStore.GetAll()}
		}
		return TickMsg{}
	})
}

func (m *Model) toggleProxyCmd() tea.Cmd {
	if m.isRunning {
		server := m.server
		return func() tea.Msg {
			if server != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := server.Shutdown(ctx); err != nil {
					return ProxyStatusMsg{Running: true, Error: err}
				}
			}
			return ProxyStatusMsg{Running: false, Error: nil}
		}
	}

	if m.proxy == nil {
		if err := os.MkdirAll("certs", 0o755); err != nil {
			m.errorMsg = fmt.Sprintf("Failed to create certs dir: %v", err)
			return nil
		}

		caCache := certs.NewCertCache()
		err := caCache.LoadOrGenerateCA("certs", "certs/httpCA.crt", "certs/httpCA.key")
		if err != nil {
			m.errorMsg = fmt.Sprintf("CA error: %v", err)
			return nil
		}

		m.proxy = proxy.NewProxy(m.sessionStore, m.logger, caCache)
	}

	m.server = &http.Server{
		Addr:     fmt.Sprintf(":%d", m.port),
		Handler:  m.proxy,
		ErrorLog: log.New(io.Discard, "", 0),
	}

	server := m.server
	logger := m.logger
	port := m.port

	return func() tea.Msg {
		if logger != nil {
			logger.LogInfo(fmt.Sprintf("Proxy listening on http://127.0.0.1:%d", port))
		}
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				if logger != nil {
					logger.LogError(err, "Server error")
				}
			}
		}()
		return ProxyStatusMsg{Running: true, Error: nil}
	}
}

func (m *Model) refreshSessionsCmd() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if m.sessionStore != nil {
			return SessionsUpdatedMsg{Sessions: m.sessionStore.GetAll()}
		}
		return nil
	})
}

type ClearStatusMsg struct{}

func clearStatusCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(3 * time.Second)
		return ClearStatusMsg{}
	}
}

type ReplayResultMsg struct {
	Error error
}

func (m *Model) replayCmd(session *sessiondata.Session) tea.Cmd {
	port := m.port
	return func() tea.Msg {
		err := session.Replay(port)
		return ReplayResultMsg{Error: err}
	}
}
