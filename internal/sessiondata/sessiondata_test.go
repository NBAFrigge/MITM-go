package sessiondata

import (
	"bytes"
	"fmt"
	"httpDebugger/internal/sortedMap"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestNewSessionData(t *testing.T) {
	req := createTestRequest("GET", "https://example.com/api", "test body", map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer token123",
	})

	req.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})
	req.AddCookie(&http.Cookie{Name: "user", Value: "testuser"})

	bodyBytes := []byte("test body")
	headers := createTestSortedMap(map[string]interface{}{
		"Content-Type":  "application/json",
		"Authorization": "Bearer token123",
	})

	session := NewSessionData(req, bodyBytes, headers)

	if session.ID == "" {
		t.Error("Expected session ID to be generated")
	}

	if session.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}

	if session.Request == nil {
		t.Fatal("Expected request data to be set")
	}

	if session.Request.Method != "GET" {
		t.Errorf("Expected method GET, got %s", session.Request.Method)
	}

	if session.Request.URL != "https://example.com/api" {
		t.Errorf("Expected URL https://example.com/api, got %s", session.Request.URL)
	}

	if session.Request.Body != "test body" {
		t.Errorf("Expected body 'test body', got %s", session.Request.Body)
	}

	if session.Request.ContentType != "application/json" {
		t.Errorf("Expected content type application/json, got %s", session.Request.ContentType)
	}

	if len(session.Request.Cookies) != 2 {
		t.Errorf("Expected 2 cookies, got %d", len(session.Request.Cookies))
	}

	if session.Request.Cookies["session"] != "abc123" {
		t.Errorf("Expected session cookie abc123, got %v", session.Request.Cookies["session"])
	}

	if session.Request.Cookies["user"] != "testuser" {
		t.Errorf("Expected user cookie testuser, got %v", session.Request.Cookies["user"])
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name: "X-Forwarded-For single IP",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.100",
			},
			remoteAddr: "10.0.0.1:8080",
			expected:   "192.168.1.100",
		},
		{
			name: "X-Forwarded-For multiple IPs",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.100, 10.0.0.1, 172.16.0.1",
			},
			remoteAddr: "10.0.0.1:8080",
			expected:   "192.168.1.100",
		},
		{
			name: "X-Real-IP",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.200",
			},
			remoteAddr: "10.0.0.1:8080",
			expected:   "192.168.1.200",
		},
		{
			name:       "RemoteAddr with port",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.300:8080",
			expected:   "192.168.1.300",
		},
		{
			name:       "RemoteAddr without port",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.400",
			expected:   "192.168.1.400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest("GET", "https://example.com", "", tt.headers)
			req.RemoteAddr = tt.remoteAddr
		})
	}
}

func TestCompareRequest(t *testing.T) {
	session1 := createTestSession("GET", "https://example.com/api", "test body",
		map[string]interface{}{"Content-Type": "application/json"},
		map[string]interface{}{"session": "abc123"})

	session2 := createTestSession("GET", "https://example.com/api", "test body",
		map[string]interface{}{"Content-Type": "application/json"},
		map[string]interface{}{"session": "abc123"})

	if !session1.CompareRequest(session2) {
		t.Error("Expected identical sessions to be equal")
	}

	session3 := createTestSession("POST", "https://example.com/api", "test body",
		map[string]interface{}{"Content-Type": "application/json"},
		map[string]interface{}{"session": "abc123"})

	if session1.CompareRequest(session3) {
		t.Error("Expected sessions with different methods to be different")
	}

	session4 := createTestSession("GET", "https://example.com/different", "test body",
		map[string]interface{}{"Content-Type": "application/json"},
		map[string]interface{}{"session": "abc123"})

	if session1.CompareRequest(session4) {
		t.Error("Expected sessions with different URLs to be different")
	}

	session5 := createTestSession("GET", "https://example.com/api", "different body",
		map[string]interface{}{"Content-Type": "application/json"},
		map[string]interface{}{"session": "abc123"})

	if session1.CompareRequest(session5) {
		t.Error("Expected sessions with different bodies to be different")
	}

	session6 := createTestSession("GET", "https://example.com/api", "test body",
		map[string]interface{}{"Content-Type": "text/plain"},
		map[string]interface{}{"session": "abc123"})

	if session1.CompareRequest(session6) {
		t.Error("Expected sessions with different content types to be different")
	}
}

func TestRequestDifferences(t *testing.T) {
	req1 := &RequestData{
		Method:      "GET",
		URL:         "https://example.com/api",
		Body:        "original body",
		ContentType: "application/json",
		Headers:     createTestSortedMap(map[string]interface{}{"Authorization": "Bearer token1"}),
		Cookies:     map[string]interface{}{"session": "abc123", "user": "john"},
	}

	req2 := &RequestData{
		Method:      "POST",
		URL:         "https://example.com/api",
		Body:        "modified body",
		ContentType: "application/xml",
		Headers:     createTestSortedMap(map[string]interface{}{"Authorization": "Bearer token2"}),
		Cookies:     map[string]interface{}{"session": "xyz789", "role": "admin"},
	}

	diff := req1.RequestDifferences(req2)

	if !diff.HasDiffs {
		t.Error("Expected differences to be found")
	}

	if !diff.Method.Changed || diff.Method.Original != "GET" || diff.Method.Other != "POST" {
		t.Error("Method differences not detected correctly")
	}

	if !diff.Body.Changed || diff.Body.Original != "original body" || diff.Body.Other != "modified body" {
		t.Error("Body differences not detected correctly")
	}

	if !diff.ContentType.Changed || diff.ContentType.Original != "application/json" || diff.ContentType.Other != "application/xml" {
		t.Error("ContentType differences not detected correctly")
	}

	if diff.URL.Changed {
		t.Error("URL should not be marked as changed")
	}

	if !diff.Cookies.Changed {
		t.Error("Cookie differences not detected")
	}
}

func TestRequestDifferencesIdentical(t *testing.T) {
	req1 := &RequestData{
		Method:      "GET",
		URL:         "https://example.com/api",
		Body:        "test body",
		ContentType: "application/json",
		Headers:     createTestSortedMap(map[string]interface{}{"Content-Type": "application/json"}),
		Cookies:     map[string]interface{}{"session": "abc123"},
	}

	req2 := &RequestData{
		Method:      "GET",
		URL:         "https://example.com/api",
		Body:        "test body",
		ContentType: "application/json",
		Headers:     createTestSortedMap(map[string]interface{}{"Content-Type": "application/json"}),
		Cookies:     map[string]interface{}{"session": "abc123"},
	}

	diff := req1.RequestDifferences(req2)

	if diff.HasDiffs {
		t.Error("Expected no differences for identical requests")
	}

	if diff.Method.Changed || diff.URL.Changed || diff.Body.Changed ||
		diff.ContentType.Changed || diff.Headers.Changed || diff.Cookies.Changed {
		t.Error("No fields should be marked as changed for identical requests")
	}
}

func TestToCurl(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		url      string
		body     string
		headers  map[string]interface{}
		cookies  map[string]interface{}
		expected string
	}{
		{
			name:   "Simple GET request",
			method: "GET",
			url:    "https://api.example.com/users",
			body:   "",
			headers: map[string]interface{}{
				"Accept":     "application/json",
				"User-Agent": "TestAgent/1.0",
			},
			cookies:  map[string]interface{}{},
			expected: "curl -X GET 'https://api.example.com/users' -H 'Accept: application/json' -H 'User-Agent: TestAgent/1.0'",
		},
		{
			name:   "POST request with body",
			method: "POST",
			url:    "https://api.example.com/users",
			body:   `{"name":"John","email":"john@example.com"}`,
			headers: map[string]interface{}{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token123",
			},
			cookies:  map[string]interface{}{},
			expected: "curl -X POST 'https://api.example.com/users' -H 'Content-Type: application/json' -H 'Authorization: Bearer token123' -d '{\"name\":\"John\",\"email\":\"john@example.com\"}'",
		},
		{
			name:   "Request with Host header (should be skipped)",
			method: "GET",
			url:    "https://api.example.com/data",
			body:   "",
			headers: map[string]interface{}{
				"Host":           "api.example.com",
				"Accept":         "application/json",
				"Content-Length": "123",
			},
			cookies:  map[string]interface{}{},
			expected: "curl -X GET 'https://api.example.com/data' -H 'Accept: application/json'",
		},
		{
			name:   "Request with multiple headers",
			method: "PUT",
			url:    "https://api.example.com/users/123",
			body:   `{"name":"Updated Name"}`,
			headers: map[string]interface{}{
				"Content-Type":    "application/json",
				"Authorization":   "Bearer token456",
				"X-Custom-Header": "custom-value",
				"Accept-Encoding": "gzip, deflate",
			},
			cookies:  map[string]interface{}{},
			expected: "curl -X PUT 'https://api.example.com/users/123' -H 'Content-Type: application/json' -H 'Authorization: Bearer token456' -H 'X-Custom-Header: custom-value' -H 'Accept-Encoding: gzip, deflate' -d '{\"name\":\"Updated Name\"}'",
		},
		{
			name:   "DELETE request without body",
			method: "DELETE",
			url:    "https://api.example.com/users/123",
			body:   "",
			headers: map[string]interface{}{
				"Authorization": "Bearer token789",
			},
			cookies:  map[string]interface{}{},
			expected: "curl -X DELETE 'https://api.example.com/users/123' -H 'Authorization: Bearer token789'",
		},
		{
			name:   "Request with special characters in URL",
			method: "GET",
			url:    "https://api.example.com/search?q=test&limit=10",
			body:   "",
			headers: map[string]interface{}{
				"Accept": "application/json",
			},
			cookies:  map[string]interface{}{},
			expected: "curl -X GET 'https://api.example.com/search?q=test&limit=10' -H 'Accept: application/json'",
		},
		{
			name:     "Request with empty headers",
			method:   "GET",
			url:      "https://api.example.com/simple",
			body:     "",
			headers:  map[string]interface{}{},
			cookies:  map[string]interface{}{},
			expected: "curl -X GET 'https://api.example.com/simple'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := createTestSession(tt.method, tt.url, tt.body, tt.headers, tt.cookies)
			result := session.ToCurl()

			if result != tt.expected {
				t.Errorf("ToCurl() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestToCurlWithInterfaceHeaders(t *testing.T) {
	session := createTestSession("POST", "https://api.example.com/data", "test data",
		map[string]interface{}{
			"Content-Type":   "text/plain",
			"Content-Length": 123,
			"Rate-Limit":     []string{"100", "per-hour"},
			"Custom-Bool":    true,
		},
		map[string]interface{}{})

	result := session.ToCurl()

	expectedStart := "curl -X POST 'https://api.example.com/data'"
	if !strings.HasPrefix(result, expectedStart) {
		t.Errorf("Expected curl command to start with %q, got %q", expectedStart, result)
	}

	expectedEnd := " -d 'test data'"
	if !strings.HasSuffix(result, expectedEnd) {
		t.Errorf("Expected curl command to end with %q, got %q", expectedEnd, result)
	}

	expectedHeaders := []string{
		"-H 'Content-Type: text/plain'",
		"-H 'Rate-Limit: [100 per-hour]'",
		"-H 'Custom-Bool: true'",
	}

	for _, header := range expectedHeaders {
		if !strings.Contains(result, header) {
			t.Errorf("Expected curl command to contain %q, got %q", header, result)
		}
	}

	if strings.Contains(result, "Content-Length") {
		t.Error("Content-Length header should be excluded from curl command")
	}
}

func TestToCurlEscaping(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		headers  map[string]interface{}
		expected string
	}{
		{
			name: "Body with single quotes",
			body: `{"message":"It's a test"}`,
			headers: map[string]interface{}{
				"Content-Type": "application/json",
			},
			expected: "curl -X POST 'https://api.example.com/test' -H 'Content-Type: application/json' -d '{\"message\":\"It's a test\"}'",
		},
		{
			name: "Header with single quotes",
			body: "",
			headers: map[string]interface{}{
				"Custom-Header": "Value with 'quotes'",
			},
			expected: "curl -X GET 'https://api.example.com/test' -H 'Custom-Header: Value with 'quotes''",
		},
		{
			name: "Body with newlines",
			body: "{\n  \"name\": \"test\"\n}",
			headers: map[string]interface{}{
				"Content-Type": "application/json",
			},
			expected: "curl -X POST 'https://api.example.com/test' -H 'Content-Type: application/json' -d '{\n  \"name\": \"test\"\n}'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := "POST"
			if tt.body == "" {
				method = "GET"
			}

			session := createTestSession(method, "https://api.example.com/test", tt.body, tt.headers, map[string]interface{}{})
			result := session.ToCurl()

			if result != tt.expected {
				t.Errorf("ToCurl() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestToCurlExcludedHeaders(t *testing.T) {
	session := createTestSession("POST", "https://api.example.com/test", "test body",
		map[string]interface{}{
			"Host":           "api.example.com",
			"Content-Length": "9",
			"Content-Type":   "text/plain",
			"Authorization":  "Bearer token123",
		},
		map[string]interface{}{})

	result := session.ToCurl()

	if strings.Contains(result, "Host:") {
		t.Error("Host header should be excluded from curl command")
	}

	if strings.Contains(result, "Content-Length:") {
		t.Error("Content-Length header should be excluded from curl command")
	}

	if !strings.Contains(result, "Content-Type: text/plain") {
		t.Error("Content-Type header should be included in curl command")
	}

	if !strings.Contains(result, "Authorization: Bearer token123") {
		t.Error("Authorization header should be included in curl command")
	}
}

func TestToCurlHeaderOrder(t *testing.T) {
	session := &Session{
		ID:        "test-id",
		Timestamp: time.Now(),
		Request: &RequestData{
			Method:      "POST",
			URL:         "https://api.example.com/ordered",
			Body:        "test body",
			ContentType: "application/json",
			Headers:     createOrderedTestSortedMap(),
			Cookies:     map[string]interface{}{},
		},
	}

	result := session.ToCurl()

	expectedHeaders := []string{"Authorization", "Content-Type", "Accept"}

	lastIndex := -1
	for _, header := range expectedHeaders {
		headerPattern := fmt.Sprintf("-H '%s:", header)
		headerIndex := strings.Index(result, headerPattern)

		if headerIndex == -1 {
			t.Errorf("Header %s not found in curl command", header)
			continue
		}

		if headerIndex <= lastIndex {
			t.Errorf("Headers are not in expected order. %s should appear after previous header", header)
		}

		lastIndex = headerIndex
	}
}

func createOrderedTestSortedMap() sortedMap.SortedMap {
	sm := sortedMap.New()
	sm.Put("Authorization", "Bearer token123")
	sm.Put("Content-Type", "application/json")
	sm.Put("Accept", "application/json")
	return sm
}

func createTestRequest(method, urlStr, body string, headers map[string]string) *http.Request {
	req := &http.Request{
		Method: method,
		Header: make(http.Header),
	}

	if urlStr != "" {
		parsedURL, _ := url.Parse(urlStr)
		req.URL = parsedURL
	}

	if body != "" {
		req.Body = &testReadCloser{bytes.NewBufferString(body)}
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req
}

func createTestSession(method, url, body string, headers, cookies map[string]interface{}) *Session {
	return &Session{
		ID:        "test-id",
		Timestamp: time.Now(),
		Request: &RequestData{
			Method:      method,
			URL:         url,
			Body:        body,
			ContentType: getContentType(headers),
			Headers:     createTestSortedMap(headers),
			Cookies:     cookies,
		},
	}
}

func createTestSortedMap(data map[string]interface{}) sortedMap.SortedMap {
	sm := sortedMap.New()
	for key, value := range data {
		sm.Put(key, value)
	}
	return sm
}

func getContentType(headers map[string]interface{}) string {
	if ct, exists := headers["Content-Type"]; exists {
		if ctStr, ok := ct.(string); ok {
			return ctStr
		}
	}
	return ""
}

type testReadCloser struct {
	*bytes.Buffer
}

func (trc *testReadCloser) Close() error {
	return nil
}

func BenchmarkNewSessionData(b *testing.B) {
	req := createTestRequest("GET", "https://example.com/api", "test body", map[string]string{
		"Content-Type": "application/json",
	})
	bodyBytes := []byte("test body")
	headers := createTestSortedMap(map[string]interface{}{"Content-Type": "application/json"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewSessionData(req, bodyBytes, headers)
	}
}

func BenchmarkCompareRequest(b *testing.B) {
	session1 := createTestSession("GET", "https://example.com/api", "test body",
		map[string]interface{}{"Content-Type": "application/json"},
		map[string]interface{}{"session": "abc123"})

	session2 := createTestSession("POST", "https://example.com/api", "different body",
		map[string]interface{}{"Content-Type": "application/xml"},
		map[string]interface{}{"session": "xyz789"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session1.CompareRequest(session2)
	}
}

func BenchmarkRequestDifferences(b *testing.B) {
	req1 := &RequestData{
		Method:      "GET",
		URL:         "https://example.com/api",
		Body:        "original body",
		ContentType: "application/json",
		Headers:     createTestSortedMap(map[string]interface{}{"Authorization": "Bearer token1"}),
		Cookies:     map[string]interface{}{"session": "abc123", "user": "john"},
	}

	req2 := &RequestData{
		Method:      "POST",
		URL:         "https://example.com/api/v2",
		Body:        "modified body",
		ContentType: "application/xml",
		Headers:     createTestSortedMap(map[string]interface{}{"Authorization": "Bearer token2"}),
		Cookies:     map[string]interface{}{"session": "xyz789", "role": "admin"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req1.RequestDifferences(req2)
	}
}

func BenchmarkToCurl(b *testing.B) {
	session := createTestSession("POST", "https://api.example.com/benchmark",
		`{"data":"benchmark test"}`,
		map[string]interface{}{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
			"Accept":        "application/json",
			"User-Agent":    "BenchmarkAgent/1.0",
		},
		map[string]interface{}{"session": "abc123"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session.ToCurl()
	}
}
