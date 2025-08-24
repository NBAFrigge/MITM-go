package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"httpDebugger/internal/certs"
	"httpDebugger/internal/proxy"
	"httpDebugger/internal/session"
	"httpDebugger/internal/sessiondata"
)

type Logger struct {
	verbose bool
}

func (l *Logger) LogRequest(session *sessiondata.Session) {
	if l.verbose {
		fmt.Printf("[REQUEST] %s %s\n", session.Request.Method, session.Request.URL)
	}
}

func (l *Logger) LogResponse(session *sessiondata.Session) {
	if l.verbose {
		fmt.Printf("[RESPONSE] %d %s\n", session.Response.StatusCode, session.Request.URL)
	}
}

func (l *Logger) LogError(err error, context string) {
	if err != nil {
		fmt.Printf("[ERROR] %s: %v\n", context, err)
	} else {
		fmt.Printf("[INFO] %s\n", context)
	}
}

func (l *Logger) LogInfo(message string) {
	fmt.Printf("[INFO] %s\n", message)
}

type App struct {
	ctx          context.Context
	sessionStore *session.InMemoryStore
	proxy        *proxy.Proxy
	server       *http.Server
	mu           sync.RWMutex
	isRunning    bool
	port         int
}

func NewApp() *App {
	return &App{
		port: 8080,
	}
}

func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx

	a.sessionStore = session.NewInMemoryStore(1000)
}

func (a *App) OnShutdown(ctx context.Context) {
	if a.isRunning {
		a.StopProxy()
	}
}

func (a *App) StartProxy(port int, verbose bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isRunning {
		return fmt.Errorf("proxy is already running")
	}

	a.port = port

	logger := &Logger{verbose: verbose}

	caCache := certs.NewCertCache()

	err := caCache.LoadOrGenerateCA("certs", "certs/httpCA.crt", "certs/httpCA.key")
	if err != nil {
		return fmt.Errorf("failed to load or generate CA certificates: %v", err)
	}

	a.proxy = proxy.NewProxy(a.sessionStore, logger, caCache)

	a.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: a.proxy,
	}

	go func() {
		logger.LogError(nil, fmt.Sprintf("Proxy listening on http://localhost:%d", port))
		logger.LogError(nil, "Configure your client to use this proxy for HTTP and HTTPS requests")
		if err := a.server.ListenAndServe(); err != http.ErrServerClosed {
			logger.LogError(err, "Server error")
		}
	}()

	a.isRunning = true
	return nil
}

func (a *App) StopProxy() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isRunning {
		return fmt.Errorf("proxy is not running")
	}

	if a.server != nil {
		err := a.server.Close()
		if err != nil {
			return err
		}
	}

	a.isRunning = false
	fmt.Println("Proxy stopped")
	return nil
}

func (a *App) GetProxyStatus() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return map[string]interface{}{
		"running": a.isRunning,
		"port":    a.port,
	}
}

func (a *App) GetSessions() []map[string]interface{} {
	sessions := a.sessionStore.GetAll()
	result := make([]map[string]interface{}, len(sessions))

	for i, session := range sessions {
		result[i] = map[string]interface{}{
			"id":        session.ID,
			"method":    session.Request.Method,
			"url":       session.Request.URL,
			"timestamp": session.Timestamp,
			"duration":  session.Duration.Milliseconds(),
		}

		if session.Type == sessiondata.HTTPSession {
			result[i]["type"] = "HTTPSession"

			if session.Response != nil {
				result[i]["status"] = session.Response.StatusCode
			} else {
				result[i]["status"] = 0
			}

			fmt.Printf("HTTP Session: %s - %s %s - Status: %v\n",
				session.ID, session.Request.Method, session.Request.URL, result[i]["status"])

		} else if session.Type == sessiondata.WebSocketSession {
			result[i]["type"] = "WebSocketSession"

			if session.Response != nil {
				result[i]["status"] = session.Response.StatusCode
			} else {
				result[i]["status"] = 0
			}

			if session.WebSocket != nil {
				result[i]["webSocketData"] = map[string]interface{}{
					"state": session.WebSocket.State,
					"messageStats": map[string]interface{}{
						"totalMessages":    session.WebSocket.MessageCount.TotalMessages,
						"inboundMessages":  session.WebSocket.MessageCount.InboundMessages,
						"outboundMessages": session.WebSocket.MessageCount.OutboundMessages,
						"textMessages":     session.WebSocket.MessageCount.TextMessages,
						"binaryMessages":   session.WebSocket.MessageCount.BinaryMessages,
						"inboundBytes":     session.WebSocket.MessageCount.InboundBytes,
						"outboundBytes":    session.WebSocket.MessageCount.OutboundBytes,
						"totalBytes":       session.WebSocket.MessageCount.TotalBytes,
						"controlFrames":    session.WebSocket.MessageCount.ControlFrames,
					},
					"connectedAt":        session.WebSocket.ConnectedAt,
					"disconnectedAt":     session.WebSocket.DisconnectedAt,
					"connectionDuration": session.WebSocket.ConnectionDuration.Milliseconds(),
					"subprotocol":        session.WebSocket.Subprotocol,
					"extensions":         session.WebSocket.Extensions,
					"closeCode":          session.WebSocket.CloseCode,
					"closeReason":        session.WebSocket.CloseReason,
				}
			} else {
				result[i]["webSocketData"] = map[string]interface{}{
					"state": "unknown",
					"messageStats": map[string]interface{}{
						"totalMessages":    0,
						"inboundMessages":  0,
						"outboundMessages": 0,
						"textMessages":     0,
						"binaryMessages":   0,
						"inboundBytes":     0,
						"outboundBytes":    0,
						"totalBytes":       0,
						"controlFrames":    0,
					},
				}
			}
		} else {
			result[i]["type"] = "Unknown"
			result[i]["status"] = 0
			fmt.Printf("Unknown Session Type: %s - %s %s\n",
				session.ID, session.Request.Method, session.Request.URL)
		}

		if session.Error != nil {
			result[i]["error"] = session.Error.Error()
		}
	}

	return result
}

func (a *App) GetSessionDetails(sessionID string) map[string]interface{} {
	session, err := a.sessionStore.Get(sessionID)
	if err != nil || session == nil {
		return nil
	}

	result := map[string]interface{}{
		"id":             session.ID,
		"method":         session.Request.Method,
		"url":            session.Request.URL,
		"timestamp":      session.Timestamp,
		"duration":       session.Duration.Milliseconds(),
		"protocol":       session.Protocol,
		"requestHeaders": session.Request.Headers,
		"requestCookies": session.Request.Cookies,
		"requestBody":    session.Request.Body,
		"tlsProfile":     TLSConfigToMap(session.TLSConfig),
	}

	if session.Type == sessiondata.HTTPSession {
		result["type"] = "HTTPSession"
	} else if session.Type == sessiondata.WebSocketSession {
		result["type"] = "WebSocketSession"

		if session.WebSocket != nil {
			connectionDuration := int64(0)
			if !session.WebSocket.ConnectedAt.IsZero() && !session.WebSocket.DisconnectedAt.IsZero() {
				connectionDuration = session.WebSocket.DisconnectedAt.Sub(session.WebSocket.ConnectedAt).Milliseconds()
			} else if !session.WebSocket.ConnectedAt.IsZero() {
				connectionDuration = time.Now().Sub(session.WebSocket.ConnectedAt).Milliseconds()
			}

			wsData := map[string]interface{}{
				"state":              session.WebSocket.State,
				"connectedAt":        session.WebSocket.ConnectedAt,
				"disconnectedAt":     session.WebSocket.DisconnectedAt,
				"connectionDuration": connectionDuration,
				"subprotocol":        session.WebSocket.Subprotocol,
				"extensions":         session.WebSocket.Extensions,
				"closeCode":          session.WebSocket.CloseCode,
				"closeReason":        session.WebSocket.CloseReason,
				"messageStats": map[string]interface{}{
					"totalMessages":    session.WebSocket.MessageCount.TotalMessages,
					"inboundMessages":  session.WebSocket.MessageCount.InboundMessages,
					"outboundMessages": session.WebSocket.MessageCount.OutboundMessages,
					"textMessages":     session.WebSocket.MessageCount.TextMessages,
					"binaryMessages":   session.WebSocket.MessageCount.BinaryMessages,
					"inboundBytes":     session.WebSocket.MessageCount.InboundBytes,
					"outboundBytes":    session.WebSocket.MessageCount.OutboundBytes,
					"totalBytes":       session.WebSocket.MessageCount.TotalBytes,
					"controlFrames":    session.WebSocket.MessageCount.ControlFrames,
				},
			}

			if session.WebSocket.UpgradeRequest != nil {
				wsData["upgradeRequest"] = map[string]interface{}{
					"method":      session.WebSocket.UpgradeRequest.Method,
					"url":         session.WebSocket.UpgradeRequest.URL,
					"headers":     session.WebSocket.UpgradeRequest.Headers,
					"cookies":     session.WebSocket.UpgradeRequest.Cookies,
					"body":        session.WebSocket.UpgradeRequest.Body,
					"contentType": session.WebSocket.UpgradeRequest.ContentType,
				}
			} else {
				wsData["upgradeRequest"] = map[string]interface{}{
					"method":      session.Request.Method,
					"url":         session.Request.URL,
					"headers":     session.Request.Headers,
					"cookies":     session.Request.Cookies,
					"body":        session.Request.Body,
					"contentType": session.Request.ContentType,
				}
			}

			if session.WebSocket.UpgradeResponse != nil {
				wsData["upgradeResponse"] = map[string]interface{}{
					"statusCode":  session.WebSocket.UpgradeResponse.StatusCode,
					"status":      session.WebSocket.UpgradeResponse.Status,
					"headers":     session.WebSocket.UpgradeResponse.Headers,
					"cookies":     session.WebSocket.UpgradeResponse.Cookies,
					"body":        session.WebSocket.UpgradeResponse.Body,
					"contentType": session.WebSocket.UpgradeResponse.ContentType,
				}
			} else if session.Response != nil {
				wsData["upgradeResponse"] = map[string]interface{}{
					"statusCode":  session.Response.StatusCode,
					"status":      session.Response.Status,
					"headers":     session.Response.Headers,
					"cookies":     session.Response.Cookies,
					"body":        session.Response.Body,
					"contentType": session.Response.ContentType,
				}
			} else {
				wsData["upgradeResponse"] = map[string]interface{}{
					"statusCode":  0,
					"status":      "No Response",
					"headers":     map[string]interface{}{},
					"cookies":     map[string]interface{}{},
					"body":        "",
					"contentType": "",
				}
			}

			messagesData, err := a.getAllWebSocketMessages(session)
			if err != nil {
				fmt.Printf("Error getting WebSocket messages for session %s: %v\n", sessionID, err)
				wsData["messages"] = map[string]interface{}{
					"messages": []interface{}{},
				}
			} else {
				wsData["messages"] = messagesData
			}

			result["webSocketData"] = wsData
		} else {
			result["webSocketData"] = map[string]interface{}{
				"state": "unknown",
				"messageStats": map[string]interface{}{
					"totalMessages":    0,
					"inboundMessages":  0,
					"outboundMessages": 0,
					"textMessages":     0,
					"binaryMessages":   0,
					"inboundBytes":     0,
					"outboundBytes":    0,
					"totalBytes":       0,
					"controlFrames":    0,
				},
				"messages": map[string]interface{}{
					"messages": []interface{}{},
				},
			}
		}
	} else {
		result["type"] = "Unknown"
	}

	if session.Response != nil {
		result["status"] = session.Response.StatusCode
		result["responseHeaders"] = session.Response.Headers
		result["responseBody"] = session.Response.Body
		result["responseCookies"] = session.Response.Cookies
	} else {
		result["status"] = 0
		result["responseHeaders"] = map[string]interface{}{}
		result["responseCookies"] = map[string]interface{}{}
		result["responseBody"] = ""
	}

	if session.Error != nil {
		result["error"] = session.Error.Error()
	}

	return result
}

func (a *App) ClearSessions() {
	a.sessionStore.Clear()
}

func (a *App) GetSessionCurl(sessionID string) string {
	session, err := a.sessionStore.Get(sessionID)
	if err != nil || session == nil {
		return ""
	}

	return session.ToCurl()
}

func (a *App) Replay(sessionID string) error {
	session, err := a.sessionStore.Get(sessionID)
	if err != nil || session == nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if session.Type != sessiondata.HTTPSession {
		return fmt.Errorf("replay only supports HTTP sessions, got: %s", session.Type)
	}

	err = session.Replay(a.port)
	if err != nil {
		return fmt.Errorf("failed to replay session %s: %v", sessionID, err)
	}

	return nil
}

func (a *App) SearchSessions(searchOptions map[string]interface{}) ([]map[string]interface{}, error) {
	opts := session.SearchOptions{}

	fmt.Printf("Search Options: %+v\n", searchOptions)

	if url, ok := searchOptions["URL"].(string); ok {
		opts.URL = url
	}
	if body, ok := searchOptions["Body"].(string); ok {
		opts.Body = body
	}
	if headersKey, ok := searchOptions["HeadersKey"].(string); ok {
		opts.HeadersKey = headersKey
	}
	if headersVal, ok := searchOptions["HeadersVal"]; ok {
		opts.HeadersVal = headersVal
	}
	if cookiesKey, ok := searchOptions["CookiesKey"].(string); ok {
		opts.CookiesKey = cookiesKey
	}
	if cookiesVal, ok := searchOptions["CookiesVal"]; ok {
		opts.CookiesVal = cookiesVal
	}

	results, err := a.sessionStore.Search(opts)
	if err != nil {
		return nil, err
	}

	var sessions []map[string]interface{}
	for _, session := range results {
		sessionMap := a.sessionToMap(session)
		sessions = append(sessions, sessionMap)
	}

	return sessions, nil
}

func (a *App) CompareSessions(sessionID1, sessionID2 string) (map[string]interface{}, error) {
	session1, err := a.sessionStore.Get(sessionID1)
	if err != nil || session1 == nil {
		return nil, fmt.Errorf("session not found: %s", sessionID1)
	}

	session2, err := a.sessionStore.Get(sessionID2)
	if err != nil || session2 == nil {
		return nil, fmt.Errorf("session not found: %s", sessionID2)
	}

	diff := session1.RequestDifferences(session2)

	result := map[string]interface{}{
		"session1": a.sessionToMap(session1),
		"session2": a.sessionToMap(session2),
	}
	if !diff.HasDiffs {
		result["differences"] = "No differences found"
		return result, nil
	}

	result["differences"] = getDifferences(diff)

	return result, nil
}

func (a *App) sessionToMap(session *sessiondata.Session) map[string]interface{} {
	return map[string]interface{}{
		"id":             session.ID,
		"method":         session.Request.Method,
		"url":            session.Request.URL,
		"status":         session.Response.StatusCode,
		"timestamp":      session.Timestamp,
		"duration":       session.Duration.Milliseconds(),
		"requestHeaders": session.Request.Headers.String(),
		"requestCookies": func() map[string]interface{} {
			cookies := make(map[string]interface{})
			for key, cookie := range session.Request.Cookies {
				cookies[key] = cookie
			}
			return cookies
		}(),
		"requestBody": func() string {
			if len(session.Request.Body) > 100 {
				return string(session.Request.Body[:100]) + "..."
			}
			return string(session.Request.Body)
		}(),
		"responseHeaders": session.Response.Headers.String(),
		"responseCookies": func() map[string]interface{} {
			if session.Response == nil {
				return map[string]interface{}{}
			}
			cookies := make(map[string]interface{})
			for key, cookie := range session.Request.Cookies {
				cookies[key] = cookie
			}
			return cookies
		}(),
		"responseBody": func() string {
			if session.Response == nil {
				return ""
			}
			if len(session.Response.Body) > 100 {
				return string(session.Response.Body[:100]) + "..."
			}
			return string(session.Response.Body)
		}(),
	}
}

func (a *App) getAllWebSocketMessages(session *sessiondata.Session) (map[string]interface{}, error) {
	if session == nil {
		return nil, fmt.Errorf("session is nil")
	}

	if session.WebSocket == nil {
		return map[string]interface{}{
			"messages": []interface{}{},
		}, nil
	}

	if session.WebSocket.Messages == nil {
		return map[string]interface{}{
			"messages": []interface{}{},
		}, nil
	}

	messages := make([]map[string]interface{}, len(session.WebSocket.Messages))
	for i, msg := range session.WebSocket.Messages {
		payloadStr := ""
		if msg.Payload != nil {
			payloadStr = string(msg.Payload)
		}

		payloadText := msg.PayloadText
		if payloadText == "" && msg.Type == sessiondata.TextMessage && len(msg.Payload) > 0 {
			payloadText = string(msg.Payload)
		}

		directionStr := "unknown"
		switch msg.Direction {
		case sessiondata.Inbound:
			directionStr = "inbound"
		case sessiondata.Outbound:
			directionStr = "outbound"
		}

		messages[i] = map[string]interface{}{
			"id":          msg.ID,
			"timestamp":   msg.Timestamp,
			"direction":   directionStr,
			"opcode":      msg.Opcode,
			"payload":     payloadStr,
			"isMasked":    msg.IsMasked,
			"isFragment":  msg.IsFragment,
			"size":        msg.Size,
			"payloadText": payloadText,
		}
	}

	return map[string]interface{}{
		"messages": messages,
	}, nil
}

func getDifferences(diff *sessiondata.RequestDifference) map[string]interface{} {
	diffMap := make(map[string]interface{}, 6)

	if !diff.HasDiffs {
		return nil
	}

	if diff.URL.Changed {
		diffMap["url"] = map[string]interface{}{
			"original": diff.URL.Original,
			"other":    diff.URL.Other,
		}
	}

	if diff.Method.Changed {
		diffMap["method"] = map[string]interface{}{
			"original": diff.Method.Original,
			"other":    diff.Method.Other,
		}
	}

	if diff.Body.Changed {
		diffMap["body"] = map[string]interface{}{
			"original": diff.Body.Original,
			"other":    diff.Body.Other,
		}
	}

	if diff.ContentType.Changed {
		diffMap["contentType"] = map[string]interface{}{
			"original": diff.ContentType.Original,
			"other":    diff.ContentType.Other,
		}
	}

	if diff.Headers.Changed {
		diffMap["headers"] = map[string]interface{}{
			"changed":  diff.Headers.Changed,
			"modified": diff.Headers.Modified,
			"added":    diff.Headers.Added,
			"removed":  diff.Headers.Removed,
		}
	}

	if diff.Cookies.Changed {
		diffMap["cookies"] = map[string]interface{}{
			"changed":  diff.Cookies.Changed,
			"modified": diff.Cookies.Modified,
			"added":    diff.Cookies.Added,
			"removed":  diff.Cookies.Removed,
		}
	}

	return diffMap
}

func TLSConfigToMap(config *tls.Config) map[string]interface{} {
	if config == nil {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})

	result["ServerName"] = config.ServerName
	result["InsecureSkipVerify"] = config.InsecureSkipVerify
	result["MinVersion"] = tlsVersionToString(config.MinVersion)
	result["MaxVersion"] = tlsVersionToString(config.MaxVersion)
	result["PreferServerCipherSuites"] = config.PreferServerCipherSuites
	result["SessionTicketsDisabled"] = config.SessionTicketsDisabled
	result["DynamicRecordSizingDisabled"] = config.DynamicRecordSizingDisabled
	result["Renegotiation"] = tlsRenegotiationToString(config.Renegotiation)

	if len(config.NextProtos) > 0 {
		result["NextProtos"] = config.NextProtos
	}

	if len(config.CipherSuites) > 0 {
		cipherNames := make([]string, len(config.CipherSuites))
		for i, cipher := range config.CipherSuites {
			cipherNames[i] = cipherSuiteToString(cipher)
		}
		result["CipherSuites"] = cipherNames
	}

	if len(config.CurvePreferences) > 0 {
		curveNames := make([]string, len(config.CurvePreferences))
		for i, curve := range config.CurvePreferences {
			curveNames[i] = curveToString(curve)
		}
		result["CurvePreferences"] = curveNames
	}

	if len(config.Certificates) > 0 {
		certs := make([]map[string]interface{}, len(config.Certificates))
		for i, cert := range config.Certificates {
			certs[i] = certificateToMap(cert)
		}
		result["Certificates"] = certs
	}

	if config.RootCAs != nil {
		result["RootCAs"] = certificatePoolToMap(config.RootCAs)
	}

	if config.ClientCAs != nil {
		result["ClientCAs"] = certificatePoolToMap(config.ClientCAs)
	}

	result["ClientAuth"] = clientAuthTypeToString(config.ClientAuth)

	return result
}

func tlsVersionToString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	case 0:
		return "Default"
	default:
		return fmt.Sprintf("Unknown (0x%04x)", version)
	}
}

func tlsRenegotiationToString(support tls.RenegotiationSupport) string {
	switch support {
	case tls.RenegotiateNever:
		return "Never"
	case tls.RenegotiateOnceAsClient:
		return "OnceAsClient"
	case tls.RenegotiateFreelyAsClient:
		return "FreelyAsClient"
	default:
		return fmt.Sprintf("Unknown (%d)", support)
	}
}

func clientAuthTypeToString(authType tls.ClientAuthType) string {
	switch authType {
	case tls.NoClientCert:
		return "NoClientCert"
	case tls.RequestClientCert:
		return "RequestClientCert"
	case tls.RequireAnyClientCert:
		return "RequireAnyClientCert"
	case tls.VerifyClientCertIfGiven:
		return "VerifyClientCertIfGiven"
	case tls.RequireAndVerifyClientCert:
		return "RequireAndVerifyClientCert"
	default:
		return fmt.Sprintf("Unknown (%d)", authType)
	}
}

func cipherSuiteToString(cipher uint16) string {
	suites := map[uint16]string{
		tls.TLS_RSA_WITH_RC4_128_SHA:                "TLS_RSA_WITH_RC4_128_SHA",
		tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA:           "TLS_RSA_WITH_3DES_EDE_CBC_SHA",
		tls.TLS_RSA_WITH_AES_128_CBC_SHA:            "TLS_RSA_WITH_AES_128_CBC_SHA",
		tls.TLS_RSA_WITH_AES_256_CBC_SHA:            "TLS_RSA_WITH_AES_256_CBC_SHA",
		tls.TLS_RSA_WITH_AES_128_CBC_SHA256:         "TLS_RSA_WITH_AES_128_CBC_SHA256",
		tls.TLS_RSA_WITH_AES_128_GCM_SHA256:         "TLS_RSA_WITH_AES_128_GCM_SHA256",
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384:         "TLS_RSA_WITH_AES_256_GCM_SHA384",
		tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA:        "TLS_ECDHE_ECDSA_WITH_RC4_128_SHA",
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA:    "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:    "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
		tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA:          "TLS_ECDHE_RSA_WITH_RC4_128_SHA",
		tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA:     "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA",
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA:      "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:      "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256: "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256",
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256:   "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256",
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:   "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256: "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:   "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384: "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305:    "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305:  "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
		tls.TLS_AES_128_GCM_SHA256:                  "TLS_AES_128_GCM_SHA256",
		tls.TLS_AES_256_GCM_SHA384:                  "TLS_AES_256_GCM_SHA384",
		tls.TLS_CHACHA20_POLY1305_SHA256:            "TLS_CHACHA20_POLY1305_SHA256",
	}

	if name, exists := suites[cipher]; exists {
		return name
	}
	return fmt.Sprintf("Unknown (0x%04x)", cipher)
}

func curveToString(curve tls.CurveID) string {
	curves := map[tls.CurveID]string{
		tls.CurveP256: "P-256",
		tls.CurveP384: "P-384",
		tls.CurveP521: "P-521",
		tls.X25519:    "X25519",
	}

	if name, exists := curves[curve]; exists {
		return name
	}
	return fmt.Sprintf("Unknown (%d)", curve)
}

func signatureSchemeToString(scheme tls.SignatureScheme) string {
	schemes := map[tls.SignatureScheme]string{
		tls.PKCS1WithSHA256:        "PKCS1WithSHA256",
		tls.PKCS1WithSHA384:        "PKCS1WithSHA384",
		tls.PKCS1WithSHA512:        "PKCS1WithSHA512",
		tls.PSSWithSHA256:          "PSSWithSHA256",
		tls.PSSWithSHA384:          "PSSWithSHA384",
		tls.PSSWithSHA512:          "PSSWithSHA512",
		tls.ECDSAWithP256AndSHA256: "ECDSAWithP256AndSHA256",
		tls.ECDSAWithP384AndSHA384: "ECDSAWithP384AndSHA384",
		tls.ECDSAWithP521AndSHA512: "ECDSAWithP521AndSHA512",
		tls.Ed25519:                "Ed25519",
		tls.PKCS1WithSHA1:          "PKCS1WithSHA1",
		tls.ECDSAWithSHA1:          "ECDSAWithSHA1",
	}

	if name, exists := schemes[scheme]; exists {
		return name
	}
	return fmt.Sprintf("Unknown (0x%04x)", scheme)
}

func certificateToMap(cert tls.Certificate) map[string]interface{} {
	result := make(map[string]interface{})

	if len(cert.Certificate) > 0 {
		x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
		if err == nil {
			result["Subject"] = x509Cert.Subject.String()
			result["Issuer"] = x509Cert.Issuer.String()
			result["SerialNumber"] = x509Cert.SerialNumber.String()
			result["NotBefore"] = x509Cert.NotBefore.Format(time.RFC3339)
			result["NotAfter"] = x509Cert.NotAfter.Format(time.RFC3339)
			result["DNSNames"] = x509Cert.DNSNames
			result["EmailAddresses"] = x509Cert.EmailAddresses
			result["IPAddresses"] = formatIPAddresses(x509Cert.IPAddresses)
			result["KeyUsage"] = keyUsageToString(x509Cert.KeyUsage)
			result["ExtKeyUsage"] = extKeyUsageToStrings(x509Cert.ExtKeyUsage)
			result["IsCA"] = x509Cert.IsCA
			result["Version"] = x509Cert.Version
			result["SignatureAlgorithm"] = x509Cert.SignatureAlgorithm.String()
			result["PublicKeyAlgorithm"] = x509Cert.PublicKeyAlgorithm.String()
		}
	}

	result["CertificateCount"] = len(cert.Certificate)

	if cert.PrivateKey != nil {
		result["HasPrivateKey"] = true
		result["PrivateKeyType"] = fmt.Sprintf("%T", cert.PrivateKey)
	} else {
		result["HasPrivateKey"] = false
	}

	if len(cert.SupportedSignatureAlgorithms) > 0 {
		schemes := make([]string, len(cert.SupportedSignatureAlgorithms))
		for i, scheme := range cert.SupportedSignatureAlgorithms {
			schemes[i] = signatureSchemeToString(scheme)
		}
		result["SupportedSignatureAlgorithms"] = schemes
	}

	if cert.OCSPStaple != nil {
		result["HasOCSPStaple"] = true
		result["OCSPStapleLength"] = len(cert.OCSPStaple)
	}

	if len(cert.SignedCertificateTimestamps) > 0 {
		result["SignedCertificateTimestamps"] = len(cert.SignedCertificateTimestamps)
	}

	return result
}

func certificatePoolToMap(pool *x509.CertPool) map[string]interface{} {
	if pool == nil {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})
	subjects := pool.Subjects()
	result["CertCount"] = len(subjects)

	if len(subjects) > 0 {
		subjectStrings := make([]string, len(subjects))
		for i, subject := range subjects {
			subjectStrings[i] = fmt.Sprintf("%x", subject)
		}
		result["Subjects"] = subjectStrings
	}

	return result
}

func formatIPAddresses(ips []net.IP) []string {
	result := make([]string, len(ips))
	for i, ip := range ips {
		result[i] = ip.String()
	}
	return result
}

func keyUsageToString(usage x509.KeyUsage) []string {
	var usages []string

	if usage&x509.KeyUsageDigitalSignature != 0 {
		usages = append(usages, "DigitalSignature")
	}
	if usage&x509.KeyUsageContentCommitment != 0 {
		usages = append(usages, "ContentCommitment")
	}
	if usage&x509.KeyUsageKeyEncipherment != 0 {
		usages = append(usages, "KeyEncipherment")
	}
	if usage&x509.KeyUsageDataEncipherment != 0 {
		usages = append(usages, "DataEncipherment")
	}
	if usage&x509.KeyUsageKeyAgreement != 0 {
		usages = append(usages, "KeyAgreement")
	}
	if usage&x509.KeyUsageCertSign != 0 {
		usages = append(usages, "CertSign")
	}
	if usage&x509.KeyUsageCRLSign != 0 {
		usages = append(usages, "CRLSign")
	}
	if usage&x509.KeyUsageEncipherOnly != 0 {
		usages = append(usages, "EncipherOnly")
	}
	if usage&x509.KeyUsageDecipherOnly != 0 {
		usages = append(usages, "DecipherOnly")
	}

	return usages
}

func extKeyUsageToStrings(usages []x509.ExtKeyUsage) []string {
	result := make([]string, len(usages))

	for i, usage := range usages {
		switch usage {
		case x509.ExtKeyUsageAny:
			result[i] = "Any"
		case x509.ExtKeyUsageServerAuth:
			result[i] = "ServerAuth"
		case x509.ExtKeyUsageClientAuth:
			result[i] = "ClientAuth"
		case x509.ExtKeyUsageCodeSigning:
			result[i] = "CodeSigning"
		case x509.ExtKeyUsageEmailProtection:
			result[i] = "EmailProtection"
		case x509.ExtKeyUsageIPSECEndSystem:
			result[i] = "IPSECEndSystem"
		case x509.ExtKeyUsageIPSECTunnel:
			result[i] = "IPSECTunnel"
		case x509.ExtKeyUsageIPSECUser:
			result[i] = "IPSECUser"
		case x509.ExtKeyUsageTimeStamping:
			result[i] = "TimeStamping"
		case x509.ExtKeyUsageOCSPSigning:
			result[i] = "OCSPSigning"
		case x509.ExtKeyUsageMicrosoftServerGatedCrypto:
			result[i] = "MicrosoftServerGatedCrypto"
		case x509.ExtKeyUsageNetscapeServerGatedCrypto:
			result[i] = "NetscapeServerGatedCrypto"
		case x509.ExtKeyUsageMicrosoftCommercialCodeSigning:
			result[i] = "MicrosoftCommercialCodeSigning"
		case x509.ExtKeyUsageMicrosoftKernelCodeSigning:
			result[i] = "MicrosoftKernelCodeSigning"
		default:
			result[i] = fmt.Sprintf("Unknown (%d)", usage)
		}
	}

	return result
}
