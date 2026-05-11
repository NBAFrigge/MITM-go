package utils

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestReadAndCloseBody_ReadsAll(t *testing.T) {
	body := "hello world"
	resp := &http.Response{
		Body: io.NopCloser(strings.NewReader(body)),
	}

	data, err := ReadAndCloseBody(resp)
	if err != nil {
		t.Fatalf("ReadAndCloseBody failed: %v", err)
	}
	if string(data) != body {
		t.Errorf("ReadAndCloseBody = %q, want %q", string(data), body)
	}
}

func TestReadAndCloseBody_Empty(t *testing.T) {
	resp := &http.Response{
		Body: io.NopCloser(bytes.NewReader(nil)),
	}

	data, err := ReadAndCloseBody(resp)
	if err != nil {
		t.Fatalf("ReadAndCloseBody failed: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("expected empty body, got %d bytes", len(data))
	}
}

func TestReadAndCloseBody_EnforcesLimit(t *testing.T) {
	large := strings.Repeat("x", maxBodySize+1000)
	resp := &http.Response{
		Body: io.NopCloser(strings.NewReader(large)),
	}

	data, err := ReadAndCloseBody(resp)
	if err != nil {
		t.Fatalf("ReadAndCloseBody failed: %v", err)
	}
	if len(data) != maxBodySize {
		t.Errorf("expected body capped at %d bytes, got %d", maxBodySize, len(data))
	}
}

func TestWriteHTTPResponse_StatusLine(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader("")),
	}

	var buf bytes.Buffer
	if err := WriteHTTPResponse(&buf, resp); err != nil {
		t.Fatalf("WriteHTTPResponse failed: %v", err)
	}

	if !strings.HasPrefix(buf.String(), "HTTP/1.1 200 OK") {
		t.Errorf("expected HTTP/1.1 200 OK prefix, got %q", buf.String()[:20])
	}
}

func TestWriteHTTPResponse_WritesBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": {"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
	}

	var buf bytes.Buffer
	WriteHTTPResponse(&buf, resp)

	if !strings.Contains(buf.String(), `{"ok":true}`) {
		t.Errorf("body not in output: %q", buf.String())
	}
}

func TestWriteHTTPResponse_FiltersHopByHop(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Header: http.Header{
			"Content-Type": {"text/plain"},
			"Connection":   {"keep-alive"},
		},
		Body: io.NopCloser(strings.NewReader("body")),
	}

	var buf bytes.Buffer
	WriteHTTPResponse(&buf, resp)
	output := buf.String()

	if strings.Contains(output, "Connection:") {
		t.Error("Connection header should be filtered")
	}
	if !strings.Contains(output, "Content-Type: text/plain") {
		t.Error("Content-Type should be present")
	}
}

func TestWriteHTTPResponse_404(t *testing.T) {
	resp := &http.Response{
		StatusCode: 404,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader("not found")),
	}

	var buf bytes.Buffer
	WriteHTTPResponse(&buf, resp)

	if !strings.HasPrefix(buf.String(), "HTTP/1.1 404 Not Found") {
		t.Errorf("expected HTTP/1.1 404 Not Found prefix, got: %q", buf.String()[:30])
	}
}

func TestSendErrorResponse_Format(t *testing.T) {
	var buf bytes.Buffer
	SendErrorResponse(&buf, 502, "Bad Gateway")

	output := buf.String()
	if !strings.HasPrefix(output, "HTTP/1.1 502 Bad Gateway") {
		t.Errorf("expected status line prefix, got %q", output)
	}
	if !strings.Contains(output, "Content-Length: 0") {
		t.Error("expected Content-Length: 0")
	}
	if !strings.Contains(output, "Connection: close") {
		t.Error("expected Connection: close")
	}
}

func TestSendErrorResponse_EndsWithCRLF(t *testing.T) {
	var buf bytes.Buffer
	SendErrorResponse(&buf, 400, "Bad Request")

	if !strings.HasSuffix(buf.String(), "\r\n\r\n") {
		t.Error("response should end with CRLFCRLF")
	}
}
