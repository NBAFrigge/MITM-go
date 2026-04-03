package sessiondata

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"httpDebugger/pkg/clientHello"
	"httpDebugger/pkg/sortedMap"

	"github.com/google/uuid"
	"golang.org/x/net/http2"
)

type Session struct {
	ID             string
	Timestamp      time.Time
	TLSFingerprint *clientHello.TLSFingerprint
	Request        *RequestData
	Response       *ResponseData
	Duration       time.Duration
	Error          error
	Protocol       string
	Type           SessionType
	WebSocket      *WebSocketData
}

func NewSessionData(r *http.Request, bodyBytes []byte, headers *sortedMap.SortedMap, tlsFingerprint *clientHello.TLSFingerprint, protocol string) *Session {
	reqCookies := r.Cookies()
	cookies := make(map[string]string, len(reqCookies))
	for _, cookie := range reqCookies {
		cookies[cookie.Name] = cookie.Value
	}

	requestData := &RequestData{
		Method:      r.Method,
		URL:         r.URL.String(),
		Headers:     headers,
		Body:        string(bodyBytes),
		Cookies:     cookies,
		ContentType: r.Header.Get("Content-Type"),
	}

	if isWsUpgradeRequest(r) {
		return &Session{
			ID:             uuid.New().String(),
			Timestamp:      time.Now(),
			Request:        requestData,
			Type:           WebSocketSession,
			TLSFingerprint: tlsFingerprint,
			Protocol:       protocol,
			WebSocket: &WebSocketData{
				State:          WSConnecting,
				UpgradeRequest: requestData,
				ConnectedAt:    time.Now(),
				Messages:       make([]WebSocketMessage, 0),
				MessageCount:   newMessageStats(),
				Subprotocol:    r.Header.Get("Sec-WebSocket-Protocol"),
				Extensions:     r.Header["Sec-WebSocket-Extensions"],
				CloseCode:      0,
				CloseReason:    "",
			},
		}
	} else {
		return &Session{
			ID:             uuid.New().String(),
			Timestamp:      time.Now(),
			Request:        requestData,
			Type:           HTTPSession,
			TLSFingerprint: tlsFingerprint,
			Protocol:       protocol,
			WebSocket:      nil,
			Response:       nil,
		}
	}
}

func (s *Session) CompareRequest(other *Session) bool {
	if s.Request.Method != other.Request.Method {
		return false
	}
	if s.Request.URL != other.Request.URL {
		return false
	}
	if s.Request.ContentType != other.Request.ContentType {
		return false
	}
	if !s.Request.Headers.Equal(other.Request.Headers) {
		return false
	}
	if s.Request.Body != other.Request.Body {
		return false
	}

	return true
}

func (s *Session) RequestDifferences(other *Session) *RequestDifference {
	diff := &RequestDifference{}

	diff.Method = compareStringField(s.Request.Method, other.Request.Method)

	diff.URL = compareStringField(s.Request.URL, other.Request.URL)

	diff.Body = compareStringField(s.Request.Body, other.Request.Body)

	diff.ContentType = compareStringField(s.Request.ContentType, other.Request.ContentType)

	diff.Headers = compareHeaders(s.Request.Headers, other.Request.Headers)

	diff.Cookies = compareCookies(s.Request.Cookies, other.Request.Cookies)

	diff.HasDiffs = diff.Method.Changed || diff.URL.Changed || diff.Body.Changed ||
		diff.ContentType.Changed || diff.Headers.Changed || diff.Cookies.Changed

	return diff
}

func (s *Session) ToCurl() string {
	if s.Type == WebSocketSession {
		return "WebSocket sessions cannot be converted to cURL commands."
	}
	curlCommand := "curl -X " + s.Request.Method + " '" + s.Request.URL + "'"

	for _, key := range s.Request.Headers.Order {
		if key == "Host" || key == "Content-Length" {
			continue
		}
		val, _ := s.Request.Headers.Get(key)
		valStr := fmt.Sprintf("%v", val)
		curlCommand += fmt.Sprintf(" -H '%s: %s'", key, valStr)
	}

	if s.Request.Body != "" {
		curlCommand += " -d '" + s.Request.Body + "'"
	}

	return curlCommand
}

func (s *Session) Replay(port int) error {
	if s.Type == WebSocketSession {
		return fmt.Errorf("WebSocket sessions cannot be replayed")
	}

	parsedURL, err := url.Parse(s.Request.URL)
	if err != nil {
		return fmt.Errorf("failed to parse URL %s: %w", s.Request.URL, err)
	}

	proxyURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("127.0.0.1:%d", port),
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	if parsedURL.Scheme == "https" {
		if s.TLSFingerprint != nil {
			transport.TLSClientConfig = s.TLSFingerprint.ToTLSConfig()
			transport.TLSClientConfig.InsecureSkipVerify = true
		} else {
			transport.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		if hasHTTP2(s.TLSFingerprint) {
			http2.ConfigureTransport(transport)
		}
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	var bodyReader io.Reader
	if s.Request.Body != "" {
		bodyReader = strings.NewReader(s.Request.Body)
	}

	req, err := http.NewRequest(s.Request.Method, s.Request.URL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for _, key := range s.Request.Headers.Order {
		val, _ := s.Request.Headers.Get(key)
		if val != nil {
			req.Header.Set(key, fmt.Sprintf("%v", val))
		}
	}

	if s.Request.Cookies != nil {
		for name, value := range s.Request.Cookies {
			req.AddCookie(&http.Cookie{
				Name:  name,
				Value: fmt.Sprintf("%v", value),
			})
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error replaying request: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

func hasHTTP2(fp *clientHello.TLSFingerprint) bool {
	if fp == nil {
		return false
	}
	for _, proto := range fp.ALPNProtocols {
		if proto == "h2" {
			return true
		}
	}
	return false
}
