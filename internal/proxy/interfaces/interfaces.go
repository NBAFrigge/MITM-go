package interfaces

import "httpDebugger/internal/sessiondata"

type SessionStore interface {
	Store(session *sessiondata.Session) error
	GetAll() []*sessiondata.Session
	Get(id string) (*sessiondata.Session, error)
}

type Logger interface {
	LogRequest(session *sessiondata.Session)
	LogResponse(session *sessiondata.Session)
	LogError(err error, context string)
	LogInfo(message string)
}
