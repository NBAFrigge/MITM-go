package tui

import (
	"fmt"
	"net/http"

	"httpDebugger/pkg/certs"
	"httpDebugger/pkg/proxy"
)

func (m *Model) StartProxy(port int, verbose bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRunning {
		return fmt.Errorf("proxy is already running")
	}

	m.port = port
	m.verbose = verbose

	logger, err := NewLogger(verbose)
	if err != nil {
		return fmt.Errorf("failed to create logger: %v", err)
	}
	m.logger = logger

	caCache := certs.NewCertCache()
	err = caCache.LoadOrGenerateCA("certs", "certs/httpCA.crt", "certs/httpCA.key")
	if err != nil {
		m.logger.LogError(err, "Failed to load or generate CA certificates")
		return fmt.Errorf("failed to load or generate CA certificates: %v", err)
	}

	m.proxy = proxy.NewProxy(m.sessionStore, m.logger, caCache)

	m.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: m.proxy,
	}

	go func() {
		if m.logger != nil {
			m.logger.LogInfo(fmt.Sprintf("Proxy listening on http://127.0.0.1:%d", m.port))
			m.logger.LogInfo("Configure your client to use this proxy for HTTP and HTTPS requests")
			m.logger.LogInfo(fmt.Sprintf("Log file: %s", m.logger.GetLogFilePath()))
		}

		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			if m.logger != nil {
				m.logger.LogError(err, "Server error")
			}
		}
	}()
	m.isRunning = true
	return nil
}

func (m *Model) StopProxy() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isRunning {
		return fmt.Errorf("proxy is not running")
	}

	if m.server != nil {
		err := m.server.Close()
		if err != nil {
			if m.logger != nil {
				m.logger.LogError(err, "Error stopping server")
			}
			return err
		}
	}

	if m.logger != nil {
		m.logger.LogInfo("Proxy stopped")
		m.logger.Close()
	}

	m.isRunning = false
	return nil
}
