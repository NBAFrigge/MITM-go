package types

import (
	"crypto/tls"
	"net/http"
	"sync"

	"httpDebugger/internal/proxy/interfaces"
)

type Config struct {
	SessionStore interfaces.SessionStore
	Logger       interfaces.Logger
	HTTPClient   *http.Client
	CACert       tls.Certificate
	Mutex        sync.Mutex
}
