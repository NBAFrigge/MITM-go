package proxy

import (
	"crypto/tls"
	"httpDebugger/internal/certs"
	"net/http"
	"time"

	"httpDebugger/internal/proxy/handlers"
	"httpDebugger/internal/proxy/interfaces"
	"httpDebugger/internal/proxy/types"
)

type Proxy struct {
	config     *types.Config
	certsCache *certs.CertCache
	handlers   *handlers.Manager
}

func NewProxy(store interfaces.SessionStore, logger interfaces.Logger, caCache *certs.CertCache) *Proxy {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	config := &types.Config{
		SessionStore: store,
		Logger:       logger,
		HTTPClient:   client,
		CACert:       caCache.CACert,
	}

	return &Proxy{
		config:     config,
		handlers:   handlers.NewManager(config, caCache),
		certsCache: caCache,
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.handlers.HandleMITM(w, r)
		return
	}
	p.handlers.HandleHTTP(w, r)
}
