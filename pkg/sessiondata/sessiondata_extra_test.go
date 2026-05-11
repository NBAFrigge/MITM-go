package sessiondata

import (
	"net/http"
	"net/url"
	"testing"

	"httpDebugger/pkg/sortedMap"
)

func TestNewSessionData_URLPortCleaning(t *testing.T) {
	tests := []struct {
		rawURL   string
		expected string
	}{
		{"https://example.com:443/path", "https://example.com/path"},
		{"http://example.com:80/path", "http://example.com/path"},
		{"https://example.com:8443/path", "https://example.com:8443/path"},
		{"https://example.com/path", "https://example.com/path"},
	}

	for _, tt := range tests {
		t.Run(tt.rawURL, func(t *testing.T) {
			parsedURL, _ := url.Parse(tt.rawURL)
			req := &http.Request{
				Method: "GET",
				URL:    parsedURL,
				Header: http.Header{},
			}
			session := NewSessionData(req, nil, sortedMap.New(), nil, HTTP11Protocol)
			if session.Request.URL != tt.expected {
				t.Errorf("URL = %q, want %q", session.Request.URL, tt.expected)
			}
		})
	}
}

func TestNewSessionData_WebSocketUpgrade(t *testing.T) {
	parsedURL, _ := url.Parse("https://example.com/ws")
	req := &http.Request{
		Method: "GET",
		URL:    parsedURL,
		Header: http.Header{
			"Upgrade":               {"websocket"},
			"Sec-Websocket-Version": {"13"},
			"Sec-Websocket-Key":     {"dGhlIHNhbXBsZSBub25jZQ=="},
		},
	}

	session := NewSessionData(req, nil, sortedMap.New(), nil, HTTP11Protocol)

	if session.Type != WebSocketSession {
		t.Errorf("expected WebSocketSession, got %v", session.Type)
	}
	if session.WebSocket == nil {
		t.Fatal("WebSocket data should be set")
	}
	if session.WebSocket.State != WSConnecting {
		t.Errorf("expected WSConnecting, got %v", session.WebSocket.State)
	}
	if session.WebSocket.UpgradeRequest == nil {
		t.Error("UpgradeRequest should be set")
	}
}

func TestNewSessionData_NonWebSocket(t *testing.T) {
	parsedURL, _ := url.Parse("https://example.com/api")
	req := &http.Request{
		Method: "POST",
		URL:    parsedURL,
		Header: http.Header{"Content-Type": {"application/json"}},
	}

	session := NewSessionData(req, []byte(`{"key":"val"}`), sortedMap.New(), nil, HTTP11Protocol)

	if session.Type != HTTPSession {
		t.Errorf("expected HTTPSession, got %v", session.Type)
	}
	if session.WebSocket != nil {
		t.Error("WebSocket should be nil for HTTP sessions")
	}
	if session.Request.Body != `{"key":"val"}` {
		t.Errorf("body = %q, want {\"key\":\"val\"}", session.Request.Body)
	}
}

func TestNewSessionData_CookieExtraction(t *testing.T) {
	parsedURL, _ := url.Parse("https://example.com/")
	req := &http.Request{
		Method: "GET",
		URL:    parsedURL,
		Header: http.Header{},
	}
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc"})
	req.AddCookie(&http.Cookie{Name: "user", Value: "bob"})

	session := NewSessionData(req, nil, sortedMap.New(), nil, HTTP11Protocol)

	if session.Request.Cookies["session"] != "abc" {
		t.Errorf("session cookie = %q, want abc", session.Request.Cookies["session"])
	}
	if session.Request.Cookies["user"] != "bob" {
		t.Errorf("user cookie = %q, want bob", session.Request.Cookies["user"])
	}
}

func TestNewSessionData_IDIsUnique(t *testing.T) {
	parsedURL, _ := url.Parse("https://example.com/")
	req := &http.Request{
		Method: "GET",
		URL:    parsedURL,
		Header: http.Header{},
	}

	s1 := NewSessionData(req, nil, sortedMap.New(), nil, HTTP11Protocol)
	s2 := NewSessionData(req, nil, sortedMap.New(), nil, HTTP11Protocol)

	if s1.ID == s2.ID {
		t.Error("sessions should have unique IDs")
	}
	if s1.ID == "" || s2.ID == "" {
		t.Error("session IDs should not be empty")
	}
}

func TestToCurl_WebSocketReturnsMessage(t *testing.T) {
	parsedURL, _ := url.Parse("https://example.com/ws")
	req := &http.Request{
		Method: "GET",
		URL:    parsedURL,
		Header: http.Header{
			"Upgrade":               {"websocket"},
			"Sec-Websocket-Version": {"13"},
			"Sec-Websocket-Key":     {"dGhlIHNhbXBsZSBub25jZQ=="},
		},
	}
	session := NewSessionData(req, nil, sortedMap.New(), nil, HTTP11Protocol)

	result := session.ToCurl()
	if result == "" {
		t.Error("ToCurl for WebSocket should return a non-empty message")
	}
}
