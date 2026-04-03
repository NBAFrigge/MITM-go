package tui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"httpDebugger/pkg/sessiondata"
)

type Logger struct {
	verbose bool
	logFile *os.File
	logger  *log.Logger
	mu      sync.Mutex
}

func NewLogger(verbose bool) (*Logger, error) {
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %v", err)
	}

	now := time.Now()
	filename := fmt.Sprintf("proxy_%s.log", now.Format("2006-01-02_15-04-05"))
	logPath := filepath.Join(logsDir, filename)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %v", err)
	}

	log.SetOutput(logFile)

	logger := log.New(logFile, "", log.LstdFlags|log.Lmicroseconds)

	l := &Logger{
		verbose: verbose,
		logFile: logFile,
		logger:  logger,
	}

	l.logger.Printf("=== HTTP Debugger Proxy Started ===")
	l.logger.Printf("Log file: %s", logPath)
	l.logger.Printf("Verbose mode: %t", verbose)
	l.logger.Printf("=====================================")

	return l, nil
}

func (l *Logger) LogRequest(session *sessiondata.Session) {
	if !l.verbose || session == nil || session.Request == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logger != nil {
		l.logger.Printf("[REQUEST] %s %s", session.Request.Method, session.Request.URL)

		if session.Request.Headers != nil && len(session.Request.Headers.Entries) > 0 {
			l.logger.Printf("[REQUEST_HEADERS] %s %s - Headers: %d",
				session.Request.Method, session.Request.URL, len(session.Request.Headers.Entries))
		}

		if len(session.Request.Body) > 0 {
			l.logger.Printf("[REQUEST_BODY] %s %s - Body size: %d bytes",
				session.Request.Method, session.Request.URL, len(session.Request.Body))
		}
	}
}

func (l *Logger) LogResponse(session *sessiondata.Session) {
	if !l.verbose || session == nil || session.Response == nil || session.Request == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logger != nil {
		l.logger.Printf("[RESPONSE] %d %s - Duration: %v",
			session.Response.StatusCode, session.Request.URL, session.Duration)

		if len(session.Response.Body) > 0 {
			l.logger.Printf("[RESPONSE_BODY] %d %s - Body size: %d bytes",
				session.Response.StatusCode, session.Request.URL, len(session.Response.Body))
		}
	}
}

func (l *Logger) LogError(err error, context string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logger != nil {
		if err != nil {
			l.logger.Printf("[ERROR] %s: %v", context, err)
		} else {
			l.logger.Printf("[INFO] %s", context)
		}
	}
}

func (l *Logger) LogInfo(message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logger != nil {
		l.logger.Printf("[INFO] %s", message)
	}
}

func (l *Logger) LogSession(session *sessiondata.Session) {
	if !l.verbose || session == nil || session.Request == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logger != nil {
		l.logger.Printf("[SESSION] ID: %s | %s %s | Duration: %v",
			session.ID, session.Request.Method, session.Request.URL, session.Duration)

		if session.Response != nil {
			l.logger.Printf("[SESSION] ID: %s | Response: %d %s",
				session.ID, session.Response.StatusCode, session.Response.Status)
		}

		if session.Error != nil {
			l.logger.Printf("[SESSION] ID: %s | Error: %v", session.ID, session.Error)
		}
	}
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logger != nil {
		l.logger.Printf("=== HTTP Debugger Proxy Stopped ===")
	}

	if l.logFile != nil {
		return l.logFile.Close()
	}

	return nil
}

func (l *Logger) GetLogFilePath() string {
	if l.logFile != nil {
		return l.logFile.Name()
	}
	return ""
}
