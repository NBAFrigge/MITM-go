package sessiondata

import (
	"crypto/tls"
	"fmt"
	"httpDebugger/internal/sortedMap"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/tls-client"
	"github.com/google/uuid"
)

type Session struct {
	ID        string
	Timestamp time.Time
	TLSConfig *tls.Config
	Request   *RequestData
	Response  *ResponseData
	Duration  time.Duration
	Error     error
	Protocol  string
	Type      SessionType
	WebSocket *WebSocketData
}

func NewSessionData(r *http.Request, bodyBytes []byte, headers *sortedMap.SortedMap, tlsConfig *tls.Config, protocol string) *Session {

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
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Request:   requestData,
			Type:      WebSocketSession,
			TLSConfig: tlsConfig,
			Protocol:  protocol,
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
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Request:   requestData,
			Type:      HTTPSession,
			TLSConfig: tlsConfig,
			Protocol:  protocol,
			WebSocket: nil,
			Response:  nil,
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

	isHTTPS := parsedURL.Scheme == "https"

	opts := []tls_client.HttpClientOption{
		tls_client.WithProxyUrl(fmt.Sprintf("http://localhost:%d", port)),
		tls_client.WithInsecureSkipVerify(),
	}
	if isHTTPS {
		client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), opts...)
		if err != nil {
			return fmt.Errorf("failed to create HTTP client: %w", err)
		}

		headers := sortedMapToHeaders(s.Request.Headers)
		var bodyReader io.Reader
		if s.Request.Body != "" {
			bodyReader = strings.NewReader(s.Request.Body)
		}

		req, err := fhttp.NewRequest(s.Request.Method, s.Request.URL, bodyReader)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header = headers

		if s.Request.Cookies != nil {
			for name, value := range s.Request.Cookies {
				cookie := &fhttp.Cookie{
					Name:  name,
					Value: fmt.Sprintf("%v", value),
				}
				req.AddCookie(cookie)
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("error replaying request: %w", err)
		}
		defer resp.Body.Close()
	} else {
		client := &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyURL(&url.URL{
					Scheme: "http",
					Host:   fmt.Sprintf("localhost:%d", port),
				}),
			},
		}
		headers := s.Request.Headers.Entries
		req, err := http.NewRequest(s.Request.Method, s.Request.URL, strings.NewReader(s.Request.Body))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		for key, values := range headers {
			req.Header.Add(key, fmt.Sprintf("%v", values))
		}
		if s.Request.Cookies != nil {
			for name, value := range s.Request.Cookies {
				cookie := &http.Cookie{
					Name:  name,
					Value: fmt.Sprintf("%v", value),
				}
				req.AddCookie(cookie)
			}
		}

		_, err = client.Do(req)
		if err != nil {
			return fmt.Errorf("error replaying request: %w", err)
		}
	}

	return nil
}
